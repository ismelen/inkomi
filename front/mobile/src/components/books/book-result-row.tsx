import React from 'react';
import { Image, View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { colors, radii, spacing } from '../../theme';
import { BookResponse } from '../../api/types/books';
import { SChip } from '../ui/s-chip';
import { SIcon } from '../ui/s-icon';
import { SIconButton } from '../ui/s-icon-button';
import { SText } from '../ui/s-text';

interface Props {
  book: BookResponse;
  added: boolean;
  onToggle: () => void;
}

export function BookResultRow({ book, added, onToggle }: Props) {
  const { t } = useTranslation();
  const subtitle = [book.author, book.language].filter(Boolean).join(' · ');
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
      <View
        style={{
          width: 44,
          height: 62,
          borderRadius: radii.sm,
          backgroundColor: colors.surface,
          overflow: 'hidden',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        {book.coverUrl ? (
          <Image
            source={{ uri: book.coverUrl }}
            style={{ width: 44, height: 62 }}
            resizeMode="cover"
          />
        ) : (
          <SIcon name="menu_book" size={22} color={colors.textSecondary} />
        )}
      </View>
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
              {book.pages} p.
            </SText>
          )}
        </View>
      </View>
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
