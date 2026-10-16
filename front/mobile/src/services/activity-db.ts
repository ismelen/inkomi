import * as SQLite from 'expo-sqlite';
import { SourceStatusEvent } from '../api/types/events';
import { SourceRow, SourceStatus } from '../api/types/uploads';
import { formatOf } from '../models/list-item';
import { isTerminal, isTreeFinished, LocalSource, UploadTree } from '../models/upload-tree';

const dbName = 'inkomi.db';

const schema = `
  DROP TABLE IF EXISTS transactions;
  CREATE TABLE IF NOT EXISTS sources (
    id TEXT PRIMARY KEY NOT NULL,
    parentId TEXT,
    title TEXT NOT NULL,
    filename TEXT NOT NULL,
    kepubify INTEGER NOT NULL DEFAULT 0,
    toCloud INTEGER NOT NULL DEFAULT 0,
    isRtl INTEGER NOT NULL DEFAULT 0,
    "join" INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL,
    errorMsg TEXT,
    size INTEGER,
    isFile INTEGER NOT NULL DEFAULT 1,
    createdAt TEXT NOT NULL,
    updatedAt TEXT NOT NULL,
    completedAt TEXT,
    position INTEGER NOT NULL DEFAULT 0,
    isResult INTEGER NOT NULL DEFAULT 0,
    kind TEXT NOT NULL,
    ext TEXT,
    finished INTEGER NOT NULL DEFAULT 0
  );
  CREATE INDEX IF NOT EXISTS idx_sources_parent ON sources (parentId, position);
  CREATE INDEX IF NOT EXISTS idx_sources_roots ON sources (parentId, finished, createdAt DESC, id DESC);
  CREATE TABLE IF NOT EXISTS sync_state (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    syncedAt TEXT,
    historyCursor TEXT,
    historyExhausted INTEGER NOT NULL DEFAULT 0
  );
  INSERT OR IGNORE INTO sync_state (id) VALUES (1);
`;

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

interface SourceRecord {
  id: string;
  parentId: string | null;
  title: string;
  filename: string;
  kepubify: number;
  toCloud: number;
  isRtl: number;
  join: number;
  status: SourceStatus;
  errorMsg: string | null;
  size: number | null;
  isFile: number;
  createdAt: string;
  updatedAt: string;
  completedAt: string | null;
  position: number;
  isResult: number;
  kind: LocalSource['kind'];
  ext: LocalSource['ext'];
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

type Listener = () => void;
const listeners = new Set<Listener>();
let dbPromise: Promise<SQLite.SQLiteDatabase> | null = null;

function db(): Promise<SQLite.SQLiteDatabase> {
  dbPromise ??= SQLite.openDatabaseAsync(dbName).then(async (conn) => {
    await conn.execAsync(schema);
    return conn;
  });
  return dbPromise;
}

function changed() {
  listeners.forEach((listener) => listener());
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
  return {
    id: r.id,
    parentId: r.parentId,
    title: r.title,
    filename: r.filename,
    kepubify: !!r.kepubify,
    toCloud: !!r.toCloud,
    isRtl: !!r.isRtl,
    join: !!r.join,
    status: r.status,
    errorMsg: r.errorMsg,
    size: r.size,
    isFile: !!r.isFile,
    createdAt: r.createdAt,
    updatedAt: r.updatedAt,
    completedAt: r.completedAt,
    position: r.position,
    isResult: !!r.isResult,
    kind: r.kind,
    ext: r.ext,
  };
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

async function rootIdOf(conn: SQLite.SQLiteDatabase, id: string): Promise<string | null> {
  const row = await conn.getFirstAsync<{ id: string; parentId: string | null }>(
    'SELECT id, parentId FROM sources WHERE id = ?',
    [id]
  );
  if (!row) return null;
  return row.parentId ?? row.id;
}

async function loadTree(conn: SQLite.SQLiteDatabase, rootId: string): Promise<UploadTree | null> {
  const root = await conn.getFirstAsync<SourceRecord>(
    'SELECT * FROM sources WHERE id = ? AND parentId IS NULL',
    [rootId]
  );
  if (!root) return null;
  const children = await conn.getAllAsync<SourceRecord>(
    'SELECT * FROM sources WHERE parentId = ? ORDER BY isResult, position',
    [rootId]
  );
  return { root: fromRecord(root), children: children.map(fromRecord) };
}

async function loadTrees(conn: SQLite.SQLiteDatabase, roots: SourceRecord[]) {
  if (roots.length === 0) return [];
  const placeholders = roots.map(() => '?').join(', ');
  const children = await conn.getAllAsync<SourceRecord>(
    `SELECT * FROM sources WHERE parentId IN (${placeholders}) ORDER BY isResult, position`,
    roots.map((r) => r.id)
  );
  const byParent = new Map<string, LocalSource[]>();
  children.forEach((c) => {
    const list = byParent.get(c.parentId!) ?? [];
    list.push(fromRecord(c));
    byParent.set(c.parentId!, list);
  });
  return roots.map((r) => ({ root: fromRecord(r), children: byParent.get(r.id) ?? [] }));
}

async function refreshFinished(conn: SQLite.SQLiteDatabase, rootIds: Iterable<string>) {
  for (const rootId of new Set(rootIds)) {
    const tree = await loadTree(conn, rootId);
    if (!tree) continue;
    await conn.runAsync('UPDATE sources SET finished = ? WHERE id = ?', [
      isTreeFinished(tree) ? 1 : 0,
      rootId,
    ]);
  }
}

async function upsert(conn: SQLite.SQLiteDatabase, s: LocalSource) {
  const current = await conn.getFirstAsync<{ status: SourceStatus }>(
    'SELECT status FROM sources WHERE id = ?',
    [s.id]
  );
  if (current && statusRank[current.status] > statusRank[s.status]) {
    // Keep the newer status but take the rest of the row.
    s = { ...s, status: current.status };
  }

  await conn.runAsync(
    `INSERT INTO sources (
      id, parentId, title, filename, kepubify, toCloud, isRtl, "join", status, errorMsg, size,
      isFile, createdAt, updatedAt, completedAt, position, isResult, kind, ext
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT (id) DO UPDATE SET
      parentId = excluded.parentId,
      title = excluded.title,
      filename = excluded.filename,
      kepubify = excluded.kepubify,
      toCloud = excluded.toCloud,
      isRtl = excluded.isRtl,
      "join" = excluded."join",
      status = excluded.status,
      errorMsg = COALESCE(excluded.errorMsg, sources.errorMsg),
      size = COALESCE(excluded.size, sources.size),
      isFile = excluded.isFile,
      createdAt = excluded.createdAt,
      updatedAt = MAX(sources.updatedAt, excluded.updatedAt),
      completedAt = COALESCE(excluded.completedAt, sources.completedAt),
      position = excluded.position,
      isResult = excluded.isResult,
      kind = sources.kind,
      ext = COALESCE(sources.ext, excluded.ext)`,
    [
      s.id,
      s.parentId,
      s.title,
      s.filename,
      s.kepubify ? 1 : 0,
      s.toCloud ? 1 : 0,
      s.isRtl ? 1 : 0,
      s.join ? 1 : 0,
      s.status,
      s.errorMsg,
      s.size,
      s.isFile ? 1 : 0,
      s.createdAt,
      s.updatedAt,
      s.completedAt,
      s.position,
      s.isResult ? 1 : 0,
      s.kind,
      s.ext,
    ]
  );
}

/** Inserts or updates whole rows (a new upload, a list page or a join result). */
export async function upsertSources(sources: LocalSource[]): Promise<void> {
  if (sources.length === 0) return;
  const conn = await db();
  await conn.withTransactionAsync(async () => {
    for (const s of sources) await upsert(conn, s);
    await refreshFinished(
      conn,
      sources.map((s) => s.parentId ?? s.id)
    );
  });
  changed();
}

/**
 * Applies a status event to a known source. Returns false when the source isn't in the database
 * (created on another device or missed while offline), so the caller can reconcile.
 */
export async function applyStatusEvent(event: SourceStatusEvent): Promise<boolean> {
  const conn = await db();
  const current = await conn.getFirstAsync<SourceRecord>('SELECT * FROM sources WHERE id = ?', [
    event.id,
  ]);
  if (!current) return false;

  const status =
    statusRank[current.status] > statusRank[event.status] ? current.status : event.status;
  const now = new Date().toISOString();
  await conn.withTransactionAsync(async () => {
    await conn.runAsync(
      `UPDATE sources SET
        status = ?,
        filename = COALESCE(?, filename),
        size = COALESCE(?, size),
        errorMsg = COALESCE(?, errorMsg),
        updatedAt = ?,
        completedAt = CASE WHEN ? THEN COALESCE(completedAt, ?) ELSE completedAt END
      WHERE id = ?`,
      [
        status,
        event.filename || null,
        event.size || null,
        event.error || null,
        now,
        isTerminal(status) ? 1 : 0,
        now,
        event.id,
      ]
    );
    await refreshFinished(conn, [current.parentId ?? current.id]);
  });
  changed();
  return true;
}

/** Sets the status of sources the app itself just moved (e.g. queued after PATCH done). */
export async function advanceStatus(ids: string[], status: SourceStatus): Promise<void> {
  for (const id of ids) await applyStatusEvent({ id, status });
}

export async function deleteTrees(rootIds: string[]): Promise<void> {
  if (rootIds.length === 0) return;
  const conn = await db();
  const placeholders = rootIds.map(() => '?').join(', ');
  await conn.runAsync(
    `DELETE FROM sources WHERE id IN (${placeholders}) OR parentId IN (${placeholders})`,
    [...rootIds, ...rootIds]
  );
  changed();
}

/** Removes uploads whose files were never confirmed (the presigned urls expire after 15 min). */
export async function pruneAbandoned(before: string): Promise<void> {
  const conn = await db();
  await conn.runAsync(
    `DELETE FROM sources WHERE id IN (
      SELECT r.id FROM sources AS r
      WHERE r.parentId IS NULL AND r.createdAt < ? AND r.status = 'pending_upload'
        AND NOT EXISTS (SELECT 1 FROM sources AS c WHERE c.parentId = r.id AND c.status != 'pending_upload')
    ) OR parentId IN (
      SELECT r.id FROM sources AS r
      WHERE r.parentId IS NULL AND r.createdAt < ? AND r.status = 'pending_upload'
        AND NOT EXISTS (SELECT 1 FROM sources AS c WHERE c.parentId = r.id AND c.status != 'pending_upload')
    )`,
    [before, before]
  );
  changed();
}

export async function getTree(id: string): Promise<UploadTree | null> {
  const conn = await db();
  const rootId = await rootIdOf(conn, id);
  return rootId ? loadTree(conn, rootId) : null;
}

/** Uploads still in progress, newest first. */
export async function listQueue(): Promise<UploadTree[]> {
  const conn = await db();
  const roots = await conn.getAllAsync<SourceRecord>(
    'SELECT * FROM sources WHERE parentId IS NULL AND finished = 0 ORDER BY createdAt DESC, id DESC'
  );
  return loadTrees(conn, roots);
}

/** Finished uploads, newest first, after `before` when given. */
export async function listHistory(limit: number, before?: HistoryCursor): Promise<UploadTree[]> {
  const conn = await db();
  const roots = before
    ? await conn.getAllAsync<SourceRecord>(
        `SELECT * FROM sources WHERE parentId IS NULL AND finished = 1
          AND (createdAt < ? OR (createdAt = ? AND id < ?))
        ORDER BY createdAt DESC, id DESC LIMIT ?`,
        [before.createdAt, before.createdAt, before.id, limit]
      )
    : await conn.getAllAsync<SourceRecord>(
        `SELECT * FROM sources WHERE parentId IS NULL AND finished = 1
        ORDER BY createdAt DESC, id DESC LIMIT ?`,
        [limit]
      );
  return loadTrees(conn, roots);
}

export async function isEmpty(): Promise<boolean> {
  const conn = await db();
  const row = await conn.getFirstAsync<{ n: number }>('SELECT COUNT(*) AS n FROM sources');
  return (row?.n ?? 0) === 0;
}

export async function getSyncState(): Promise<SyncState> {
  const conn = await db();
  const row = await conn.getFirstAsync<{
    syncedAt: string | null;
    historyCursor: string | null;
    historyExhausted: number;
  }>('SELECT syncedAt, historyCursor, historyExhausted FROM sync_state WHERE id = 1');
  return {
    syncedAt: row?.syncedAt ?? null,
    historyCursor: row?.historyCursor ?? null,
    historyExhausted: !!row?.historyExhausted,
  };
}

export async function setSyncState(patch: Partial<SyncState>): Promise<void> {
  const conn = await db();
  const next = { ...(await getSyncState()), ...patch };
  await conn.runAsync(
    'UPDATE sync_state SET syncedAt = ?, historyCursor = ?, historyExhausted = ? WHERE id = 1',
    [next.syncedAt, next.historyCursor, next.historyExhausted ? 1 : 0]
  );
}

/** Data is per user: signing out drops everything. */
export async function clearActivity(): Promise<void> {
  const conn = await db();
  await conn.execAsync(
    'DELETE FROM sources; UPDATE sync_state SET syncedAt = NULL, historyCursor = NULL, historyExhausted = 0;'
  );
  changed();
}
