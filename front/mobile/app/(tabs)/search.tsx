import React, { useState, useEffect } from 'react';
import { View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import SText from '../../src/components/shared/SText';
import { colors, hexToRgba } from '../../src/theme/colors';
import { ScrollView } from 'react-native-gesture-handler';
import SIcon from '../../src/components/icons/SIcon';
import { useLibgen } from '../../src/hooks/useLibgen';
import SearchedBookCard from '../../src/components/search/searched-book-card';
import { useShallow } from 'zustand/react/shallow';
import SButton from '../../src/components/shared/SButton';
import { router } from 'expo-router';
import SearchedBookCardSkeleton from '../../src/components/search/search-book-card-skeleton';
import STextInput from '../../src/components/shared/STextInput';
import { TransactionSource } from '../../src/models/transaction-source';
import { useTranslation } from 'react-i18next';
import { BACKEND_API_URL } from '../../src/constants';

export default function Search() {
  const insets = useSafeAreaInsets();
  const { search, selectBook, selectedBooks } = useLibgen(
    useShallow((s) => ({
      search: s.search,
      selectBook: s.selectBook,
      selectedBooks: s.selected,
    }))
  );

  const selectedCant = Object.entries(selectedBooks).length;
  const { t } = useTranslation();

  const [query, setQuery] = useState('');

  const [results, setResults] = useState<TransactionSource[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [serviceActive, setServiceActive] = useState<boolean | null>(null);

  useEffect(() => {
    const checkMirror = async () => {
      try {
        const res = await fetch(`${BACKEND_API_URL}/books/check-mirror`);
        const data = await res.json();
        setServiceActive(data.active);
      } catch (e) {
        setServiceActive(false);
      }
    };
    checkMirror();
  }, []);

  const handleSearch = async () => {
    setIsSearching(true);
    const res = await search(query);
    setIsSearching(false);
    setResults((s) => res ?? s);
  };

  return (
    <View style={{ flex: 1 }}>
      <View style={{ paddingHorizontal: 24 }}>
        <View style={{ flexDirection: 'row', alignItems: 'center', gap: 10 }}>
          <SText style={{ fontFamily: 'bold', fontSize: 28 }}>{t('search.title')}</SText>
          {serviceActive === false && (
            <View
              style={{
                flexDirection: 'row',
                alignItems: 'center',
                gap: 4,
                backgroundColor: colors.error_container,
                paddingHorizontal: 10,
                paddingVertical: 4,
                borderRadius: 12,
              }}
            >
              <SIcon name="info" color={colors.error} size={14} type="outlined" />
              <SText style={{ color: colors.error, fontSize: 10, fontFamily: 'semibold' }}>
                {t('search.serviceDown')}
              </SText>
            </View>
          )}
        </View>

        <View style={{ flexDirection: 'row', gap: 8, marginTop: 10 }}>
          <STextInput
            placeholder={t('search.placeholder')}
            onChangeText={setQuery}
            onSubmitEditing={handleSearch}
            style={{
              fontSize: 16,
              flex: 1,
              borderColor: hexToRgba(colors.outline_variant, 0.2),
              borderWidth: 1,
              borderRadius: 12,
              backgroundColor: colors.surface_container_low,
            }}
          />
          <SButton
            onPress={handleSearch}
            style={{ backgroundColor: colors.primary_container, borderRadius: 12, padding: 10 }}
          >
            <SIcon name="search" color={colors.on_primary_container} size={26} type="outlined" />
          </SButton>
        </View>
      </View>
      <ScrollView
        style={{ paddingHorizontal: 24, marginTop: 10 }}
        contentContainerStyle={{ paddingBottom: 80 + insets.bottom }}
        bounces
      >
        {isSearching
          ? Array.from({ length: 3 }).map((_, i) => <SearchedBookCardSkeleton key={i} />)
          : results.map((e) => (
              <SearchedBookCard
                key={e.src}
                book={e}
                onSelect={() => selectBook(e)}
                selected={!!selectedBooks[e.src]}
              />
            ))}
      </ScrollView>

      {selectedCant !== 0 && (
        <SButton
          onPress={() => router.navigate('/send-libgen')}
          style={{
            position: 'absolute',
            bottom: 20 + insets.bottom,
            right: 20,
            left: 20,
            backgroundColor: colors.primary_container,
            borderRadius: 12,
            paddingVertical: 15,
            paddingHorizontal: 20,
            justifyContent: 'center',
            alignItems: 'center',
            flexDirection: 'row',
            gap: 10,
          }}
        >
          <SText style={{ color: colors.on_primary, fontFamily: 'semibold', fontSize: 16 }}>
            {t('search.sendBooks', { count: selectedCant })}
          </SText>
          <SIcon name="upload_file" color={colors.on_primary} size={26} />
        </SButton>
      )}
    </View>
  );
}
