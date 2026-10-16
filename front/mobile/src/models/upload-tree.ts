import { BookFormat } from '../api/types/books';
import { SourceRow, SourceStatus } from '../api/types/uploads';

export type SourceKind = 'file' | 'folder' | 'library';

/** A source row as kept in the local database, plus what only the app knows about it. */
export interface LocalSource extends SourceRow {
  kind: SourceKind;
  /** Format picked by the user; the filename turns into .epub once processed. */
  ext: BookFormat | null;
}

/** An upload (root source) with its items and join results. */
export interface UploadTree {
  root: LocalSource;
  children: LocalSource[];
}

export function isTerminal(status: SourceStatus): boolean {
  return status === 'done' || status === 'sent' || status === 'failed';
}

export function itemsOf(tree: UploadTree): LocalSource[] {
  return tree.children.filter((c) => !c.isResult).sort((a, b) => a.position - b.position);
}

export function resultsOf(tree: UploadTree): LocalSource[] {
  return tree.children.filter((c) => c.isResult).sort((a, b) => a.position - b.position);
}

export function isFolder(tree: UploadTree): boolean {
  return !tree.root.isFile;
}

/** Items of a folder that finished their part: terminal, or merged into a result. */
export function settledItems(tree: UploadTree): LocalSource[] {
  return itemsOf(tree).filter((i) => isTerminal(i.status) || i.status === 'joined');
}

/**
 * A folder without join never leaves pending_upload on the server: it is finished when all its
 * items are. A joined folder is closed by the server, or stuck if an item failed before the join.
 */
export function isTreeFinished(tree: UploadTree): boolean {
  const { root } = tree;
  if (root.isFile) return isTerminal(root.status);
  if (isTerminal(root.status)) return true;

  const items = itemsOf(tree);
  if (items.length === 0) return false;
  if (!root.join) return items.every((i) => isTerminal(i.status));

  const results = resultsOf(tree);
  return (
    settledItems(tree).length === items.length &&
    items.some((i) => i.status === 'failed') &&
    results.every((r) => isTerminal(r.status))
  );
}

export function treeError(tree: UploadTree): string | null {
  if (tree.root.status === 'failed') return tree.root.errorMsg ?? '';
  const failed = tree.children.find((c) => c.status === 'failed');
  return failed ? (failed.errorMsg ?? '') : null;
}

export function treeFinishedAt(tree: UploadTree): string {
  const times = [tree.root, ...tree.children]
    .map((s) => s.completedAt)
    .filter((t): t is string => !!t)
    .sort();
  return times[times.length - 1] ?? tree.root.updatedAt;
}

/** Files the user can download: the root file, the join results or the items of a plain folder. */
export function downloadablesOf(tree: UploadTree): LocalSource[] {
  if (tree.root.isFile) return [tree.root];
  return tree.root.join ? resultsOf(tree) : itemsOf(tree);
}

export function isDownloadable(source: LocalSource): boolean {
  return source.isFile && (source.status === 'done' || source.status === 'sent');
}

export function treeSize(tree: UploadTree): number {
  return downloadablesOf(tree).reduce((total, s) => total + (s.size ?? 0), 0);
}
