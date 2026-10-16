import { useNavigate } from 'react-router';
import { useTranslation } from 'react-i18next';
import { UploadFile, useUpload } from '../../stores/use-upload';
import { Button } from '../ui/button';
import { cx } from '../ui/cx';
import { Modal } from '../ui/modal';
import { Spinner } from '../ui/spinner';

function UploadRow({ file }: { file: UploadFile }) {
  const { t } = useTranslation();
  const percent = Math.round(file.progress * 100);
  const status = {
    waiting: { text: t('sending.waiting'), className: 'text-muted' },
    uploading: { text: `${percent}%`, className: 'text-muted' },
    uploaded: { text: t('sending.uploaded'), className: 'text-success' },
    failed: { text: t('sending.failed'), className: 'text-danger' },
  }[file.state];
  const bar = {
    waiting: 'bg-transparent',
    uploading: 'bg-primary',
    uploaded: 'bg-success',
    failed: 'bg-danger',
  }[file.state];

  return (
    <li className="py-2">
      <div className="flex items-baseline gap-3">
        <span className="min-w-0 flex-1 truncate text-base font-medium">{file.title}</span>
        <span className={cx('font-mono text-xs', status.className)}>{status.text}</span>
      </div>
      <div className="mt-1.5 h-1 rounded-full bg-primary/10">
        <div
          className={cx('h-1 rounded-full transition-all', bar)}
          style={{ width: file.state === 'uploading' ? `${percent}%` : '100%' }}
        />
      </div>
      {file.state === 'failed' && file.error && (
        <p className="mt-1 text-sm text-danger">{t(file.error)}</p>
      )}
    </li>
  );
}

export function SendingModal() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { visible, phase, files, error, cancel, close } = useUpload();
  const busy = phase === 'requesting' || phase === 'uploading';

  const viewQueue = () => {
    close();
    navigate('/activity');
  };

  return (
    <Modal
      open={visible}
      persistent={busy}
      size="lg"
      onClose={close}
      title={phase === 'finished' ? t('sending.finished') : t('sending.title')}
      footer={
        <>
          <Button variant="secondary" size="lg" className="flex-1" onClick={busy ? cancel : close}>
            {busy ? t('common.cancel') : t('common.close')}
          </Button>
          <Button
            size="lg"
            className="flex-1"
            disabled={phase === 'requesting'}
            onClick={viewQueue}
          >
            {t('sending.viewQueue')}
          </Button>
        </>
      }
    >
      {phase === 'requesting' && (
        <p className="flex items-center gap-2 py-3 text-base text-muted">
          <Spinner /> {t('sending.requesting')}
        </p>
      )}
      {error && <p className="py-2 text-base text-danger">{t(error)}</p>}
      <ul>
        {files.map((file) => (
          <UploadRow key={file.sourceId} file={file} />
        ))}
      </ul>
      {busy && <p className="pt-3 text-sm text-muted">{t('sending.keepOpen')}</p>}
    </Modal>
  );
}
