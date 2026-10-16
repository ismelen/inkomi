import { ReactNode, useEffect, useRef } from 'react';
import { createPortal } from 'react-dom';
import { useTranslation } from 'react-i18next';
import { cx } from './cx';
import { IconButton } from './icon-button';

interface Props {
  open: boolean;
  onClose: () => void;
  title?: string;
  children: ReactNode;
  footer?: ReactNode;
  /** Width on desktop; on phones the modal is a bottom sheet. */
  size?: 'sm' | 'md' | 'lg';
  /** Esc and the scrim don't close it (an upload in progress). */
  persistent?: boolean;
  /** Full height sheet on phones (long forms). */
  tall?: boolean;
  /** Focus the first field when opening (default) or just the dialog (no keyboard on phones). */
  focusField?: boolean;
}

const widths = { sm: 'md:max-w-sm', md: 'md:max-w-md', lg: 'md:max-w-xl' };

export function Modal({
  open,
  onClose,
  title,
  children,
  footer,
  size = 'md',
  persistent,
  tall,
  focusField = true,
}: Props) {
  const { t } = useTranslation();
  const panel = useRef<HTMLDivElement>(null);
  const closeRef = useRef(onClose);
  closeRef.current = onClose;

  useEffect(() => {
    if (!open) return;
    const previous = document.activeElement as HTMLElement | null;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && !persistent) closeRef.current();
    };
    document.addEventListener('keydown', onKey);
    const first = focusField
      ? panel.current?.querySelector<HTMLElement>('input, select, textarea, [data-autofocus]')
      : null;
    (first ?? panel.current)?.focus();
    return () => {
      document.removeEventListener('keydown', onKey);
      previous?.focus?.();
    };
  }, [open, persistent, focusField]);

  if (!open) return null;

  return createPortal(
    <div className="fixed inset-0 z-50 flex items-end justify-center bg-scrim md:items-center md:p-6">
      <div className="absolute inset-0" onClick={persistent ? undefined : onClose} />
      <div
        ref={panel}
        role="dialog"
        aria-modal
        aria-label={title}
        tabIndex={-1}
        className={cx(
          'pb-safe relative flex w-full flex-col overflow-hidden rounded-t-xl bg-card shadow-modal outline-none md:rounded-xl',
          tall ? 'h-[92dvh] md:h-auto md:max-h-[88dvh]' : 'max-h-[88dvh]',
          widths[size]
        )}
      >
        {title && (
          <div className="flex items-center gap-2 px-5 pt-5 pb-3 md:px-6 md:pt-6">
            <h2 className="flex-1 text-xl font-semibold md:text-center">{title}</h2>
            {!persistent && (
              <IconButton
                icon="close"
                label={t('common.close')}
                onClick={onClose}
                className="md:hidden"
              />
            )}
          </div>
        )}
        <div className="min-h-0 flex-1 overflow-y-auto px-5 pb-5 md:px-6">{children}</div>
        {footer && <div className="flex gap-3 px-5 pb-5 md:px-6 md:pb-6">{footer}</div>}
      </div>
    </div>,
    document.body
  );
}
