import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  downloadablesOf,
  isDownloadable,
  isFolder,
  treeError,
  treeFinishedAt,
  treeSize,
  UploadTree,
} from '../../models/upload-tree';
import { formatBytes, formatWhen } from '../../utils/format';
import { Button } from '../ui/button';
import { cx } from '../ui/cx';
import { IconButton } from '../ui/icon-button';

interface Props {
  tree: UploadTree;
  busyId: string | null;
  onDownload: (sourceId: string) => void;
}

export function HistoryRow({ tree, busyId, onDownload }: Props) {
  const { t, i18n } = useTranslation();
  const [open, setOpen] = useState(false);
  const error = treeError(tree);
  const files = downloadablesOf(tree);
  const multiple = isFolder(tree);
  const single = !multiple && isDownloadable(tree.root) ? tree.root : null;
  const summary = [
    tree.root.join && multiple
      ? t('history.joined')
      : t('history.books', { count: Math.max(files.length, 1) }),
    formatBytes(treeSize(tree)),
    formatWhen(treeFinishedAt(tree), i18n.language),
  ]
    .filter(Boolean)
    .join(' · ');

  return (
    <li className="border-b border-surface last:border-b-0">
      <div className="flex items-center gap-3 py-2.5">
        <div className="min-w-0 flex-1">
          <p className="truncate text-base font-medium">{tree.root.title}</p>
          <p className="truncate text-xs text-muted">{summary}</p>
          {error !== null && (
            <p className="truncate text-xs text-danger">
              {error ? t('history.error', { error }) : t('history.errorUnknown')}
            </p>
          )}
        </div>
        {error === null && (
          <span className="flex items-center gap-1.5 text-sm font-medium text-success">
            <span className="size-1.5 rounded-full bg-success" />
            {t('history.sent')}
          </span>
        )}
        {single && (
          <IconButton
            icon="download"
            label={t('history.download')}
            disabled={busyId !== null}
            onClick={() => onDownload(single.id)}
          />
        )}
        {multiple && files.length > 0 && (
          <IconButton
            icon={open ? 'expand_less' : 'expand_more'}
            label={t('history.results')}
            aria-expanded={open}
            onClick={() => setOpen((o) => !o)}
          />
        )}
      </div>
      {open && (
        <ul className="mb-2.5 flex flex-col gap-1 rounded-input bg-subtle p-2">
          {files.map((file) => (
            <li key={file.id} className="flex items-center gap-3 pl-2">
              <span
                className={cx(
                  'min-w-0 flex-1 truncate text-sm',
                  !isDownloadable(file) && 'text-muted'
                )}
              >
                {file.title}
                {file.size ? ` · ${formatBytes(file.size)}` : ''}
              </span>
              <Button
                variant="ghost"
                size="sm"
                icon="download"
                loading={busyId === file.id}
                disabled={!isDownloadable(file) || busyId !== null}
                onClick={() => onDownload(file.id)}
              >
                {t('history.download')}
              </Button>
            </li>
          ))}
        </ul>
      )}
    </li>
  );
}
