import { uploadsApi } from '../api/uploads-api';
import { wsClient } from '../api/ws-client';
import { isResultEvent, SentPushData, SourceEvent } from '../api/types/events';
import { UploadDTO } from '../api/types/uploads';
import { LocalSource } from '../models/upload-tree';
import {
  applyStatusEvent,
  getSyncState,
  isEmpty,
  pruneAbandoned,
  setSyncState,
  toLocalSource,
  upsertSources,
} from './activity-db';

const pageSize = 50;
// Presigned PUT urls expire after 15 min; older unconfirmed uploads will never start.
const abandonedAfterMs = 20 * 60_000;

let running: Promise<void> | null = null;
let again = false;
let retryTimer: ReturnType<typeof setTimeout> | null = null;
let unsubscribe: (() => void)[] = [];

function treeRows(upload: UploadDTO): LocalSource[] {
  const { children, ...root } = upload;
  return [toLocalSource(root), ...children.map((c) => toLocalSource(c))];
}

async function storePage(uploads: UploadDTO[]) {
  await upsertSources(uploads.flatMap(treeRows));
}

/** First page of uploads into an empty database (new install, other device, after sign in). */
async function rebuild() {
  const page = await uploadsApi.list({ limit: pageSize });
  await storePage(page.uploads);
  await setSyncState({
    syncedAt: page.syncedAt,
    historyCursor: page.nextCursor,
    historyExhausted: page.nextCursor === null,
  });
}

/** Every tree changed since the last sync, so events missed while offline are applied. */
async function catchUp(syncedAt: string) {
  let cursor: string | undefined;
  let nextSyncedAt: string | null = null;
  do {
    const page = await uploadsApi.list({ limit: pageSize, cursor, updatedSince: syncedAt });
    // The first clock read is the earliest one: nothing changed after it is lost.
    nextSyncedAt ??= page.syncedAt;
    await storePage(page.uploads);
    cursor = page.nextCursor ?? undefined;
  } while (cursor);
  await setSyncState({ syncedAt: nextSyncedAt });
}

async function runReconcile() {
  const state = await getSyncState();
  if (!state.syncedAt || (await isEmpty())) await rebuild();
  else await catchUp(state.syncedAt);
  await pruneAbandoned(new Date(Date.now() - abandonedAfterMs).toISOString());
}

/** Rebuilds or catches up with GET /uploads. Concurrent calls collapse into one more run. */
export function reconcile(): Promise<void> {
  if (running) {
    again = true;
    return running;
  }
  running = runReconcile()
    .catch((err) => {
      console.warn('reconcile', err);
    })
    .finally(() => {
      running = null;
      if (again) {
        again = false;
        void reconcile();
      }
    });
  return running;
}

function reconcileSoon() {
  if (retryTimer) return;
  retryTimer = setTimeout(() => {
    retryTimer = null;
    void reconcile();
  }, 1_000);
}

export async function handleSourceEvent(event: SourceEvent): Promise<void> {
  if (isResultEvent(event)) {
    await upsertSources([toLocalSource(event.result)]);
    return;
  }
  if (!(await applyStatusEvent(event))) reconcileSoon();
}

export async function handleSentPush(data: SentPushData): Promise<void> {
  // Unknown ids are fetched by the catch up on the next websocket connection.
  await applyStatusEvent({ id: data.id, status: 'sent', filename: data.filename || undefined });
}

/** Older history from the server once the local one runs out. Returns false when exhausted. */
export async function loadOlderHistory(): Promise<boolean> {
  const state = await getSyncState();
  if (state.historyExhausted) return false;
  if (!state.historyCursor) {
    await reconcile();
    return !(await getSyncState()).historyExhausted;
  }

  const page = await uploadsApi.list({ limit: pageSize, cursor: state.historyCursor });
  await storePage(page.uploads);
  await setSyncState({
    historyCursor: page.nextCursor,
    historyExhausted: page.nextCursor === null,
  });
  return page.nextCursor !== null;
}

/** Opens the websocket first and reconciles on every connection, so no event falls in between. */
export function startSync() {
  stopSync();
  unsubscribe = [
    wsClient.onEvent((event) => void handleSourceEvent(event)),
    wsClient.onOpen(() => void reconcile()),
  ];
  wsClient.start();
}

export function stopSync() {
  unsubscribe.forEach((fn) => fn());
  unsubscribe = [];
  wsClient.stop();
}
