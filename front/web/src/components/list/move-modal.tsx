import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useList } from '../../stores/use-list';
import { Button } from '../ui/button';
import { cx } from '../ui/cx';
import { TextInput } from '../ui/field';
import { Modal } from '../ui/modal';

const ungrouped = '__ungrouped__';
const newGroup = '__new__';

function Choice({
  label,
  detail,
  selected,
  onClick,
}: {
  label: string;
  detail?: string;
  selected: boolean;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      role="radio"
      aria-checked={selected}
      onClick={onClick}
      className={cx(
        'flex w-full cursor-pointer items-center gap-3 rounded-md border px-3 py-2.5 text-left',
        selected ? 'border-strong bg-surface' : 'border-line hover:bg-hover'
      )}
    >
      <span className="min-w-0 flex-1 truncate text-base">{label}</span>
      {detail && <span className="text-sm text-muted">{detail}</span>}
    </button>
  );
}

export function MoveModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const { t } = useTranslation();
  const groups = useList((s) => s.groups);
  const rootOrder = useList((s) => s.rootOrder);
  const selection = useList((s) => s.selection);
  const [target, setTarget] = useState(ungrouped);
  const [name, setName] = useState('');

  const close = () => {
    setTarget(ungrouped);
    setName('');
    onClose();
  };

  const confirm = () => {
    const { move, createGroup } = useList.getState();
    let groupId: string | null = target === ungrouped ? null : target;
    if (target === newGroup) groupId = createGroup(name.trim() || t('group.defaultTitle'));
    move(selection, groupId);
    close();
  };

  const groupIds = rootOrder.filter((id) => groups[id] && !selection.includes(id));

  return (
    <Modal
      open={open}
      onClose={close}
      title={t('move.title', { count: selection.length })}
      size="sm"
      footer={
        <>
          <Button variant="secondary" size="lg" className="flex-1" onClick={close}>
            {t('common.cancel')}
          </Button>
          <Button size="lg" className="flex-1" onClick={confirm}>
            {t('common.move')}
          </Button>
        </>
      }
    >
      <div role="radiogroup" className="flex flex-col gap-2">
        <Choice
          label={t('move.ungrouped')}
          selected={target === ungrouped}
          onClick={() => setTarget(ungrouped)}
        />
        {groupIds.map((id) => (
          <Choice
            key={id}
            label={groups[id].title}
            detail={t('common.files', { count: groups[id].itemIds.length })}
            selected={target === id}
            onClick={() => setTarget(id)}
          />
        ))}
        <Choice
          label={t('move.newGroup')}
          selected={target === newGroup}
          onClick={() => setTarget(newGroup)}
        />
        {target === newGroup && (
          <TextInput
            autoFocus
            aria-label={t('move.newGroupPlaceholder')}
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder={t('move.newGroupPlaceholder')}
          />
        )}
      </div>
    </Modal>
  );
}
