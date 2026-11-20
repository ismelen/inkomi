import React, { useEffect, useMemo, useState } from 'react';
import { ActivityIndicator, Alert, FlatList, View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { AddWishRequest, WishKind } from '../../src/api/types/wishlist';
import { AddWishModal } from '../../src/components/wishlist/add-wish-modal';
import { Header } from '../../src/components/ui/header';
import { Screen } from '../../src/components/ui/screen';
import { SChip } from '../../src/components/ui/s-chip';
import { SIconButton } from '../../src/components/ui/s-icon-button';
import { SText } from '../../src/components/ui/s-text';
import { openLibrarySearch } from '../../src/services/library-search';
import { useWishlist } from '../../src/stores/use-wishlist';
import { colors, spacing } from '../../src/theme';

export default function WishlistScreen() {
  const { t } = useTranslation();
  const { items, status, error, load, add, remove } = useWishlist();
  const [kind, setKind] = useState<WishKind>('book');
  const [adding, setAdding] = useState<WishKind | null>(null);

  useEffect(() => {
    void load();
  }, [load]);

  const shown = useMemo(() => items.filter((item) => item.kind === kind), [items, kind]);

  const showError = (key?: string) => key && Alert.alert(t('errors.title'), t(key));

  const onAdd = async (item: AddWishRequest) => {
    const err = await add(item);
    if (err) {
      showError(err);
      return;
    }
    setAdding(null);
  };

  return (
    <Screen bottomInset>
      <Header
        title={t('wishlist.title')}
        right={<SIconButton icon="add" label={t('wishlist.add')} onPress={() => setAdding(kind)} />}
      />
      <View style={{ flexDirection: 'row', gap: spacing.sm, paddingHorizontal: spacing.lg }}>
        <SChip
          label={t('wishlist.books')}
          selected={kind === 'book'}
          onPress={() => setKind('book')}
        />
        <SChip
          label={t('wishlist.authors')}
          selected={kind === 'author'}
          onPress={() => setKind('author')}
        />
      </View>

      <FlatList
        data={shown}
        keyExtractor={(item) => String(item.id)}
        contentContainerStyle={{
          paddingHorizontal: spacing.lg,
          paddingVertical: spacing.sm,
          flexGrow: 1,
        }}
        renderItem={({ item }) => (
          <View
            style={{
              flexDirection: 'row',
              alignItems: 'center',
              gap: spacing.sm,
              paddingVertical: spacing.sm,
            }}
          >
            <View style={{ flex: 1, gap: spacing.xxs }}>
              <SText variant="label" numberOfLines={2}>
                {item.title}
              </SText>
              {!!item.author && (
                <SText variant="small" tone="textSecondary" numberOfLines={1}>
                  {item.author}
                </SText>
              )}
            </View>
            <SIconButton
              icon="search"
              label={t('wishlist.searchLibrary')}
              size={20}
              onPress={() =>
                openLibrarySearch(item.title, item.kind === 'author' ? 'author' : 'title')
              }
            />
            <SIconButton
              icon="delete"
              label={t('wishlist.remove')}
              size={20}
              onPress={() => void remove(item.id).then(showError)}
            />
          </View>
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
                {status === 'error' && error
                  ? t(error)
                  : t(kind === 'author' ? 'wishlist.emptyAuthors' : 'wishlist.emptyBooks')}
              </SText>
            )}
          </View>
        }
      />
      <AddWishModal kind={adding} onClose={() => setAdding(null)} onAdd={onAdd} />
    </Screen>
  );
}
