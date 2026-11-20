import React, { useEffect } from 'react';
import { ActivityIndicator, Alert, FlatList, View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { Recommendation, RecommendationSection } from '../../src/api/types/recommendations';
import { RecommendationRow } from '../../src/components/discover/recommendation-row';
import { Header } from '../../src/components/ui/header';
import { Screen } from '../../src/components/ui/screen';
import { SChip } from '../../src/components/ui/s-chip';
import { SText } from '../../src/components/ui/s-text';
import { openLibrarySearch } from '../../src/services/library-search';
import { useRecommendations } from '../../src/stores/use-recommendations';
import { useWishlist } from '../../src/stores/use-wishlist';
import { colors, spacing } from '../../src/theme';

const sections: { key: RecommendationSection; label: string }[] = [
  { key: 'for_you', label: 'discover.forYou' },
  { key: 'new', label: 'discover.new' },
  { key: 'discover', label: 'discover.discover' },
];

export default function DiscoverScreen() {
  const { t } = useTranslation();
  const { section, bySection, status, error, setSection, load, dismiss } = useRecommendations();
  const wishes = useWishlist((s) => s.items);
  const addWish = useWishlist((s) => s.add);

  useEffect(() => {
    void load(true);
    if (useWishlist.getState().status === 'idle') void useWishlist.getState().load();
  }, [load]);

  const recs = bySection[section] ?? [];
  const wished = (rec: Recommendation) => wishes.some((w) => w.workKey === rec.workKey);

  const showError = (key?: string) => key && Alert.alert(t('errors.title'), t(key));

  return (
    <Screen bottomInset>
      <Header title={t('discover.title')} />
      <View style={{ flexDirection: 'row', gap: spacing.sm, paddingHorizontal: spacing.lg }}>
        {sections.map((s) => (
          <SChip
            key={s.key}
            label={t(s.label)}
            selected={section === s.key}
            onPress={() => setSection(s.key)}
          />
        ))}
      </View>

      <FlatList
        data={status === 'loading' && recs.length === 0 ? [] : recs}
        keyExtractor={(rec) => rec.workKey}
        refreshing={status === 'loading' && recs.length > 0}
        onRefresh={() => void load(true)}
        contentContainerStyle={{
          paddingHorizontal: spacing.lg,
          paddingVertical: spacing.sm,
          flexGrow: 1,
        }}
        renderItem={({ item }) => (
          <RecommendationRow
            rec={item}
            wished={wished(item)}
            onSearch={() => openLibrarySearch(item.title)}
            onWish={() =>
              void addWish({
                kind: 'book',
                title: item.title,
                author: item.authors.join(', '),
                workKey: item.workKey,
              }).then(showError)
            }
            onDismiss={() => void dismiss(item.workKey).then(showError)}
          />
        )}
        ListEmptyComponent={
          <View
            style={{ flex: 1, alignItems: 'center', justifyContent: 'center', padding: spacing.xl }}
          >
            {status === 'loading' ? (
              <ActivityIndicator color={colors.primary} />
            ) : (
              <SText
                tone={status === 'error' ? 'danger' : 'textSecondary'}
                style={{ textAlign: 'center' }}
              >
                {status === 'error' && error ? t(error) : t('discover.empty')}
              </SText>
            )}
          </View>
        }
      />
    </Screen>
  );
}
