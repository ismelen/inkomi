import { HTMLAttributes, Ref } from 'react';
import { useTranslation } from 'react-i18next';
import { Group } from '../../models/group';
import { isManga, ListItem } from '../../models/list-item';
import { findEReader } from '../../models/ereaders';
import { useList } from '../../stores/use-list';
import { useSettings } from '../../stores/use-settings';
import { formatBytes } from '../../utils/format';
import { Checkbox } from '../ui/checkbox';
import { cx } from '../ui/cx';
import { Icon } from '../ui/icon';
import { IconButton } from '../ui/icon-button';

/** Where a dragged row would land on this one. */
export type DropMark = 'before' | 'after' | 'inside' | null;

interface RowShellProps extends HTMLAttributes<HTMLDivElement> {
  ref?: Ref<HTMLDivElement>;
  focused: boolean;
  dropMark: DropMark;
  dragging: boolean;
}

function RowShell({ focused, dropMark, dragging, className, children, ...rest }: RowShellProps) {
  return (
    <div
      className={cx(
        'relative flex min-h-12 cursor-pointer touch-manipulation items-center gap-3 rounded-input border px-3 py-1.5 shadow-card transition select-none',
        focused ? 'border-primary bg-selected' : 'border-transparent bg-card hover:shadow-raised',
        dropMark === 'inside' && 'ring-2 ring-primary',
        dragging && 'opacity-40',
        className
      )}
      {...rest}
    >
      {dropMark === 'before' && (
        <span className="absolute inset-x-0 -top-[5px] h-0.5 rounded-full bg-primary" />
      )}
      {dropMark === 'after' && (
        <span className="absolute inset-x-0 -bottom-[5px] h-0.5 rounded-full bg-primary" />
      )}
      {children}
    </div>
  );
}

interface CommonProps {
  selected: boolean;
  focused: boolean;
  dropMark: DropMark;
  dragging: boolean;
  onToggleSelected: () => void;
  onFocus: () => void;
  /** dnd-kit's activator: listeners, attributes and the node ref. */
  dragProps: HTMLAttributes<HTMLElement>;
  setNodeRef: (node: HTMLElement | null) => void;
}

export function useItemDetails(item: ListItem, inheritedRtl?: boolean): string {
  const { t } = useTranslation();
  const kepub = useSettings((s) => findEReader(s.ereaderKey).kepub);
  const rtl = item.groupId ? inheritedRtl : item.rtl;
  return [
    item.ext.toUpperCase(),
    item.kind === 'file' ? formatBytes(item.size) : (item.author ?? t('home.library')),
    isManga(item.ext) && rtl ? t('common.rtl') : null,
    kepub ? t('list.kepubify') : null,
  ]
    .filter(Boolean)
    .join(' · ');
}

export function ItemRow({
  item,
  nested,
  groupRtl,
  selected,
  focused,
  dropMark,
  dragging,
  onToggleSelected,
  onFocus,
  dragProps,
  setNodeRef,
}: CommonProps & { item: ListItem; nested: boolean; groupRtl?: boolean }) {
  const { t } = useTranslation();
  const details = useItemDetails(item, groupRtl);
  return (
    <RowShell
      ref={setNodeRef}
      focused={focused}
      dropMark={dropMark}
      dragging={dragging}
      onClick={onFocus}
      className={nested ? 'ml-4' : undefined}
      {...dragProps}
    >
      {nested ? (
        <Icon name="drag_indicator" size={18} className="-mx-1 cursor-grab text-faint" />
      ) : (
        <Checkbox
          checked={selected}
          onChange={onToggleSelected}
          label={t('list.select', { title: item.title })}
        />
      )}
      <div className="min-w-0 flex-1">
        <p className="truncate text-base font-medium">{item.title}</p>
        <p className="truncate font-mono text-[11px] text-muted md:text-xs">{details}</p>
      </div>
      {!nested && isManga(item.ext) && (
        <IconButton
          icon="tune"
          label={t('options.title')}
          className="size-8"
          onClick={(e) => {
            e.stopPropagation();
            onFocus();
          }}
        />
      )}
      <IconButton
        icon="close"
        label={t('list.remove')}
        className="size-8"
        onClick={(e) => {
          e.stopPropagation();
          useList.getState().remove([item.id]);
        }}
      />
    </RowShell>
  );
}

export function GroupRow({
  group,
  formats,
  selected,
  focused,
  dropMark,
  dragging,
  onToggleSelected,
  onFocus,
  dragProps,
  setNodeRef,
}: CommonProps & { group: Group; formats: string[] }) {
  const { t } = useTranslation();
  const details = [
    t('common.files', { count: group.itemIds.length }),
    formats.join(', '),
    group.join ? t('group.join') : null,
  ]
    .filter(Boolean)
    .join(' · ');
  const toggle = () => useList.getState().updateGroup(group.id, { collapsed: !group.collapsed });

  return (
    <RowShell
      ref={setNodeRef}
      focused={focused}
      dropMark={dropMark}
      dragging={dragging}
      onClick={onFocus}
      {...dragProps}
    >
      <Checkbox
        checked={selected}
        onChange={onToggleSelected}
        label={t('list.select', { title: group.title })}
      />
      <IconButton
        icon={group.collapsed ? 'chevron_right' : 'expand_more'}
        label={group.collapsed ? t('list.expand') : t('list.collapse')}
        aria-expanded={!group.collapsed}
        className="-mx-2 size-7"
        onClick={(e) => {
          e.stopPropagation();
          toggle();
        }}
      />
      <div className="min-w-0 flex-1">
        <p className="truncate text-base font-semibold">{group.title}</p>
        <p
          className={cx(
            'truncate font-mono text-[11px] md:text-xs',
            group.join ? 'text-info' : 'text-muted'
          )}
        >
          {details}
        </p>
      </div>
      <IconButton
        icon="tune"
        label={t('options.title')}
        className="size-8"
        onClick={(e) => {
          e.stopPropagation();
          onFocus();
        }}
      />
    </RowShell>
  );
}
