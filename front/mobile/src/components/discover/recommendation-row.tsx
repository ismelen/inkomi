import React from 'react';
import { View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { Recommendation } from '../../api/types/recommendations';
import { spacing } from '../../theme';
import { BookCover } from '../books/book-cover';
import { SIconButton } from '../ui/s-icon-button';
import { SText } from '../ui/s-text';

interface Props {
  rec: Recommendation;
  wished: boolean;
  onSearch: () => void;
  onWish: () => void;
  onDismiss: () => void;
}

export function RecommendationRow({ rec, wished, onSearch, onWish, onDismiss }: Props) {
  const { t } = useTranslation();
  const subtitle = [rec.authors.join(', '), rec.year || ''].filter(Boolean).join(' · ');

  return (
    <View style={{ flexDirection: 'row', gap: spacing.md, paddingVertical: spacing.sm }}>
      <BookCover url={rec.coverUrl} />
      <View style={{ flex: 1, gap: spacing.xxs }}>
        <SText variant="label" numberOfLines={2}>
          {rec.title}
        </SText>
        {!!subtitle && (
          <SText variant="small" tone="textSecondary" numberOfLines={1}>
            {subtitle}
          </SText>
        )}
        {rec.reason && (
          <SText variant="small" tone="textSecondary" numberOfLines={1}>
            {t('discover.because', { title: rec.reason.title })}
          </SText>
        )}
        <View style={{ flexDirection: 'row', marginLeft: -spacing.sm }}>
          <SIconButton icon="search" label={t('discover.search')} size={20} onPress={onSearch} />
          <SIconButton
            icon={wished ? 'check' : 'favorite'}
            label={t('discover.wish')}
            size={20}
            disabled={wished}
            onPress={onWish}
          />
          <SIconButton icon="close" label={t('discover.dismiss')} size={20} onPress={onDismiss} />
        </View>
      </View>
    </View>
  );
}
