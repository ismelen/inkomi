import { DBSchema, IDBPDatabase, IDBPTransaction, openDB } from 'idb';
import { SourceStatusEvent } from '../api/types/events';
import { SourceRow, SourceStatus } from '../api/types/uploads';
import { formatOf } from '../models/list-item';
import { isTerminal, isTreeFinished, LocalSource, UploadTree } from '../models/upload-tree';

const dbName = 'inkomi-activity';

// Later steps never go back: a stale list response can't undo a newer websocket event.
const statusRank: Record<SourceStatus, number> = {
  pending_upload: 0,
  queued: 1,
  processing: 2,
  joining: 3,
  joined: 4,
  kepubifying: 5,
  done: 6,
  sent: 7,
  failed: 8,
};

/** IndexedDB can't index null: roots use an empty parent key. */
interface SourceRecord extends LocalSource {
  parentKey: string;
  finished: 0 | 1;
}

export interface SyncState {
  syncedAt: string | null;
  historyCursor: string | null;
  historyExhausted: boolean;
}

export interface HistoryCursor {
  createdAt: string;
  id: string;
}

interface ActivitySchema extends DBSchema {
  sources: {
    key: string;
    value: SourceRecord;
    indexes: { byParent: string; byState: [string, number] };
  };
  syncState: { key: string; value: SyncState };
}

type Db = IDBPDatabase<ActivitySchema>;
type Tx = IDBPTransaction<ActivitySchema, ['sources'], 'readwrite'>;

const syncKey = 'state';
const emptySync: SyncState = { syncedAt: null, historyCursor: null, historyExhausted: false };

type Listener = () => void;
const listeners = new Set<Listener>();
let dbPromise: Promise<Db> | null = null;

// The tab that owns the websocket writes; every tab of the browser shows the same activity.
const channel =
  typeof BroadcastChannel === 'undefined' ? null : new BroadcastChannel('inkomi-activity');
channel?.addEventListener('message', () => listeners.forEach((listener) => listener()));

function db(): Promise<Db> {
  dbPromise ??= openDB<ActivitySchema>(dbName, 1, {
    upgrade(conn) {
      const sources = conn.createObjectStore('sources', { keyPath: 'id' });
      sources.createIndex('byParent', 'parentKey');
      sources.createIndex('byState', ['parentKey', 'finished']);
      conn.createObjectStore('syncState');
    },
  });
  return dbPromise;
}

function changed() {
  listeners.forEach((listener) => listener());
  channel?.postMessage('changed');
}

export function onActivityChange(listener: Listener): () => void {
  listeners.add(listener);
  return () => listeners.delete(listener);
}

/** ISO with milliseconds, so local and server times compare as text. */
export function normalizeTime(value: string | null | undefined): string | null {
  if (!value) return null;
  const time = Date.parse(value);
  return Number.isNaN(time) ? value : new Date(time).toISOString();
}

function fromRecord(r: SourceRecord): LocalSource {
  const { parentKey: _parentKey, finished: _finished, ...source } = r;
  return source;
}

function toRecord(s: LocalSource, finished: 0 | 1): SourceRecord {
  return { ...s, parentKey: s.parentId ?? '', finished };
}

/** Local view of a server row: the kind and format are guessed when the app didn't create it. */
export function toLocalSource(row: SourceRow, known?: Partial<Pick<LocalSource, 'kind' | 'ext'>>) {
  const kind: LocalSource['kind'] =
    known?.kind ?? (!row.isFile ? 'folder' : row.size ? 'file' : 'library');
  return {
    ...row,
    createdAt: normalizeTime(row.createdAt)!,
    updatedAt: normalizeTime(row.updatedAt)!,
    completedAt: normalizeTime(row.completedAt),
    kind,
    ext: known?.ext ?? (kind === 'file' ? formatOf(row.filename) : null),
  } satisfies LocalSource;
}

const byChildOrder = (a: SourceRecord, b: SourceRecord) =>
  Number(a.isResult) - Number(b.isResult) || a.position - b.position;

const newestFirst = (a: SourceRecord, b: SourceRecord) =>
  a.createdAt === b.createdAt ? (a.id < b.id ? 1 : -1) : a.createdAt < b.createdAt ? 1 : -1;

async function loadTreeTx(tx: Tx, rootId: string): Promise<UploadTree | null> {
  const store = tx.objectStore('sources');
  const root = await store.get(rootId);
  if (!root || root.parentId) return null;
  const children = await store.index('byParent').getAll(rootId);
  return { root: fromRecord(root), children: children.sort(byChildOrder).map(fromRecord) };
}

async function refreshFinished(tx: Tx, rootIds: Iterable<string>) {
  const store = tx.objectStore('sources');
  for (const rootId of new Set(rootIds)) {
    const tree = await loadTreeTx(tx, rootId);
    if (!tree) continue;
    const root = await store.get(rootId);
    if (root) await store.put({ ...root, finished: isTreeFinished(tree) ? 1 : 0 });
  }
}

const maxTime = (a: string, b: string) => (a > b ? a : b);

async function upsert(tx: Tx, s: LocalSource) {
  const store = tx.objectStore('sources');
  const current = await store.get(s.id);
  if (!current) {
    await store.put(toRecord(s, 0));
    return;
  }

  await store.put(
    toRecord(
      {
        ...s,
        // Keep the newer status but take the rest of the row.
        status: statusRank[current.status] > statusRank[s.status] ? current.status : s.status,
        errorMsg: s.errorMsg ?? current.errorMsg,
        size: s.size ?? current.size,
        updatedAt: maxTime(current.updatedAt, s.updatedAt),
        completedAt: s.completedAt ?? current.completedAt,
        kind: current.kind,
        ext: current.ext ?? s.ext,
      },
      current.finished
    )
  );
}

/** Inserts or updates whole rows (a new upload, a list page or a join result). */
export async function upsertSources(sources: LocalSource[]): Promise<void> {
  if (sources.length === 0) return;
  const tx = (await db()).transaction('sources', 'readwrite');
  for (const s of sources) await upsert(tx, s);
  await refreshFinished(
    tx,
    sources.map((s) => s.parentId ?? s.id)
  );
  await tx.done;
  changed();
}

/**
 * Applies a status event to a known source. Returns false when the source isn't in the database
 * (created on another device or missed while offline), so the caller can reconcile.
 */
export async function applyStatusEvent(event: SourceStatusEvent): Promise<boolean> {
  const tx = (await db()).transaction('sources', 'readwrite');
  const store = tx.objectStore('sources');
  const current = await store.get(event.id);
  if (!current) {
    await tx.done;
    return false;
  }

  const status =
    statusRank[current.status] > statusRank[event.status] ? current.status : event.status;
  const now = new Date().toISOString();
  await store.put({
    ...current,
    status,
    filename: event.filename || current.filename,
    size: event.size || current.size,
    errorMsg: event.error || current.errorMsg,
    updatedAt: now,
    completedAt: isTerminal(status) ? (current.completedAt ?? now) : current.completedAt,
  });
  await refreshFinished(tx, [current.parentId ?? current.id]);
  await tx.done;
  changed();
  return true;
}

/** Sets the status of sources the app itself just moved (e.g. queued after PATCH done). */
export async function advanceStatus(ids: string[], status: SourceStatus): Promise<void> {
  for (const id of ids) await applyStatusEvent({ id, status });
}

async function deleteTreesTx(tx: Tx, rootIds: string[]) {
  const store = tx.objectStore('sources');
  for (const rootId of rootIds) {
    const childIds = await store.index('byParent').getAllKeys(rootId);
    for (const id of [rootId, ...childIds]) await store.delete(id);
  }
}

export async function deleteTrees(rootIds: string[]): Promise<void> {
  if (rootIds.length === 0) return;
  const tx = (await db()).transaction('sources', 'readwrite');
  await deleteTreesTx(tx, rootIds);
  await tx.done;
  changed();
}

/** Removes uploads whose files were never confirmed (the presigned urls expire after 15 min). */
export async function pruneAbandoned(before: string): Promise<void> {
  const tx = (await db()).transaction('sources', 'readwrite');
  const store = tx.objectStore('sources');
  const roots = await store.index('byParent').getAll('');
  const abandoned: string[] = [];
  for (const root of roots) {
    if (root.createdAt >= before || root.status !== 'pending_upload') continue;
    const children = await store.index('byParent').getAll(root.id);
    if (children.every((c) => c.status === 'pending_upload')) abandoned.push(root.id);
  }
  await deleteTreesTx(tx, abandoned);
  await tx.done;
  if (abandoned.length > 0) changed();
}

async function loadTrees(conn: Db, roots: SourceRecord[]): Promise<UploadTree[]> {
  const tx = conn.transaction('sources');
  const index = tx.objectStore('sources').index('byParent');
  const trees = await Promise.all(
    roots.map(async (root) => ({
      root: fromRecord(root),
      children: (await index.getAll(root.id)).sort(byChildOrder).map(fromRecord),
    }))
  );
  await tx.done;
  return trees;
}

export async function getTree(id: string): Promise<UploadTree | null> {
  const conn = await db();
  const source = await conn.get('sources', id);
  if (!source) return null;
  const root = source.parentId ? await conn.get('sources', source.parentId) : source;
  return root ? (await loadTrees(conn, [root]))[0] : null;
}

/** Uploads still in progress, newest first. */
export async function listQueue(): Promise<UploadTree[]> {
  const conn = await db();
  const roots = await conn.getAllFromIndex('sources', 'byState', ['', 0]);
  return loadTrees(conn, roots.sort(newestFirst));
}

/** Finished uploads, newest first, after `before` when given. */
export async function listHistory(limit: number, before?: HistoryCursor): Promise<UploadTree[]> {
  const conn = await db();
  const roots = (await conn.getAllFromIndex('sources', 'byState', ['', 1]))
    .filter(
      (r) =>
        !before ||
        r.createdAt < before.createdAt ||
        (r.createdAt === before.createdAt && r.id < before.id)
    )
    .sort(newestFirst)
    .slice(0, limit);
  return loadTrees(conn, roots);
}

export async function isEmpty(): Promise<boolean> {
  return (await (await db()).count('sources')) === 0;
}

export async function getSyncState(): Promise<SyncState> {
  return (await (await db()).get('syncState', syncKey)) ?? emptySync;
}

export async function setSyncState(patch: Partial<SyncState>): Promise<void> {
  const conn = await db();
  const next = { ...(await getSyncState()), ...patch };
  await conn.put('syncState', next, syncKey);
}

/** Data is per user: signing out drops everything. */
export async function clearActivity(): Promise<void> {
  const conn = await db();
  const tx = conn.transaction(['sources', 'syncState'], 'readwrite');
  await tx.objectStore('sources').clear();
  await tx.objectStore('syncState').clear();
  await tx.done;
  changed();
}
