import { ReactNode, useEffect } from 'react';
import { IconName } from '../icons/icons';
import { cx } from './cx';
import { Icon } from './icon';

interface Props {
  open: boolean;
  onClose: () => void;
  children: ReactNode;
  className?: string;
}

/** Dropdown under its (relative) parent on desktop; bottom sheet on phones. */
export function Menu({ open, onClose, children, className }: Props) {
  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && onClose();
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, [open, onClose]);

  if (!open) return null;
  return (
    <>
      <div className="fixed inset-0 z-40 bg-scrim md:bg-transparent" onClick={onClose} />
      <div
        role="menu"
        className={cx(
          'pb-safe fixed inset-x-0 bottom-0 z-50 rounded-t-lg border border-strong bg-bg p-2 md:absolute md:inset-x-auto md:top-full md:right-0 md:bottom-auto md:mt-1 md:w-64 md:rounded-lg md:p-1.5',
          className
        )}
      >
        {children}
      </div>
    </>
  );
}

interface ItemProps {
  icon: IconName;
  label: string;
  hint?: string;
  onSelect: () => void;
}

export function MenuItem({ icon, label, hint, onSelect }: ItemProps) {
  return (
    <button
      type="button"
      role="menuitem"
      onClick={onSelect}
      className="flex w-full cursor-pointer items-center gap-3 rounded-md px-2.5 py-2.5 text-left hover:bg-surface md:py-2"
    >
      <Icon name={icon} size={22} />
      <span className="min-w-0">
        <span className="block text-md">{label}</span>
        {hint && <span className="block text-xs text-muted">{hint}</span>}
      </span>
    </button>
  );
}
