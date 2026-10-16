import { useTranslation } from 'react-i18next';
import { useToast } from '../../stores/use-toast';
import { cx } from './cx';
import { IconButton } from './icon-button';

export function Toaster() {
  const { t } = useTranslation();
  const { toasts, dismiss } = useToast();
  return (
    <div
      aria-live="polite"
      className="pointer-events-none fixed inset-x-4 bottom-20 z-[60] flex flex-col items-center gap-2 md:bottom-6"
    >
      {toasts.map((toast) => (
        <div
          key={toast.id}
          role={toast.tone === 'danger' ? 'alert' : 'status'}
          className={cx(
            'pointer-events-auto flex max-w-md items-center gap-2 rounded-input py-1.5 pr-1.5 pl-4 text-base shadow-modal',
            toast.tone === 'danger' ? 'bg-danger-soft text-danger' : 'bg-card text-fg'
          )}
        >
          <span className="flex-1">{toast.message}</span>
          <IconButton
            icon="close"
            label={t('common.close')}
            size={18}
            onClick={() => dismiss(toast.id)}
          />
        </div>
      ))}
    </div>
  );
}
