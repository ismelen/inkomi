import { Platform } from 'react-native';
import { File } from 'expo-file-system';
import { createUploadTask, FileSystemUploadType, UploadTask } from 'expo-file-system/legacy';
import { copyToCache } from 'file-handler';
import { uploadsApi } from '../api/uploads-api';
import { ConfigDTO } from '../api/types/manga-config';
import { NoItemsSourceDTO, SourceDTO, UploadRequestDTO } from '../api/types/uploads';
import { findEReader } from '../models/ereaders';
import { Group } from '../models/group';
import { isManga, ListItem } from '../models/list-item';
import { MangaConfig, toConfigDTO } from '../models/manga-config';
import { LocalSource } from '../models/upload-tree';
import { advanceStatus, deleteTrees, upsertSources } from './activity-db';

const concurrency = 3;

export interface ListSnapshot {
  items: Record<string, ListItem>;
  groups: Record<string, Group>;
  rootOrder: string[];
}

export interface SendOptions {
  ereaderKey: string;
  toCloud: boolean;
  defaultManga: MangaConfig;
}

/** One node of the request that has to be confirmed: a file to PUT or a library item. */
export interface UploadJob {
  sourceId: string;
  rootId: string;
  listItemId: string;
  title: string;
  /** Presigned PUT url; absent for library items. */
  url?: string;
  item: ListItem;
}

export interface UploadHandlers {
  onJobs(jobs: UploadJob[]): void;
  onProgress(sourceId: string, fraction: number): void;
  onConfirmed(job: UploadJob): void;
  onFailed(job: UploadJob, error: unknown): void;
}

interface Plan {
  request: UploadRequestDTO;
  /** List item ids per request source: [root item] for loose items, the group items otherwise. */
  listIds: { rootListId: string; itemIds: string[] }[];
}

class ConfigKeys {
  private keys = new Map<string, string>();
  readonly configs: Record<string, ConfigDTO> = {};

  constructor(private ereaderKey: string) {}

  keyOf(config: MangaConfig): string {
    const dto = toConfigDTO(config, this.ereaderKey);
    const json = JSON.stringify(dto);
    let key = this.keys.get(json);
    if (!key) {
      key = `c${this.keys.size}`;
      this.keys.set(json, key);
      this.configs[key] = dto;
    }
    return key;
  }
}

function itemSource(
  item: ListItem,
  configKey: string | undefined,
  opts: SendOptions,
  rtl: boolean
): NoItemsSourceDTO {
  const direction = rtl ? 'rtl' : 'ltr';
  if (item.kind === 'library') {
    return {
      title: item.title,
      type: 'library',
      filename: item.md5,
      config_hash: configKey,
      reading_direction: direction,
      to_cloud: opts.toCloud,
    };
  }
  return {
    title: item.title,
    type: 'file',
    size: item.size,
    filename: item.name,
    config_hash: configKey,
    reading_direction: direction,
    to_cloud: opts.toCloud,
  };
}

export function buildPlan(list: ListSnapshot, opts: SendOptions): Plan {
  const configs = new ConfigKeys(opts.ereaderKey);
  const sources: SourceDTO[] = [];
  const listIds: Plan['listIds'] = [];

  for (const id of list.rootOrder) {
    const group = list.groups[id];
    if (group) {
      const items = group.itemIds.map((itemId) => list.items[itemId]).filter(Boolean);
      if (items.length === 0) continue;
      const groupConfig = group.config ?? opts.defaultManga;
      const groupHasManga = items.some((i) => isManga(i.ext));
      sources.push({
        title: group.title,
        type: 'folder',
        filename: group.title,
        config_hash: groupHasManga ? configs.keyOf(groupConfig) : undefined,
        reading_direction: group.rtl ? 'rtl' : 'ltr',
        to_cloud: opts.toCloud,
        should_join: group.join,
        items: items.map((item) =>
          itemSource(
            item,
            isManga(item.ext) && item.config ? configs.keyOf(item.config) : undefined,
            opts,
            group.rtl
          )
        ),
      });
      listIds.push({ rootListId: group.id, itemIds: items.map((i) => i.id) });
      continue;
    }

    const item = list.items[id];
    if (!item) continue;
    const configKey = isManga(item.ext)
      ? configs.keyOf(item.config ?? opts.defaultManga)
      : undefined;
    sources.push({ ...itemSource(item, configKey, opts, item.rtl), should_join: false, items: [] });
    listIds.push({ rootListId: item.id, itemIds: [] });
  }

  return {
    request: { ereader_key: opts.ereaderKey, sources, config: configs.configs },
    listIds,
  };
}

function localRow(
  dto: NoItemsSourceDTO,
  item: ListItem | undefined,
  extra: Pick<LocalSource, 'parentId' | 'position' | 'join' | 'isRtl' | 'kepubify'>,
  now: string
): LocalSource {
  return {
    id: dto.id!,
    title: dto.title,
    filename: dto.filename,
    toCloud: dto.to_cloud,
    status: 'pending_upload',
    errorMsg: null,
    size: dto.size ?? null,
    isFile: dto.type !== 'folder',
    createdAt: now,
    updatedAt: now,
    completedAt: null,
    isResult: false,
    kind: dto.type,
    ext: item?.ext ?? null,
    ...extra,
  };
}

/** Rows of the created sources, kept locally as pending until each one is confirmed. */
function localRows(plan: Plan, response: SourceDTO[], list: ListSnapshot, kepubify: boolean) {
  const now = new Date().toISOString();
  return response.flatMap((root, i) => {
    const ids = plan.listIds[i];
    const rootItem = list.items[ids.rootListId];
    const rtl = root.reading_direction === 'rtl';
    const rows = [
      localRow(
        root,
        rootItem,
        { parentId: null, position: 0, join: root.should_join, isRtl: rtl, kepubify },
        now
      ),
    ];
    root.items.forEach((child, j) =>
      rows.push(
        localRow(
          child,
          list.items[ids.itemIds[j]],
          { parentId: root.id!, position: j, join: root.should_join, isRtl: rtl, kepubify },
          now
        )
      )
    );
    return rows;
  });
}

function jobsOf(plan: Plan, response: SourceDTO[], list: ListSnapshot): UploadJob[] {
  return response.flatMap((root, i) => {
    const ids = plan.listIds[i];
    if (root.type !== 'folder') {
      return [
        {
          sourceId: root.id!,
          rootId: root.id!,
          listItemId: ids.rootListId,
          title: root.title,
          url: root.url,
          item: list.items[ids.rootListId],
        },
      ];
    }
    return root.items.map((child, j) => ({
      sourceId: child.id!,
      rootId: root.id!,
      listItemId: ids.itemIds[j],
      title: child.title,
      url: child.url,
      item: list.items[ids.itemIds[j]],
    }));
  });
}

/** Android pickers return content:// uris; the uploader needs a file in the cache. */
async function uploadableUri(item: ListItem & { kind: 'file' }) {
  if (Platform.OS === 'android' && item.uri.startsWith('content://')) {
    return { uri: await copyToCache(item.uri, item.name), temporary: true };
  }
  return { uri: item.uri, temporary: false };
}

export class UploadRun {
  private cancelled = false;
  private tasks = new Set<UploadTask>();
  private confirmedRoots = new Set<string>();
  private rootIds: string[] = [];

  constructor(private handlers: UploadHandlers) {}

  /**
   * request → PUT of every file → PATCH done per file and library item (never on a folder).
   * Returns the list item ids that were confirmed.
   */
  async run(list: ListSnapshot, opts: SendOptions): Promise<string[]> {
    const plan = buildPlan(list, opts);
    const response = await uploadsApi.request(plan.request);
    const kepubify = findEReader(opts.ereaderKey).kepub;
    await upsertSources(localRows(plan, response, list, kepubify));
    this.rootIds = response.map((r) => r.id!);

    const jobs = jobsOf(plan, response, list);
    this.handlers.onJobs(jobs);
    const confirmed: string[] = [];

    const queue = [...jobs];
    const worker = async () => {
      for (let job = queue.shift(); job && !this.cancelled; job = queue.shift()) {
        try {
          await this.process(job);
          confirmed.push(job.listItemId);
          this.confirmedRoots.add(job.rootId);
          this.handlers.onConfirmed(job);
        } catch (err) {
          if (!this.cancelled) this.handlers.onFailed(job, err);
        }
      }
    };
    await Promise.all(Array.from({ length: concurrency }, worker));

    await this.dropUnconfirmed();
    return confirmed;
  }

  /** Aborts the uploads in flight; what was already confirmed keeps being processed. */
  cancel() {
    this.cancelled = true;
    this.tasks.forEach((task) => void task.cancelAsync());
  }

  private async process(job: UploadJob) {
    if (job.url && job.item.kind === 'file') await this.put(job, job.item);
    if (this.cancelled) throw new Error('cancelled');
    await uploadsApi.done(job.sourceId);
    await advanceStatus([job.sourceId], 'queued');
  }

  private async put(job: UploadJob, item: ListItem & { kind: 'file' }) {
    const { uri, temporary } = await uploadableUri(item);
    try {
      const task = createUploadTask(
        job.url!,
        uri,
        {
          httpMethod: 'PUT',
          uploadType: FileSystemUploadType.BINARY_CONTENT,
          headers: { 'Content-Type': item.mime ?? 'application/octet-stream' },
        },
        ({ totalBytesSent, totalBytesExpectedToSend }) => {
          if (totalBytesExpectedToSend > 0) {
            this.handlers.onProgress(job.sourceId, totalBytesSent / totalBytesExpectedToSend);
          }
        }
      );
      this.tasks.add(task);
      const result = await task.uploadAsync();
      this.tasks.delete(task);
      if (!result || result.status >= 300) {
        throw new Error(`upload failed: ${result?.status ?? 'cancelled'}`);
      }
    } finally {
      if (temporary) {
        try {
          new File(uri).delete();
        } catch {
          // the cache is cleaned by the system anyway
        }
      }
    }
  }

  /** Uploads with nothing confirmed will never start on the server: forget them locally. */
  private async dropUnconfirmed() {
    await deleteTrees(this.rootIds.filter((id) => !this.confirmedRoots.has(id)));
  }
}
