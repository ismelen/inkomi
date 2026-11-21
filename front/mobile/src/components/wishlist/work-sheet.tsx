import React, { useEffect, useState } from 'react';
import { ActivityIndicator, ScrollView, View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { errorKey } from '../../api/api-error';
import { wishlistApi } from '../../api/wishlist-api';
import { WorkDetails } from '../../api/types/wishlist';
import { openLibrarySearch } from '../../services/library-search';
import { colors, spacing } from '../../theme';
import { BookCover } from '../books/book-cover';
import { SButton } from '../ui/s-button';
import { SChip } from '../ui/s-chip';
import { SSheet } from '../ui/s-sheet';
import { SText } from '../ui/s-text';

/** What the list already knows of the book; the sheet fetches the rest by workKey. */
export interface SheetBook {
  title: string;
  author?: string;
  workKey?: string;
  coverUrl?: string;
}

interface Props {
  book: SheetBook | null;
  onClose: () => void;
}

/** Cover, synopsis and subjects of a wished book, with a shortcut to the library search. */
export function WorkSheet({ book, onClose }: Props) {
  const { t } = useTranslation();
  const [work, setWork] = useState<WorkDetails>();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>();

  useEffect(() => {
    setWork(undefined);
    setError(undefined);
    if (!book?.workKey) return;
    const ctrl = new AbortController();
    setLoading(true);
    wishlistApi
      .work(book.workKey, ctrl.signal)
      .then((found) => !ctrl.signal.aborted && setWork(found))
      .catch((err) => !ctrl.signal.aborted && setError(errorKey(err)))
      .finally(() => !ctrl.signal.aborted && setLoading(false));
    return () => ctrl.abort();
  }, [book?.workKey]);

  if (!book) return null;

  const cover = book.coverUrl || work?.coverUrl;
  const subtitle = [book.author, work?.firstPublished].filter(Boolean).join(' · ');
  const search = () => {
    onClose();
    openLibrarySearch(book.title, 'title');
  };

  return (
    <SSheet visible title={t('workSheet.title')} onClose={onClose}>
      <ScrollView style={{ flexGrow: 0 }} contentContainerStyle={{ gap: spacing.md }}>
        <View style={{ flexDirection: 'row', gap: spacing.md }}>
          <BookCover
            url={cover}
            width={96}
            height={136}
            accessibilityLabel={t('search.cover', { title: book.title })}
          />
          <View style={{ flex: 1, gap: spacing.xs }}>
            <SText variant="title">{book.title}</SText>
            {!!subtitle && (
              <SText variant="small" tone="textSecondary">
                {subtitle}
              </SText>
            )}
            {!!work?.subjects.length && (
              <View style={{ flexDirection: 'row', flexWrap: 'wrap', gap: spacing.xs }}>
                {work.subjects.slice(0, 4).map((s) => (
                  <SChip key={s} label={s} />
                ))}
              </View>
            )}
          </View>
        </View>
        {loading ? (
          <ActivityIndicator color={colors.primary} style={{ padding: spacing.lg }} />
        ) : (
          <SText tone={work?.description ? 'text' : 'textSecondary'}>
            {work?.description ||
              (error && error !== 'errors.workNotFound' ? t(error) : t('workSheet.noSynopsis'))}
          </SText>
        )}
      </ScrollView>
      <SButton label={t('wishlist.searchLibrary')} icon="search" onPress={search} />
    </SSheet>
  );
}
