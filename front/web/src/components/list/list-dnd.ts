import { DropMark } from './list-rows';
import { useList } from '../../stores/use-list';

export type DragKind = 'item' | 'group';

export interface DragData {
  kind: DragKind;
  /** Group of a nested item. */
  groupId?: string;
}

/**
 * What dropping `active` over a row means, from the pointer position inside that row: items
 * dropped on the middle of another item group with it, on a group go into it; near the edges
 * they are placed before or after. Groups only reorder.
 */
export function dropMarkFor(active: DragData, over: DragData, relativeY: number): DropMark {
  if (active.kind === 'group') return relativeY < 0.5 ? 'before' : 'after';
  if (over.kind === 'group') return relativeY < 0.3 ? 'before' : 'inside';
  if (relativeY < 0.25) return 'before';
  if (relativeY > 0.75) return 'after';
  return 'inside';
}

function containerOf(id: string): string | null {
  const { items } = useList.getState();
  return items[id]?.groupId ?? null;
}

function listOf(groupId: string | null): string[] {
  const { groups, rootOrder } = useList.getState();
  return groupId ? groups[groupId].itemIds : rootOrder;
}

/** Puts id before or after refId, in refId's list (a group or the root). */
function place(id: string, refId: string, after: boolean) {
  const list = useList.getState();
  const target = containerOf(refId);
  const without = listOf(target).filter((x) => x !== id);
  const index = without.indexOf(refId) + (after ? 1 : 0);

  const sameList = list.groups[id] ? target === null : containerOf(id) === target;
  if (sameList) {
    const next = [...without.slice(0, index), id, ...without.slice(index)];
    if (target) list.reorderGroup(target, next);
    else list.reorderRoot(next);
  } else {
    list.moveItem(id, target, index);
  }
}

export function applyDrop(
  activeId: string,
  active: DragData,
  overId: string,
  over: DragData,
  mark: DropMark,
  newGroupTitle: string
) {
  if (!mark || activeId === overId) return;
  const list = useList.getState();

  if (active.kind === 'group') {
    // Groups live at the root: a nested row stands for its group.
    const refId = over.kind === 'item' && over.groupId ? over.groupId : overId;
    if (refId !== activeId) place(activeId, refId, mark === 'after');
    return;
  }

  if (over.kind === 'group') {
    if (mark === 'inside') {
      if (active.groupId !== overId) list.moveItem(activeId, overId);
    } else {
      place(activeId, overId, false);
    }
    return;
  }

  if (mark === 'inside') list.groupWith(activeId, overId, newGroupTitle);
  else place(activeId, overId, mark === 'after');
}
