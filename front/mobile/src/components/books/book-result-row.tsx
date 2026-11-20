import React from 'react';
import { View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { colors, spacing } from '../../theme';
import { BookResponse } from '../../api/types/books';
import { bookLanguagesOf, isBookLanguage } from '../../models/book-languages';
import { SChip } from '../ui/s-chip';
import { SIconButton } from '../ui/s-icon-button';
import { SText } from '../ui/s-text';
import { BookCover } from './book-cover';

interface Props {
  book: BookResponse;
  added: boolean;
  onToggle: () => void;
  wished?: boolean;
  /** Adds the book to the wishlist; hidden when absent. */
  onWish?: () => void;
}

export function BookResultRow({ book, added, onToggle, wished, onWish }: Props) {
  const { t } = useTranslation();
  const language = bookLanguagesOf(book.language)
    .map((l) => (isBookLanguage(l) ? t(`bookLanguages.${l}`) : l))
    .join(', ');
  const subtitle = [book.author, language].filter(Boolean).join(' · ');
  const addable = !!book.md5;

  return (
    <View
      style={{
        flexDirection: 'row',
        alignItems: 'center',
        gap: spacing.md,
        paddingVertical: spacing.sm,
      }}
    >
      <BookCover
        url={book.coverUrl}
        accessibilityLabel={t('search.cover', { title: book.title })}
      />
      <View style={{ flex: 1, gap: spacing.xxs }}>
        <SText variant="label" numberOfLines={2}>
          {book.title}
        </SText>
        {!!subtitle && (
          <SText variant="small" tone="textSecondary" numberOfLines={1}>
            {subtitle}
          </SText>
        )}
        <View style={{ flexDirection: 'row', alignItems: 'center', gap: spacing.sm }}>
          {!!book.extension && <SChip label={book.extension.toUpperCase()} />}
          {book.pages > 0 && (
            <SText variant="small" tone="textSecondary">
              {t('search.pages', { count: book.pages })}
            </SText>
          )}
        </View>
      </View>
      {onWish && (
        <SIconButton
          icon={wished ? 'check' : 'favorite'}
          label={t('wishlist.wish')}
          size={20}
          disabled={wished}
          onPress={onWish}
        />
      )}
      {addable ? (
        <SIconButton
          icon={added ? 'check' : 'add'}
          label={book.title}
          color={added ? colors.success : colors.text}
          onPress={onToggle}
        />
      ) : (
        <SText variant="small" tone="textSecondary">
          {t('search.noMd5')}
        </SText>
      )}
    </View>
  );
}
