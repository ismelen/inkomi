import { FormEvent, useRef, useState } from 'react';
import { useNavigate } from 'react-router';
import { useTranslation } from 'react-i18next';
import { BookFormat, BookResponse } from '../api/types/books';
import { PageHeader, useBack } from '../components/layout/page-header';
import { Md5Modal } from '../components/list/md5-modal';
import { Button } from '../components/ui/button';
import { cx } from '../components/ui/cx';
import { TextInput } from '../components/ui/field';
import { Icon } from '../components/ui/icon';
import { IconButton } from '../components/ui/icon-button';
import { supportedFormats } from '../models/list-item';
import { isLibraryAdded, useList } from '../stores/use-list';
import { useSearch } from '../stores/use-search';

const debounceMs = 500;

function FormatTag({ extension }: { extension: string }) {
  if (!extension) return null;
  return (
    <span className="inline-block rounded-md bg-surface px-2 py-0.5 font-mono text-[11px] text-primary uppercase">
      {extension}
    </span>
  );
}

function AddToggle({ book, added }: { book: BookResponse; added: boolean }) {
  const { t } = useTranslation();
  if (!book.md5)
    return (
      <span className="justify-self-end text-right text-xs text-muted">{t('search.noMd5')}</span>
    );
  return (
    <button
      type="button"
      aria-pressed={added}
      aria-label={book.title}
      onClick={() => useList.getState().toggleLibrary(book)}
      className={cx(
        'inline-flex size-9 shrink-0 cursor-pointer items-center justify-center justify-self-end rounded-full transition',
        added
          ? 'bg-primary text-on-primary shadow-card'
          : 'bg-surface text-primary hover:bg-variant'
      )}
    >
      <Icon name={added ? 'check' : 'add'} size={18} />
    </button>
  );
}

function Cover({ book }: { book: BookResponse }) {
  return (
    <div className="flex h-[56px] w-10 shrink-0 items-center justify-center overflow-hidden rounded-md bg-surface text-primary shadow-card">
      {book.coverUrl ? (
        <img src={book.coverUrl} alt="" loading="lazy" className="size-full object-cover" />
      ) : (
        <Icon name="menu_book" size={18} />
      )}
    </div>
  );
}

function ResultRow({ book, added }: { book: BookResponse; added: boolean }) {
  const { t } = useTranslation();
  const meta = [book.language, book.pages > 0 ? t('search.pages', { count: book.pages }) : null]
    .filter(Boolean)
    .join(' · ');
  return (
    <li className="grid grid-cols-[1fr_auto] items-center gap-x-4 border-b border-surface py-3 last:border-b-0 md:grid-cols-[minmax(0,1fr)_90px_72px] lg:grid-cols-[minmax(0,1fr)_minmax(0,240px)_90px_72px]">
      <div className="flex min-w-0 items-center gap-3 md:gap-4">
        <Cover book={book} />
        <div className="min-w-0">
          <p className="line-clamp-2 text-md font-medium">{book.title}</p>
          <p className="truncate text-xs text-muted">
            <span className="lg:hidden">{[book.author, meta].filter(Boolean).join(' · ')}</span>
            <span className="max-lg:hidden">{meta}</span>
          </p>
          <div className="mt-1 md:hidden">
            <FormatTag extension={book.extension} />
          </div>
        </div>
      </div>
      <p className="truncate text-sm text-muted max-lg:hidden">{book.author}</p>
      <div className="max-md:hidden">
        <FormatTag extension={book.extension} />
      </div>
      <AddToggle book={book} added={added} />
    </li>
  );
}

export function SearchPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const goBack = useBack();
  const { query, format, results, status, error, setQuery, setFormat, search } = useSearch();
  const items = useList((s) => s.items);
  const [md5Open, setMd5Open] = useState(false);
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const libraryCount = Object.values(items).filter((i) => i.kind === 'library').length;

  const onChangeQuery = (text: string) => {
    setQuery(text);
    if (timer.current) clearTimeout(timer.current);
    timer.current = setTimeout(() => void search(), debounceMs);
  };

  const submit = (e: FormEvent) => {
    e.preventDefault();
    if (timer.current) clearTimeout(timer.current);
    void search();
  };

  const formats: (BookFormat | null)[] = [null, ...supportedFormats];

  return (
    <>
      <PageHeader title={t('search.title')} className="max-md:hidden">
        <Button variant="secondary" icon="search" onClick={() => setMd5Open(true)}>
          <span className="max-sm:hidden">{t('search.addByMd5')}</span>
          <span className="sm:hidden">{t('list.md5')}</span>
        </Button>
      </PageHeader>
      <div className="flex flex-col gap-3 px-4 pt-3 max-md:pt-2 md:px-8 lg:flex-row lg:items-center lg:gap-6">
        {/* Phones: the app's search screen, back and the box in the header. */}
        <form
          role="search"
          onSubmit={submit}
          className="flex items-center gap-1 max-md:-ml-2 lg:w-[600px]"
        >
          <IconButton
            icon="chevron_left"
            label={t('common.back')}
            size={24}
            className="text-fg md:hidden"
            onClick={goBack}
          />
          <TextInput
            className="min-w-0 flex-1"
            autoFocus
            type="text"
            enterKeyHint="search"
            aria-label={t('search.placeholder')}
            placeholder={t('search.placeholder')}
            value={query}
            onChange={(e) => onChangeQuery(e.target.value)}
            left={<Icon name="search" className="text-primary" />}
            right={
              query ? (
                <IconButton
                  icon="close"
                  label={t('search.clear')}
                  size={18}
                  className="-mr-1.5 size-8"
                  onClick={() => onChangeQuery('')}
                />
              ) : undefined
            }
          />
        </form>
        <div className="flex gap-2">
          {formats.map((f) => (
            <button
              key={f ?? 'all'}
              type="button"
              aria-pressed={format === f}
              onClick={() => setFormat(f)}
              className={cx(
                'h-8 cursor-pointer rounded-full px-4 text-sm font-medium transition',
                format === f
                  ? 'bg-primary text-on-primary shadow-card'
                  : 'bg-card text-muted shadow-card hover:text-primary'
              )}
            >
              {f ? f.toUpperCase() : t('search.all')}
            </button>
          ))}
        </div>
      </div>

      <div className="min-h-0 flex-1 overflow-y-auto px-4 pt-4 pb-6 md:px-8">
        {results.length > 0 && (
          <div className="rounded-lg bg-card px-4 shadow-card">
            <div className="grid grid-cols-[minmax(0,1fr)_90px_72px] gap-x-4 border-b border-surface pt-3 pb-2 text-section text-muted uppercase max-md:hidden lg:grid-cols-[minmax(0,1fr)_minmax(0,240px)_90px_72px]">
              <span>{t('search.colTitle')}</span>
              <span className="max-lg:hidden">{t('search.colAuthor')}</span>
              <span>{t('search.colFormat')}</span>
              <span />
            </div>
            <ul>
              {results.map((book, i) => (
                <ResultRow
                  key={book.md5 || `${book.title}-${i}`}
                  book={book}
                  added={isLibraryAdded(items, book.md5)}
                />
              ))}
            </ul>
          </div>
        )}
        {status === 'loading' && <p className="pt-3 text-xs text-muted">{t('search.searching')}</p>}
        {status === 'error' && error && <p className="pt-6 text-center text-danger">{t(error)}</p>}
        {status === 'done' && results.length === 0 && (
          <p className="pt-10 text-center text-muted">{t('search.empty')}</p>
        )}
        {status === 'idle' && <p className="pt-10 text-center text-muted">{t('search.idle')}</p>}
      </div>

      {libraryCount > 0 && (
        <div className="flex shrink-0 items-center gap-4 bg-card px-4 py-3 shadow-bar md:px-8">
          <p className="flex-1 text-base font-medium">
            {t('search.added', { count: libraryCount })}
          </p>
          <Button size="lg" onClick={() => navigate('/')}>
            {t('search.goToList')}
          </Button>
        </div>
      )}
      <Md5Modal open={md5Open} onClose={() => setMd5Open(false)} />
    </>
  );
}
