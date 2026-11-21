import React, { useEffect, useMemo, useState } from 'react';
import { ActivityIndicator, Alert, FlatList, Pressable, View } from 'react-native';
import { router } from 'expo-router';
import { useTranslation } from 'react-i18next';
import { AddWishRequest, WishItem, WishKind } from '../../src/api/types/wishlist';
import { AddWishModal } from '../../src/components/wishlist/add-wish-modal';
import { Header } from '../../src/components/ui/header';
import { Screen } from '../../src/components/ui/screen';
import { SIcon } from '../../src/components/ui/s-icon';
import { SIconButton } from '../../src/components/ui/s-icon-button';
import { SSegmented } from '../../src/components/ui/s-segmented';
import { SText } from '../../src/components/ui/s-text';
import { openLibrarySearch } from '../../src/services/library-search';
import { useWishlist } from '../../src/stores/use-wishlist';
import { colors, radii, spacing } from '../../src/theme';

function openAuthorWorks(item: WishItem) {
  router.push({
    pathname: '/author-works',
    params: { name: item.title, authorKey: item.authorKey },
  });
}

export default function WishlistScreen() {
  const { t } = useTranslation();
  const { items, status, error, load, add, remove } = useWishlist();
  const [kind, setKind] = useState<WishKind>('book');
  const [adding, setAdding] = useState(false);

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
    setKind(item.kind);
    setAdding(false);
  };

  return (
    <Screen bottomInset>
      <Header
        title={t('wishlist.title')}
        right={<SIconButton icon="add" label={t('wishlist.add')} onPress={() => setAdding(true)} />}
      />
      <View style={{ paddingHorizontal: spacing.lg, paddingBottom: spacing.xs }}>
        <SSegmented
          size="lg"
          value={kind}
          onChange={setKind}
          options={[
            { value: 'book', label: t('wishlist.books') },
            { value: 'author', label: t('wishlist.authors') },
          ]}
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
        renderItem={({ item }) => {
          const isAuthor = item.kind === 'author';
          return (
            <View style={{ flexDirection: 'row', alignItems: 'center', gap: spacing.sm }}>
              <Pressable
                disabled={!isAuthor}
                onPress={() => openAuthorWorks(item)}
                accessibilityRole={isAuthor ? 'button' : undefined}
                accessibilityHint={isAuthor ? t('authorWorks.open') : undefined}
                android_ripple={isAuthor ? { color: colors.ripple } : undefined}
                style={{
                  flex: 1,
                  flexDirection: 'row',
                  alignItems: 'center',
                  gap: spacing.xs,
                  paddingVertical: spacing.sm,
                  borderRadius: radii.md,
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
                {isAuthor && <SIcon name="chevron_right" size={20} color={colors.textSecondary} />}
              </Pressable>
              <SIconButton
                icon="search"
                label={t('wishlist.searchLibrary')}
                size={20}
                onPress={() => openLibrarySearch(item.title, isAuthor ? 'author' : 'title')}
              />
              <SIconButton
                icon="delete"
                label={t('wishlist.remove')}
                size={20}
                onPress={() => void remove(item.id).then(showError)}
              />
            </View>
          );
        }}
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
      <AddWishModal
        visible={adding}
        initialKind={kind}
        onClose={() => setAdding(false)}
        onAdd={onAdd}
      />
    </Screen>
  );
}
