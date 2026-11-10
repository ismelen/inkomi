import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { uploadsApi } from '../../api/uploads-api';
import { ActivityLine, activityView, stepOrder, Tone } from '../../models/activity-view';
import { unfinishedIds, UploadTree } from '../../models/upload-tree';
import { advanceStatus } from '../../services/activity-db';
import { showError } from '../../stores/use-toast';
import { Button } from '../ui/button';
import { Modal } from '../ui/modal';
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
  const [confirming, setConfirming] = useState(false);
  const [cancelling, setCancelling] = useState(false);

  const cancel = async () => {
    setConfirming(false);
    setCancelling(true);
    try {
      await uploadsApi.cancel(tree.root.id);
      // the websocket brings the same statuses; this shows them right away
      await advanceStatus(unfinishedIds(tree), 'cancelled');
    } catch (err) {
      showError(err);
    } finally {
      setCancelling(false);
    }
  };

  return (
    <article className="rounded-lg bg-card px-4 py-3.5 shadow-card">
      <div className="flex items-center gap-3">
        <h3 className="min-w-0 flex-1 truncate text-md font-semibold">{tree.root.title}</h3>
        {view.counter && (
          <span className="rounded-md bg-surface px-1.5 font-mono text-xs text-primary tabular-nums">
            {view.counter.done}/{view.counter.total}
          </span>
        )}
        <Button
          variant="ghost"
          size="sm"
          icon="close"
          loading={cancelling}
          onClick={() => setConfirming(true)}
        >
          {t('activity.cancel')}
        </Button>
      </div>
      <p className={cx('mt-0.5 text-sm', toneText[view.tone])}>{lineText(view.line)}</p>
      {!queued && (
        <>
          <div className="mt-2.5 h-1 rounded-full bg-primary/10">
            <div
              className={cx('h-1 rounded-full transition-all', toneBar[view.tone])}
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
                        : 'text-faint'
                  )}
                >
                  {t(`activity.steps.${step}`)}
                </li>
              );
            })}
          </ol>
        </>
      )}
      <Modal
        open={confirming}
        onClose={() => setConfirming(false)}
        title={t('activity.cancelTitle')}
        size="sm"
        focusField={false}
        footer={
          <>
            <Button variant="secondary" onClick={() => setConfirming(false)}>
              {t('activity.cancelKeep')}
            </Button>
            <Button variant="danger" onClick={() => void cancel()}>
              {t('activity.cancel')}
            </Button>
          </>
        }
      >
        <p className="text-base text-muted">{t('activity.cancelMessage')}</p>
      </Modal>
    </article>
  );
}
