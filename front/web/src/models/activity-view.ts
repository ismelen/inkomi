import { SourceStatus } from '../api/types/uploads';
import {
  isFolder,
  isTerminal,
  itemsOf,
  LocalSource,
  resultsOf,
  settledItems,
  UploadTree,
} from './upload-tree';

export type StepKey = 'download' | 'process' | 'join' | 'kepub' | 'upload';
export type StepState = 'done' | 'active' | 'todo' | 'skip';
export type Tone = 'textSecondary' | 'warning' | 'info' | 'success' | 'danger';

export const stepOrder: StepKey[] = ['download', 'process', 'join', 'kepub', 'upload'];

export type ActivityLine =
  | { key: 'pendingUpload' | 'queued' | 'downloading' | 'joining' | 'kepub' | 'uploading' }
  | { key: 'processing'; done?: number; total?: number }
  | { key: 'waitingJoin'; done: number; total: number }
  | { key: 'failed'; error: string };

export interface ActivityView {
  counter?: { done: number; total: number };
  line: ActivityLine;
  tone: Tone;
  steps: Record<StepKey, StepState>;
  progress: number;
}

function needsProcessing(source: LocalSource): boolean {
  return source.ext ? source.ext !== 'epub' : source.kind !== 'library';
}

function applicableSteps(tree: UploadTree): Set<StepKey> {
  const nodes = isFolder(tree) ? itemsOf(tree) : [tree.root];
  const steps = new Set<StepKey>(['upload']);
  if (nodes.some((n) => n.kind === 'library')) steps.add('download');
  if (nodes.some(needsProcessing)) steps.add('process');
  if (!tree.root.isFile && tree.root.join) steps.add('join');
  if (tree.root.kepubify || nodes.some((n) => n.kepubify)) steps.add('kepub');
  return steps;
}

function activeStep(source: LocalSource): StepKey | null {
  switch (source.status) {
    case 'processing':
      return source.kind === 'library' ? 'download' : 'process';
    case 'joining':
    case 'joined':
      return 'join';
    case 'kepubifying':
      return 'kepub';
    case 'done':
      return 'upload';
    default:
      return null;
  }
}

function lineOf(status: SourceStatus, source: LocalSource): ActivityLine {
  switch (status) {
    case 'pending_upload':
      return { key: 'pendingUpload' };
    case 'processing':
      return source.kind === 'library' ? { key: 'downloading' } : { key: 'processing' };
    case 'joining':
    case 'joined':
      return { key: 'joining' };
    case 'kepubifying':
      return { key: 'kepub' };
    case 'done':
      return { key: 'uploading' };
    case 'failed':
      return { key: 'failed', error: source.errorMsg ?? '' };
    default:
      return { key: 'queued' };
  }
}

const toneOf: Record<ActivityLine['key'], Tone> = {
  pendingUpload: 'textSecondary',
  queued: 'textSecondary',
  downloading: 'warning',
  processing: 'warning',
  waitingJoin: 'info',
  joining: 'info',
  kepub: 'warning',
  uploading: 'success',
  failed: 'danger',
};

/** Node whose status describes the card: the root file, a join result or the busiest item. */
function currentNode(tree: UploadTree): LocalSource {
  if (tree.root.isFile) return tree.root;
  const pending = (nodes: LocalSource[]) => nodes.find((n) => !isTerminal(n.status));
  const results = resultsOf(tree);
  if (results.length > 0) return pending(results) ?? results[0];

  const items = itemsOf(tree);
  return (
    items.find((i) => i.status === 'processing' || i.status === 'kepubifying') ??
    pending(items.filter((i) => i.status !== 'joined')) ??
    items[0] ??
    tree.root
  );
}

export function activityView(tree: UploadTree): ActivityView {
  const applicable = applicableSteps(tree);
  const node = currentNode(tree);
  const folder = isFolder(tree);
  const items = itemsOf(tree);
  const settled = settledItems(tree).length;

  let line = lineOf(node.status, node);
  if (folder && tree.root.join && resultsOf(tree).length === 0 && items.length > 0) {
    if (items.some((i) => i.status === 'processing')) {
      line = { key: 'processing', done: settled, total: items.length };
    } else if (settled < items.length && settled > 0) {
      line = { key: 'waitingJoin', done: settled, total: items.length };
    }
  }

  const active = activeStep(node);
  const activeIndex = active ? stepOrder.indexOf(active) : -1;
  const steps = {} as Record<StepKey, StepState>;
  stepOrder.forEach((step, i) => {
    if (!applicable.has(step)) steps[step] = 'skip';
    else if (i < activeIndex) steps[step] = 'done';
    else if (i === activeIndex) steps[step] = 'active';
    else steps[step] = 'todo';
  });

  const ordered = stepOrder.filter((s) => applicable.has(s));
  const doneSteps = active ? ordered.indexOf(active) + 0.5 : 0;
  let progress = ordered.length ? doneSteps / ordered.length : 0;
  if (folder && items.length > 0 && !resultsOf(tree).length) {
    progress = Math.max(progress, (settled / items.length) * 0.8);
  }

  return {
    counter: folder ? { done: settled, total: items.length } : undefined,
    line,
    tone: toneOf[line.key],
    steps,
    progress: Math.min(1, Math.max(0.03, progress)),
  };
}
