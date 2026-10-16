import { useEffect, useMemo, useRef, useState } from 'react';
import { Navigate, useLocation } from 'react-router';
import { useTranslation } from 'react-i18next';
import { wsClient } from '../api/ws-client';
import { ActivityCard } from '../components/activity/activity-card';
import { HistoryRow } from '../components/history/history-row';
import { PageHeader } from '../components/layout/page-header';
import { cx } from '../components/ui/cx';
import { SectionLabel } from '../components/ui/field';
import { phoneQuery, useMediaQuery } from '../hooks/use-media-query';
import { treeFinishedAt, UploadTree } from '../models/upload-tree';
import { downloadSource } from '../services/download-service';
import { useActivity } from '../stores/use-activity';
import { showError } from '../stores/use-toast';
import { dayBucket, DayBucket } from '../utils/format';

const bucketOrder: DayBucket[] = ['today', 'yesterday', 'week', 'older'];

function useRealtime(): boolean {
  const [connected, setConnected] = useState(wsClient.isConnected());
  useEffect(() => wsClient.onStatus(setConnected), []);
  return connected;
}

function InProgress() {
  const { t } = useTranslation();
  const queue = useActivity((s) => s.queue);
  const connected = useRealtime();
  return (
    <section className="min-w-0">
      <div className="mb-2.5 flex items-center gap-3">
        <SectionLabel className="flex-1 max-md:hidden">{t('activity.title')}</SectionLabel>
        <span className="flex items-center gap-1.5 text-xs text-muted">
          <span className={cx('size-1.5 rounded-full', connected ? 'bg-success' : 'bg-line')} />
          {connected ? t('activity.realtime') : t('activity.offline')}
        </span>
      </div>
      {queue.length === 0 ? (
        <p className="py-10 text-center text-base text-muted">{t('activity.empty')}</p>
      ) : (
        <div className="flex flex-col gap-3">
          {queue.map((tree) => (
            <ActivityCard key={tree.root.id} tree={tree} />
          ))}
        </div>
      )}
    </section>
  );
}

function History() {
  const { t } = useTranslation();
  const { history, historyEnd, loadingMore, loadMoreHistory } = useActivity();
  const [busyId, setBusyId] = useState<string | null>(null);
  const sentinel = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const el = sentinel.current;
    if (!el) return;
    const observer = new IntersectionObserver((entries) => {
      if (entries.some((e) => e.isIntersecting)) void useActivity.getState().loadMoreHistory();
    });
    observer.observe(el);
    return () => observer.disconnect();
  }, []);

  const sections = useMemo(() => {
    const byBucket = new Map<DayBucket, UploadTree[]>();
    history.forEach((tree) => {
      const bucket = dayBucket(treeFinishedAt(tree));
      byBucket.set(bucket, [...(byBucket.get(bucket) ?? []), tree]);
    });
    return bucketOrder
      .filter((b) => byBucket.has(b))
      .map((b) => ({ key: b, trees: byBucket.get(b)! }));
  }, [history]);

  // The presigned link is asked on every click: it expires and must never be cached.
  const download = async (sourceId: string) => {
    setBusyId(sourceId);
    try {
      await downloadSource(sourceId);
    } catch (err) {
      showError(err);
    } finally {
      setBusyId(null);
    }
  };

  return (
    <section className="min-w-0">
      <SectionLabel className="mb-2.5 max-md:hidden">{t('history.title')}</SectionLabel>
      {sections.map((section) => (
        <div key={section.key} className="mb-4">
          <SectionLabel className="pt-1">{t(`history.${section.key}`)}</SectionLabel>
          <ul className="mt-1.5 rounded-lg bg-card px-4 shadow-card">
            {section.trees.map((tree) => (
              <HistoryRow
                key={tree.root.id}
                tree={tree}
                busyId={busyId}
                onDownload={(id) => void download(id)}
              />
            ))}
          </ul>
        </div>
      ))}
      {history.length === 0 && !loadingMore && historyEnd && (
        <p className="py-10 text-center text-base text-muted">{t('history.empty')}</p>
      )}
      <div ref={sentinel} className="h-4" />
      {loadingMore && !historyEnd && (
        <p className="text-xs text-muted">{t('history.loadingMore')}</p>
      )}
      {!loadingMore && !historyEnd && history.length > 0 && (
        <button
          type="button"
          className="cursor-pointer text-sm font-semibold text-primary hover:underline"
          onClick={() => void loadMoreHistory()}
        >
          {t('history.loadingMore')}
        </button>
      )}
    </section>
  );
}

/**
 * Desktop and tablet: "En curso" and the history side by side. Phones follow the app: two screens
 * (/activity and /history) opened from the list's header.
 */
export function ActivityPage() {
  const { t } = useTranslation();
  const { pathname } = useLocation();
  const phone = useMediaQuery(phoneQuery);
  const history = pathname === '/history';

  if (!phone && history) return <Navigate to="/activity" replace />;

  return (
    <>
      <PageHeader
        back
        title={
          phone ? (history ? t('history.title') : t('activity.title')) : t('activity.pageTitle')
        }
      />
      <div className="min-h-0 flex-1 overflow-y-auto px-4 pt-2 pb-6 md:px-8 md:pt-5">
        <div className="grid gap-8 md:grid-cols-[minmax(0,560px)_minmax(0,1fr)] lg:gap-10">
          <div className={cx(history && 'max-md:hidden')}>
            <InProgress />
          </div>
          <div className={cx(!history && 'max-md:hidden')}>
            <History />
          </div>
        </div>
      </div>
    </>
  );
}
