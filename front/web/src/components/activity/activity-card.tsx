import { useTranslation } from 'react-i18next';
import { ActivityLine, activityView, stepOrder, Tone } from '../../models/activity-view';
import { UploadTree } from '../../models/upload-tree';
import { cx } from '../ui/cx';

export const toneText: Record<Tone, string> = {
  textSecondary: 'text-muted',
  warning: 'text-warning',
  info: 'text-info',
  success: 'text-success',
  danger: 'text-danger',
};

const toneBar: Record<Tone, string> = {
  textSecondary: 'bg-muted',
  warning: 'bg-warning',
  info: 'bg-info',
  success: 'bg-success',
  danger: 'bg-danger',
};

function useLineText() {
  const { t } = useTranslation();
  return (line: ActivityLine) => {
    switch (line.key) {
      case 'processing':
        return line.total
          ? t('activity.line.processingCount', { done: line.done, total: line.total })
          : t('activity.line.processing');
      case 'waitingJoin':
        return t('activity.line.waitingJoin', { done: line.done, total: line.total });
      case 'failed':
        return t('activity.line.failed', { error: line.error });
      default:
        return t(`activity.line.${line.key}`);
    }
  };
}

export function ActivityCard({ tree }: { tree: UploadTree }) {
  const { t } = useTranslation();
  const lineText = useLineText();
  const view = activityView(tree);
  const queued = view.line.key === 'queued' || view.line.key === 'pendingUpload';

  return (
    <article className="rounded-lg border border-strong px-4 py-3">
      <div className="flex items-center gap-3">
        <h3 className="min-w-0 flex-1 truncate text-md">{tree.root.title}</h3>
        {view.counter && (
          <span className="text-sm text-muted tabular-nums">
            {view.counter.done}/{view.counter.total}
          </span>
        )}
      </div>
      <p className={cx('mt-0.5 text-sm', toneText[view.tone])}>{lineText(view.line)}</p>
      {!queued && (
        <>
          <div className="mt-2 h-1.5 rounded-full bg-surface">
            <div
              className={cx('h-1.5 rounded-full transition-all', toneBar[view.tone])}
              style={{ width: `${Math.round(view.progress * 100)}%` }}
            />
          </div>
          <ol className="mt-2 grid grid-cols-5 text-xs">
            {stepOrder.map((step) => {
              const state = view.steps[step];
              return (
                <li
                  key={step}
                  className={cx(
                    'truncate',
                    state === 'active'
                      ? toneText[view.tone]
                      : state === 'done'
                        ? 'text-fg'
                        : 'text-line'
                  )}
                >
                  {t(`activity.steps.${step}`)}
                </li>
              );
            })}
          </ol>
        </>
      )}
    </article>
  );
}
