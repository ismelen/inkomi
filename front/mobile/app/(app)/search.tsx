import React, { useEffect, useRef, useState } from 'react';
import { ActivityIndicator, FlatList, View } from 'react-native';
import { router } from 'expo-router';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useTranslation } from 'react-i18next';
import { BookFormat } from '../../src/api/types/books';
import { BookResultRow } from '../../src/components/books/book-result-row';
import { Screen } from '../../src/components/ui/screen';
import { SButton } from '../../src/components/ui/s-button';
import { SChip } from '../../src/components/ui/s-chip';
import { SIcon } from '../../src/components/ui/s-icon';
import { SIconButton } from '../../src/components/ui/s-icon-button';
import { SText } from '../../src/components/ui/s-text';
import { STextInput } from '../../src/components/ui/s-text-input';
import { supportedFormats } from '../../src/models/list-item';
import { isLibraryAdded, useList } from '../../src/stores/use-list';
import { useSearch } from '../../src/stores/use-search';
import { colors, spacing } from '../../src/theme';

const debounceMs = 500;

export default function SearchScreen() {
  const { t } = useTranslation();
  const insets = useSafeAreaInsets();
  const { query, format, results, status, error, setQuery, setFormat, search, reset } = useSearch();
  const items = useList((s) => s.items);
  const toggleLibrary = useList((s) => s.toggleLibrary);
  const [added, setAdded] = useState(0);
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => () => reset(), [reset]);

  const onChangeQuery = (text: string) => {
    setQuery(text);
    if (timer.current) clearTimeout(timer.current);
    timer.current = setTimeout(() => void search(), debounceMs);
  };

  const formats: (BookFormat | null)[] = [null, ...supportedFormats];

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
            onSubmitEditing={() => void search()}
            returnKeyType="search"
            placeholder={t('search.placeholder')}
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
      <View style={{ flexDirection: 'row', gap: spacing.sm, paddingHorizontal: spacing.lg }}>
        {formats.map((f) => (
          <SChip
            key={f ?? 'all'}
            label={f ? f.toUpperCase() : t('search.all')}
            selected={format === f}
            onPress={() => setFormat(f)}
          />
        ))}
      </View>

      <FlatList
        data={status === 'done' ? results : []}
        keyExtractor={(book, i) => book.md5 || `${book.title}-${i}`}
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
