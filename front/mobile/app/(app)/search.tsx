import React, { useEffect, useRef, useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  FlatList,
  Keyboard,
  Pressable,
  ScrollView,
  View,
} from 'react-native';
import { router } from 'expo-router';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useTranslation } from 'react-i18next';
import { BookFormat, SearchField } from '../../src/api/types/books';
import { WishSuggestion } from '../../src/api/types/wishlist';
import { BookResultRow } from '../../src/components/books/book-result-row';
import { Screen } from '../../src/components/ui/screen';
import { SButton } from '../../src/components/ui/s-button';
import { SIcon } from '../../src/components/ui/s-icon';
import { SIconButton } from '../../src/components/ui/s-icon-button';
import { SSelect } from '../../src/components/ui/s-select';
import { SText } from '../../src/components/ui/s-text';
import { STextInput } from '../../src/components/ui/s-text-input';
import { useCatalogSuggestions } from '../../src/hooks/use-catalog-suggestions';
import { BookLanguage, bookLanguageCodes } from '../../src/models/book-languages';
import { supportedFormats } from '../../src/models/list-item';
import { isLibraryAdded, useList } from '../../src/stores/use-list';
import { defaultSearchLanguage, useSearch } from '../../src/stores/use-search';
import { useWishlist } from '../../src/stores/use-wishlist';
import { colors, radii, spacing } from '../../src/theme';

const debounceMs = 500;
const maxSuggestions = 5;

export default function SearchScreen() {
  const { t } = useTranslation();
  const insets = useSafeAreaInsets();
  const {
    query,
    field,
    format,
    language,
    results,
    nextPage,
    loadingMore,
    status,
    error,
    setQuery,
    setField,
    setFormat,
    setLanguage,
    search,
    loadMore,
    reset,
  } = useSearch();
  const { suggestions, lookup, clear } = useCatalogSuggestions();
  const [focused, setFocused] = useState(false);
  const items = useList((s) => s.items);
  const toggleLibrary = useList((s) => s.toggleLibrary);
  const wishes = useWishlist((s) => s.items);
  const addWish = useWishlist((s) => s.add);
  const [added, setAdded] = useState(0);
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => () => reset(), [reset]);
  useEffect(() => {
    if (useWishlist.getState().status === 'idle') void useWishlist.getState().load();
  }, []);

  const onChangeQuery = (text: string) => {
    setQuery(text);
    lookup(field === 'author' ? 'author' : 'book', text);
    if (timer.current) clearTimeout(timer.current);
    timer.current = setTimeout(() => void search(), debounceMs);
  };

  const onChangeField = (next: SearchField) => {
    clear();
    setField(next);
  };

  // the catalog spelling of a name finds what a typo misses
  const pickSuggestion = (s: WishSuggestion) => {
    if (timer.current) clearTimeout(timer.current);
    clear();
    Keyboard.dismiss();
    setQuery(s.title);
    void search();
  };

  const shownSuggestions = focused && query.trim() ? suggestions.slice(0, maxSuggestions) : [];

  const formats: (BookFormat | null)[] = [null, ...supportedFormats];
  const fields: SearchField[] = ['title', 'author'];
  const appLanguage = defaultSearchLanguage();
  // the app language first, the rest by their translated name
  const languages: (BookLanguage | null)[] = [
    ...(appLanguage ? [appLanguage] : []),
    ...bookLanguageCodes
      .filter((c) => c !== appLanguage)
      .sort((a, b) => t(`bookLanguages.${a}`).localeCompare(t(`bookLanguages.${b}`))),
    null,
  ];

  return (
    <Screen>
      <View
        style={{ flexDirection: 'row', alignItems: 'center', gap: spacing.xs, padding: spacing.sm }}
      >
        <SIconButton icon="chevron_left" label={t('common.back')} onPress={() => router.back()} />
        <View style={{ flex: 1 }}>
          <STextInput
            autoFocus
            value={query}
            onChangeText={onChangeQuery}
            onSubmitEditing={() => {
              clear();
              void search();
            }}
            onFocus={() => setFocused(true)}
            onBlur={() => setFocused(false)}
            returnKeyType="search"
            placeholder={t(
              field === 'author' ? 'search.placeholderAuthor' : 'search.placeholderTitle'
            )}
            left={<SIcon name="search" size={20} color={colors.textSecondary} />}
            right={
              query ? (
                <SIconButton
                  icon="close"
                  label={t('common.close')}
                  size={18}
                  onPress={() => onChangeQuery('')}
                />
              ) : undefined
            }
          />
        </View>
      </View>
      <ScrollView
        horizontal
        showsHorizontalScrollIndicator={false}
        keyboardShouldPersistTaps="handled"
        style={{ flexGrow: 0 }}
        contentContainerStyle={{
          gap: spacing.sm,
          paddingHorizontal: spacing.lg,
          paddingBottom: spacing.sm,
        }}
      >
        <SSelect
          label={t('search.by')}
          value={field}
          onChange={onChangeField}
          options={fields.map((f) => ({
            value: f,
            label: t(f === 'author' ? 'search.byAuthor' : 'search.byTitle'),
          }))}
        />
        <SSelect
          label={t('search.language')}
          value={language === undefined ? appLanguage : language}
          onChange={setLanguage}
          options={languages.map((l) => ({
            value: l,
            label: l ? t(`bookLanguages.${l}`) : t('search.allLanguages'),
          }))}
        />
        <SSelect
          label={t('search.format')}
          value={format}
          onChange={setFormat}
          options={formats.map((f) => ({
            value: f,
            label: f ? f.toUpperCase() : t('search.all'),
          }))}
        />
      </ScrollView>

      <View style={{ flex: 1 }}>
        <FlatList
          data={status === 'done' ? results : []}
          onEndReached={() => nextPage > 0 && void loadMore()}
          onEndReachedThreshold={0.5}
          ListFooterComponent={
            loadingMore ? (
              <ActivityIndicator color={colors.primary} style={{ padding: spacing.lg }} />
            ) : null
          }
          keyExtractor={(book, i) => book.md5 || `${book.title}-${i}`}
          keyboardShouldPersistTaps="handled"
          contentContainerStyle={{
            paddingHorizontal: spacing.lg,
            paddingVertical: spacing.sm,
            flexGrow: 1,
          }}
          renderItem={({ item: book }) => (
            <BookResultRow
              book={book}
              added={isLibraryAdded(items, book.md5)}
              onToggle={() => setAdded((n) => n + (toggleLibrary(book) ? 1 : -1))}
              wished={wishes.some((w) => w.kind === 'book' && w.title === book.title)}
              onWish={() =>
                void addWish({ kind: 'book', title: book.title, author: book.author }).then(
                  (err) => err && Alert.alert(t('errors.title'), t(err))
                )
              }
            />
          )}
          ListEmptyComponent={
            <View
              style={{
                flex: 1,
                alignItems: 'center',
                justifyContent: 'center',
                padding: spacing.xl,
              }}
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
                    : status === 'done'
                      ? t('search.empty')
                      : t('search.idle')}
                </SText>
              )}
            </View>
          }
        />
        {shownSuggestions.length > 0 && (
          <View
            style={{
              position: 'absolute',
              top: 0,
              left: spacing.lg,
              right: spacing.lg,
              paddingVertical: spacing.xs,
              borderRadius: radii.md,
              borderWidth: 1,
              borderColor: colors.border,
              backgroundColor: colors.background,
              elevation: 4,
            }}
          >
            {shownSuggestions.map((s, i) => (
              <Pressable
                key={`${s.authorKey || s.workKey}-${i}`}
                onPress={() => pickSuggestion(s)}
                android_ripple={{ color: colors.ripple }}
                accessibilityRole="button"
                style={{
                  flexDirection: 'row',
                  alignItems: 'center',
                  gap: spacing.sm,
                  paddingHorizontal: spacing.md,
                  paddingVertical: spacing.sm,
                }}
              >
                <SIcon name="search" size={18} color={colors.textSecondary} />
                <View style={{ flex: 1 }}>
                  <SText variant="label" numberOfLines={1}>
                    {s.title}
                  </SText>
                  {!!s.author && (
                    <SText variant="small" tone="textSecondary" numberOfLines={1}>
                      {s.author}
                    </SText>
                  )}
                </View>
              </Pressable>
            ))}
          </View>
        )}
      </View>

      {added > 0 && (
        <View
          style={{
            flexDirection: 'row',
            alignItems: 'center',
            gap: spacing.md,
            paddingHorizontal: spacing.lg,
            paddingTop: spacing.md,
            paddingBottom: spacing.md + insets.bottom,
            borderTopWidth: 1,
            borderTopColor: colors.surface,
          }}
        >
          <SText variant="label" style={{ flex: 1 }}>
            {t('search.added', { count: added })}
          </SText>
          <SButton label={t('common.done')} onPress={() => router.back()} compact />
        </View>
      )}
    </Screen>
  );
}
