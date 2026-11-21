import React, { useCallback, useEffect, useRef, useState } from 'react';
import { ActivityIndicator, Alert, FlatList, ScrollView, View } from 'react-native';
import { useLocalSearchParams } from 'expo-router';
import { useTranslation } from 'react-i18next';
import { errorKey } from '../../src/api/api-error';
import { wishlistApi } from '../../src/api/wishlist-api';
import { WishSuggestion, WorksSort } from '../../src/api/types/wishlist';
import { BookCover } from '../../src/components/books/book-cover';
import { Header } from '../../src/components/ui/header';
import { Screen } from '../../src/components/ui/screen';
import { SIconButton } from '../../src/components/ui/s-icon-button';
import { SSelect } from '../../src/components/ui/s-select';
import { SText } from '../../src/components/ui/s-text';
import { BookLanguage, bookLanguageCodes } from '../../src/models/book-languages';
import { openLibrarySearch } from '../../src/services/library-search';
import { defaultSearchLanguage } from '../../src/stores/use-search';
import { useWishlist } from '../../src/stores/use-wishlist';
import { colors, spacing } from '../../src/theme';

const sorts: WorksSort[] = ['popular', 'new', 'old', 'title'];

type Status = 'loading' | 'done' | 'error';

/** The books of a wished author, from Open Library, sorted and paginated. */
export default function AuthorWorksScreen() {
  const { t } = useTranslation();
  const { name = '', authorKey = '' } = useLocalSearchParams<{
    name?: string;
    authorKey?: string;
  }>();
  const wishes = useWishlist((s) => s.items);
  const addWish = useWishlist((s) => s.add);

  const appLanguage = defaultSearchLanguage();
  const [sort, setSort] = useState<WorksSort>('popular');
  const [language, setLanguage] = useState<BookLanguage | null>(appLanguage);
  const [works, setWorks] = useState<WishSuggestion[]>([]);
  const [page, setPage] = useState(0);
  const [hasMore, setHasMore] = useState(false);
  const [total, setTotal] = useState(0);
  const [status, setStatus] = useState<Status>('loading');
  const [error, setError] = useState<string>();
  const [loadingMore, setLoadingMore] = useState(false);
  const controller = useRef<AbortController | null>(null);

  const fetchPage = useCallback(
    async (next: number) => {
      controller.current?.abort();
      const ctrl = new AbortController();
      controller.current = ctrl;
      if (next === 1) {
        setStatus('loading');
        setWorks([]);
      } else {
        setLoadingMore(true);
      }
      try {
        const result = await wishlistApi.authorWorks(
          { authorKey, author: name, sort, lang: language ?? undefined, page: next },
          ctrl.signal
        );
        if (ctrl.signal.aborted) return;
        setWorks((prev) => {
          const merged = next === 1 ? result.works : [...prev, ...result.works];
          return merged.filter((w, i) => merged.findIndex((o) => o.workKey === w.workKey) === i);
        });
        setPage(result.page);
        setHasMore(result.hasMore);
        setTotal(result.total);
        setStatus('done');
      } catch (err) {
        if (ctrl.signal.aborted) return;
        if (next === 1) {
          setStatus('error');
          setError(errorKey(err));
        }
      } finally {
        if (!ctrl.signal.aborted) setLoadingMore(false);
      }
    },
    [authorKey, name, sort, language]
  );

  useEffect(() => {
    void fetchPage(1);
    return () => controller.current?.abort();
  }, [fetchPage]);

  const loadMore = () => {
    if (status === 'done' && hasMore && !loadingMore) void fetchPage(page + 1);
  };

  const languages: (BookLanguage | null)[] = [
    ...(appLanguage ? [appLanguage] : []),
    ...bookLanguageCodes
      .filter((c) => c !== appLanguage)
      .sort((a, b) => t(`bookLanguages.${a}`).localeCompare(t(`bookLanguages.${b}`))),
    null,
  ];

  const isWished = (w: WishSuggestion) =>
    wishes.some(
      (wish) =>
        wish.kind === 'book' &&
        ((!!w.workKey && wish.workKey === w.workKey) ||
          wish.title.toLowerCase() === w.title.toLowerCase())
    );

  const wish = (w: WishSuggestion) =>
    void addWish({ kind: 'book', title: w.title, author: w.author, workKey: w.workKey }).then(
      (err) => err && Alert.alert(t('errors.title'), t(err))
    );

  return (
    <Screen bottomInset>
      <Header title={name || t('authorWorks.title')} />
      <ScrollView
        horizontal
        showsHorizontalScrollIndicator={false}
        style={{ flexGrow: 0 }}
        contentContainerStyle={{
          gap: spacing.sm,
          paddingHorizontal: spacing.lg,
          paddingBottom: spacing.sm,
          alignItems: 'center',
        }}
      >
        <SSelect
          label={t('authorWorks.sort')}
          value={sort}
          onChange={setSort}
          options={sorts.map((s) => ({ value: s, label: t(`authorWorks.sorts.${s}`) }))}
        />
        <SSelect
          label={t('search.language')}
          value={language}
          onChange={setLanguage}
          options={languages.map((l) => ({
            value: l,
            label: l ? t(`bookLanguages.${l}`) : t('search.allLanguages'),
          }))}
        />
        {status === 'done' && total > 0 && (
          <SText variant="small" tone="textSecondary">
            {t('authorWorks.count', { count: total })}
          </SText>
        )}
      </ScrollView>

      <FlatList
        data={works}
        keyExtractor={(w, i) => w.workKey || `${w.title}-${i}`}
        onEndReached={loadMore}
        onEndReachedThreshold={0.5}
        contentContainerStyle={{
          paddingHorizontal: spacing.lg,
          paddingVertical: spacing.sm,
          flexGrow: 1,
        }}
        renderItem={({ item }) => {
          const wished = isWished(item);
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
                url={item.coverUrl}
                accessibilityLabel={t('search.cover', { title: item.title })}
              />
              <View style={{ flex: 1, gap: spacing.xxs }}>
                <SText variant="label" numberOfLines={2}>
                  {item.title}
                </SText>
                {!!item.year && (
                  <SText variant="small" tone="textSecondary">
                    {item.year}
                  </SText>
                )}
              </View>
              <SIconButton
                icon={wished ? 'check' : 'favorite'}
                label={t('wishlist.wish')}
                size={20}
                disabled={wished}
                onPress={() => wish(item)}
              />
              <SIconButton
                icon="search"
                label={t('wishlist.searchLibrary')}
                size={20}
                onPress={() => openLibrarySearch(item.title, 'title')}
              />
            </View>
          );
        }}
        ListFooterComponent={
          loadingMore ? (
            <ActivityIndicator color={colors.primary} style={{ padding: spacing.lg }} />
          ) : null
        }
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
                {status === 'error' && error ? t(error) : t('authorWorks.empty')}
              </SText>
            )}
          </View>
        }
      />
    </Screen>
  );
}
