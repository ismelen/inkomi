import { FormEvent, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { BookFormat } from '../../api/types/books';
import { supportedFormats } from '../../models/list-item';
import { useList } from '../../stores/use-list';
import { useToast } from '../../stores/use-toast';
import { Button } from '../ui/button';
import { SectionLabel, TextInput } from '../ui/field';
import { Modal } from '../ui/modal';
import { Segmented } from '../ui/segmented';

const md5Pattern = /^[a-f0-9]{32}$/i;

/** A library book added by its md5: the server downloads it like a search result. */
export function Md5Modal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const { t } = useTranslation();
  const [md5, setMd5] = useState('');
  const [title, setTitle] = useState('');
  const [format, setFormat] = useState<BookFormat>('epub');
  const [touched, setTouched] = useState(false);

  const value = md5.trim();
  const valid = md5Pattern.test(value);

  const close = () => {
    setMd5('');
    setTitle('');
    setFormat('epub');
    setTouched(false);
    onClose();
  };

  const submit = (e: FormEvent) => {
    e.preventDefault();
    setTouched(true);
    if (!valid) return;
    useList.getState().addLibrary(value.toLowerCase(), title.trim() || value, format);
    useToast.getState().show(t('search.added', { count: 1 }));
    close();
  };

  return (
    <Modal open={open} onClose={close} title={t('md5.title')} size="sm">
      <form onSubmit={submit} className="flex flex-col gap-4">
        <TextInput
          label={t('md5.md5')}
          placeholder={t('md5.placeholder')}
          value={md5}
          spellCheck={false}
          autoComplete="off"
          className="font-mono"
          onChange={(e) => setMd5(e.target.value)}
          onBlur={() => setTouched(true)}
          error={touched && value && !valid ? t('md5.invalid') : undefined}
        />
        <TextInput
          label={t('md5.bookTitle')}
          placeholder={t('md5.bookTitlePlaceholder')}
          value={title}
          onChange={(e) => setTitle(e.target.value)}
        />
        <div>
          <SectionLabel className="mb-1.5">{t('md5.format')}</SectionLabel>
          <Segmented
            label={t('md5.format')}
            value={format}
            onChange={setFormat}
            options={supportedFormats.map((f) => ({ value: f, label: f.toUpperCase() }))}
          />
        </div>
        <div className="mt-1 flex gap-3">
          <Button variant="secondary" size="lg" className="flex-1" onClick={close}>
            {t('common.cancel')}
          </Button>
          <Button type="submit" size="lg" className="flex-1" disabled={!valid}>
            {t('md5.add')}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
