import { DragEvent, useEffect, useMemo, useState } from 'react';
import {
  DndContext,
  DragEndEvent,
  DragMoveEvent,
  DragOverlay,
  DragStartEvent,
  MouseSensor,
  pointerWithin,
  TouchSensor,
  useDraggable,
  useDroppable,
  useSensor,
  useSensors,
} from '@dnd-kit/core';
import { Link, useNavigate } from 'react-router';
import { useTranslation } from 'react-i18next';
import { Brand } from '../components/layout/nav';
import { PageHeader } from '../components/layout/page-header';
import { AddMenu } from '../components/list/add-menu';
import { applyDrop, DragData, dropMarkFor } from '../components/list/list-dnd';
import { DropMark, GroupRow, ItemRow } from '../components/list/list-rows';
import { MoveModal } from '../components/list/move-modal';
import { OptionsPanel } from '../components/options/options-panel';
import { Button } from '../components/ui/button';
import { Checkbox } from '../components/ui/checkbox';
import { cx } from '../components/ui/cx';
import { Icon } from '../components/ui/icon';
import { IconButton } from '../components/ui/icon-button';
import { Modal } from '../components/ui/modal';
import { Switch } from '../components/ui/switch';
import { desktopQuery, phoneQuery, useMediaQuery } from '../hooks/use-media-query';
import { Group } from '../models/group';
import { ListItem } from '../models/list-item';
import { readDrop } from '../services/file-picker';
import { useActivity } from '../stores/use-activity';
import { useList } from '../stores/use-list';
import { useProfile } from '../stores/use-profile';
import { useSettings } from '../stores/use-settings';
import { showError, useToast } from '../stores/use-toast';
import { useUpload } from '../stores/use-upload';

type Row =
  | { type: 'group'; group: Group }
  | { type: 'item'; item: ListItem; groupId?: string; groupRtl?: boolean };

interface Drop {
  overId: string;
  over: DragData;
  mark: DropMark;
}

function rowId(row: Row): string {
  return row.type === 'group' ? row.group.id : row.item.id;
}

function rowData(row: Row): DragData {
  return row.type === 'group' ? { kind: 'group' } : { kind: 'item', groupId: row.groupId };
}

function pointerY(event: Event): number | null {
  if ('touches' in event) {
    const touch = (event as TouchEvent).touches[0] ?? (event as TouchEvent).changedTouches[0];
    return touch?.clientY ?? null;
  }
  return 'clientY' in event ? (event as MouseEvent).clientY : null;
}

/** A row that can be dragged and dropped onto. */
function DndRow({
  row,
  drop,
  draggingId,
  focusedId,
  selection,
  onFocus,
}: {
  row: Row;
  drop: Drop | null;
  draggingId: string | null;
  focusedId: string | null;
  selection: string[];
  onFocus: (id: string) => void;
}) {
  const id = rowId(row);
  const data = rowData(row);
  const draggable = useDraggable({ id, data });
  const droppable = useDroppable({ id, data });
  const setNodeRef = (node: HTMLElement | null) => {
    draggable.setNodeRef(node);
    droppable.setNodeRef(node);
  };
  const { selecting, startSelection, toggleSelected } = useList.getState();
  const common = {
    selected: selection.includes(id),
    focused: focusedId === id,
    dropMark: drop?.overId === id ? drop.mark : null,
    dragging: draggingId === id,
    onToggleSelected: () => (selecting ? toggleSelected(id) : startSelection(id)),
    onFocus: () => onFocus(id),
    dragProps: { ...draggable.attributes, ...draggable.listeners, role: undefined },
    setNodeRef,
  };

  if (row.type === 'group') {
    const items = useList.getState().items;
    const formats = [
      ...new Set(row.group.itemIds.map((i) => items[i]?.ext.toUpperCase()).filter(Boolean)),
    ];
    return <GroupRow group={row.group} formats={formats} {...common} />;
  }
  return <ItemRow item={row.item} nested={!!row.groupId} groupRtl={row.groupRtl} {...common} />;
}

function hasFiles(e: DragEvent) {
  return e.dataTransfer.types.includes('Files');
}

export function ListPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const desktop = useMediaQuery(desktopQuery);
  const phone = useMediaQuery(phoneQuery);
  const { items, groups, rootOrder, selecting, selection } = useList();
  const queueCount = useActivity((s) => s.queue.length);
  const toCloudDefault = useSettings((s) => s.toCloudDefault);
  const dropboxLinked = useProfile((s) => s.profile?.dropbox.linked ?? false);
  const sendUpload = useUpload((s) => s.send);
  const sending = useUpload((s) => s.phase === 'requesting' || s.phase === 'uploading');

  const [toCloud, setToCloud] = useState(toCloudDefault);
  const [focusedId, setFocusedId] = useState<string | null>(null);
  const [moveOpen, setMoveOpen] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [filesOver, setFilesOver] = useState(false);
  const [draggingId, setDraggingId] = useState<string | null>(null);
  const [drop, setDrop] = useState<Drop | null>(null);

  useEffect(() => setToCloud(toCloudDefault), [toCloudDefault]);

  // The focused element may leave the list (removed, sent).
  useEffect(() => {
    if (focusedId && !items[focusedId] && !groups[focusedId]) setFocusedId(null);
  }, [focusedId, items, groups]);

  const rows = useMemo<Row[]>(
    () =>
      rootOrder.flatMap((id): Row[] => {
        const group = groups[id];
        if (group) {
          const children = group.collapsed
            ? []
            : group.itemIds
                .filter((itemId) => items[itemId])
                .map((itemId) => ({
                  type: 'item' as const,
                  item: items[itemId],
                  groupId: id,
                  groupRtl: group.rtl,
                }));
          return [{ type: 'group', group }, ...children];
        }
        return items[id] ? [{ type: 'item', item: items[id] }] : [];
      }),
    [rootOrder, groups, items]
  );

  const itemCount = Object.keys(items).length;

  const sensors = useSensors(
    useSensor(MouseSensor, { activationConstraint: { distance: 6 } }),
    useSensor(TouchSensor, { activationConstraint: { delay: 250, tolerance: 8 } })
  );

  const onDragStart = (e: DragStartEvent) => setDraggingId(String(e.active.id));

  // dnd-kit's `over` can lag behind rows that moved; the collision and the live rect don't.
  const onDragMove = (e: DragMoveEvent) => {
    const { active, activatorEvent, delta, collisions } = e;
    const startY = pointerY(activatorEvent);
    const overId = collisions?.[0] ? String(collisions[0].id) : '';
    const row = rows.find((r) => rowId(r) === overId);
    const node = row && document.querySelector(`[data-row-id="${overId}"]`);
    if (!row || !node || overId === active.id || startY === null) return setDrop(null);

    const rect = node.getBoundingClientRect();
    const relativeY = (startY + delta.y - rect.top) / rect.height;
    const over = rowData(row);
    const mark = dropMarkFor(active.data.current as DragData, over, relativeY);
    setDrop((prev) =>
      prev?.overId === overId && prev.mark === mark ? prev : { overId, over, mark }
    );
  };

  const onDragEnd = (e: DragEndEvent) => {
    if (drop) {
      applyDrop(
        String(e.active.id),
        e.active.data.current as DragData,
        drop.overId,
        drop.over,
        drop.mark,
        t('group.defaultTitle')
      );
    }
    setDraggingId(null);
    setDrop(null);
  };

  const onDragCancel = () => {
    setDraggingId(null);
    setDrop(null);
  };

  const onFileDrop = async (e: DragEvent) => {
    if (!hasFiles(e)) return;
    e.preventDefault();
    setFilesOver(false);
    try {
      const { files, folders } = await readDrop(e.dataTransfer);
      const list = useList.getState();
      if (files.length) list.addFiles(files);
      folders.forEach((folder) => {
        if (folder.files.length === 0) useToast.getState().show(t('add.emptyFolder'), 'danger');
        else list.addFolder(folder);
      });
    } catch (err) {
      showError(err);
    }
  };

  const toggleCloud = (value: boolean) => {
    if (!value || dropboxLinked) return setToCloud(value);
    // Linking leaves the page for Dropbox; it comes back to the list.
    useProfile.getState().linkDropbox().catch(showError);
  };

  const dragged = draggingId ? (items[draggingId]?.title ?? groups[draggingId]?.title) : null;
  const allSelected = rootOrder.length > 0 && selection.length >= rootOrder.length;

  // Phones select like the app: while selecting, a tap toggles instead of opening the options.
  const focusRow = (id: string) => {
    if (phone && selecting) useList.getState().toggleSelected(id);
    else setFocusedId(id);
  };

  const optionsPanel = focusedId && (
    <OptionsPanel
      key={focusedId}
      targetId={focusedId}
      onClose={desktop ? undefined : () => setFocusedId(null)}
      onSaved={desktop ? undefined : () => setFocusedId(null)}
    />
  );

  return (
    <div
      className="relative flex min-h-0 flex-1 flex-col"
      onDragOver={(e) => {
        if (!hasFiles(e)) return;
        e.preventDefault();
        setFilesOver(true);
      }}
      onDragLeave={(e) => {
        if (e.currentTarget === e.target || !e.currentTarget.contains(e.relatedTarget as Node))
          setFilesOver(false);
      }}
      onDrop={(e) => void onFileDrop(e)}
    >
      {/* Phones: the app's home header (sections as icons) and its search + attach row. */}
      <header className="flex h-14 shrink-0 items-center gap-1 px-2 md:hidden">
        {selecting ? (
          <>
            <IconButton
              icon="close"
              label={t('home.cancelSelection')}
              size={24}
              className="text-fg"
              onClick={() => useList.getState().clearSelection()}
            />
            <h1 className="min-w-0 flex-1 truncate text-lg font-semibold">
              {t('home.selected', { count: selection.length })}
            </h1>
            <IconButton
              icon="checklist"
              label={t('home.toggleAll')}
              size={24}
              className="text-fg"
              onClick={() => useList.getState().toggleSelectAll()}
            />
          </>
        ) : (
          <>
            <Brand className="flex-1 pl-2" />
            <IconButton
              icon="history"
              label={t('home.history')}
              size={24}
              className="text-fg"
              onClick={() => navigate('/history')}
            />
            <IconButton
              icon="progress_activity"
              label={t('home.activity')}
              size={24}
              badge={queueCount}
              className="text-fg"
              onClick={() => navigate('/activity')}
            />
            <IconButton
              icon="settings"
              label={t('home.settings')}
              size={24}
              className="text-fg"
              onClick={() => navigate('/settings')}
            />
          </>
        )}
      </header>
      {!selecting && (
        <div className="flex shrink-0 gap-2 px-4 md:hidden">
          <Link
            to="/search"
            className="flex h-11 min-w-0 flex-1 items-center gap-2 rounded-input bg-subtle px-3 text-base text-faint"
          >
            <Icon name="search" className="text-primary" />
            <span className="truncate">{t('home.searchPlaceholder')}</span>
          </Link>
          <AddMenu compact />
        </div>
      )}

      <PageHeader title={t('list.title')} className="max-md:hidden">
        <label className="flex items-center gap-2 text-sm font-medium text-muted">
          {t('home.dropbox')}
          <Switch label={t('home.dropbox')} checked={toCloud} onChange={toggleCloud} />
        </label>
        <Button
          size="lg"
          className="min-w-[110px] lg:min-w-[140px]"
          disabled={itemCount === 0 || sending}
          onClick={() => void sendUpload(toCloud)}
        >
          {itemCount > 0 ? t('list.sendCount', { count: itemCount }) : t('list.send')}
        </Button>
      </PageHeader>

      <div className="flex min-h-0 flex-1 gap-6 px-4 pt-3 md:px-8 md:pt-5">
        <div className="flex min-w-0 flex-1 flex-col lg:max-w-[600px]">
          <div className="hidden min-h-[84px] items-center gap-4 rounded-lg border-2 border-dashed border-line bg-card px-5 md:flex">
            <span className="inline-flex size-12 shrink-0 items-center justify-center rounded-full bg-primary text-on-primary shadow-card">
              <Icon name="attach_file" size={24} />
            </span>
            <div className="min-w-0 flex-1">
              <p className="text-md font-semibold">{t('list.dropTitle')}</p>
              <p className="text-xs text-muted">{t('list.dropHint')}</p>
            </div>
            <AddMenu />
          </div>

          <div className="-mx-1 min-h-0 md:mt-5 flex-1 overflow-y-auto px-1 pb-4">
            {rows.length === 0 ? (
              <p className="px-4 py-16 text-center text-base text-muted">{t('list.empty')}</p>
            ) : (
              <DndContext
                sensors={sensors}
                collisionDetection={pointerWithin}
                onDragStart={onDragStart}
                onDragMove={onDragMove}
                onDragEnd={onDragEnd}
                onDragCancel={onDragCancel}
              >
                <ul className="flex flex-col gap-2">
                  {rows.map((row, i) => {
                    const next = rows[i + 1];
                    const lastOfGroup =
                      row.type === 'item' &&
                      row.groupId &&
                      !(next?.type === 'item' && next.groupId);
                    return (
                      <li
                        key={rowId(row)}
                        data-row-id={rowId(row)}
                        className={cx(
                          row.type === 'item' &&
                            row.groupId &&
                            'relative before:absolute before:top-0 before:-bottom-2 before:left-0 before:w-1 before:rounded-full before:bg-selected',
                          lastOfGroup && 'mb-2 before:bottom-0'
                        )}
                      >
                        <DndRow
                          row={row}
                          drop={drop}
                          draggingId={draggingId}
                          focusedId={focusedId}
                          selection={selection}
                          onFocus={focusRow}
                        />
                      </li>
                    );
                  })}
                </ul>
                <DragOverlay dropAnimation={null}>
                  {dragged && (
                    <div className="max-w-sm truncate rounded-input bg-card px-3 py-2.5 text-base font-medium shadow-modal">
                      {dragged}
                    </div>
                  )}
                </DragOverlay>
              </DndContext>
            )}
          </div>

          {rows.length > 0 && (
            <div className="mb-6 flex min-h-12 shrink-0 items-center gap-3 rounded-lg bg-card px-3 shadow-raised max-md:hidden">
              <Checkbox
                checked={allSelected}
                onChange={() => useList.getState().toggleSelectAll()}
                label={t('home.toggleAll')}
              />
              <span className="min-w-0 flex-1 truncate text-sm font-medium">
                {selection.length > 0 ? t('home.selected', { count: selection.length }) : ''}
              </span>
              <Button
                variant="ghost"
                size="sm"
                icon="checklist_rtl"
                onClick={() => useList.getState().toggleSelectAll()}
              >
                <span className="max-sm:hidden">{t('list.selectAll')}</span>
              </Button>
              <Button
                variant="ghost"
                size="sm"
                icon="drive_file_move"
                disabled={selection.length === 0}
                onClick={() => setMoveOpen(true)}
              >
                <span className="max-sm:hidden">{t('common.move')}</span>
              </Button>
              <Button
                variant="ghost"
                size="sm"
                icon="delete"
                className="text-danger"
                disabled={selection.length === 0}
                onClick={() => setConfirmDelete(true)}
              >
                <span className="max-sm:hidden">{t('common.delete')}</span>
              </Button>
            </div>
          )}
        </div>

        {desktop && (
          <aside className="mb-6 w-[372px] shrink-0 overflow-hidden rounded-lg bg-card shadow-card">
            {optionsPanel || (
              <p className="p-8 text-center text-base text-muted">{t('list.noFocus')}</p>
            )}
          </aside>
        )}
      </div>

      {/* Phones: the app's bottom bar (Dropbox + send, or delete + move while selecting). */}
      <div
        className="flex shrink-0 items-center gap-3 bg-card px-4 pt-3 shadow-bar md:hidden"
        style={{ paddingBottom: 'calc(12px + env(safe-area-inset-bottom))' }}
      >
        {selecting ? (
          <>
            <Button
              variant="secondary"
              size="lg"
              icon="delete"
              className="flex-1"
              disabled={selection.length === 0}
              onClick={() => setConfirmDelete(true)}
            >
              {t('common.delete')}
            </Button>
            <Button
              variant="secondary"
              size="lg"
              icon="drive_file_move"
              className="flex-1"
              disabled={selection.length === 0}
              onClick={() => setMoveOpen(true)}
            >
              {t('common.move')}
            </Button>
          </>
        ) : (
          <>
            <label className="flex items-center gap-2 text-base font-medium">
              <Switch label={t('home.dropbox')} checked={toCloud} onChange={toggleCloud} />
              {t('home.dropbox')}
            </label>
            <Button
              size="lg"
              className="flex-1"
              disabled={itemCount === 0 || sending}
              onClick={() => void sendUpload(toCloud)}
            >
              {itemCount > 0 ? t('list.sendCount', { count: itemCount }) : t('list.send')}
            </Button>
          </>
        )}
      </div>

      {!desktop && (
        <Modal
          open={!!focusedId}
          onClose={() => setFocusedId(null)}
          size="md"
          tall
          focusField={false}
        >
          <div className="-mx-5 -mb-5 h-[min(760px,88dvh)] md:-mx-6">{optionsPanel}</div>
        </Modal>
      )}

      {filesOver && (
        <div className="pointer-events-none absolute inset-2 z-30 flex items-center justify-center rounded-lg border-2 border-dashed border-primary bg-bg/85 text-lg font-semibold text-primary">
          {t('list.dropNow')}
        </div>
      )}

      <MoveModal open={moveOpen} onClose={() => setMoveOpen(false)} />
      <Modal
        open={confirmDelete}
        onClose={() => setConfirmDelete(false)}
        size="sm"
        title={t('delete.title', { count: selection.length })}
        footer={
          <>
            <Button
              variant="secondary"
              size="lg"
              className="flex-1"
              onClick={() => setConfirmDelete(false)}
            >
              {t('common.cancel')}
            </Button>
            <Button
              variant="danger"
              size="lg"
              className="flex-1"
              onClick={() => {
                const list = useList.getState();
                list.remove(selection);
                list.clearSelection();
                setConfirmDelete(false);
              }}
            >
              {t('common.delete')}
            </Button>
          </>
        }
      >
        <p className="text-base text-muted md:text-center">{t('delete.message')}</p>
      </Modal>
    </div>
  );
}
