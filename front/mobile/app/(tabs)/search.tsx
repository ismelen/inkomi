import React, { useEffect, useState } from 'react';
import { View } from 'react-native';
import SText from '../../src/components/shared/SText';
import { colors, hexToRgba } from '../../src/theme/colors';
import { ScrollView, TextInput } from 'react-native-gesture-handler';
import SIcon from '../../src/components/icons/SIcon';
import { useLibgen } from '../../src/hooks/useLibgen';
import { useDebounce } from '../../src/hooks/useDebouncer';
import { LibgenBook } from '../../src/models/book';
import SearchedBookCard from '../../src/components/search/searched-book-card';
import { useShallow } from 'zustand/react/shallow';
import SButton from '../../src/components/shared/SButton';
import { router } from 'expo-router';
import SearchedBookCardSkeleton from '../../src/components/search/search-book-card-skeleton';

export default function Search() {
  const { search, selectBook, selectedBooks } = useLibgen(
    useShallow((s) => ({
      search: s.search,
      selectBook: s.selectBook,
      selectedBooks: s.selected,
    }))
  );

  const selectedCant = Object.entries(selectedBooks).length;

  const [query, setQuery] = useState('');
  const debouncedQuery = useDebounce<string>(query, 500);

  const [results, setResults] = useState<LibgenBook[]>([]);
  const [isSearching, setIsSearching] = useState(false);

  useEffect(() => {
    if (!debouncedQuery) {
      setResults([]);
      return;
    }
    setIsSearching(true);
    search(debouncedQuery)
      .then((res) => {
        if (!res) return;
        setResults(res);
      })
      .finally(() => setIsSearching(false));
  }, [debouncedQuery]);

  return (
    <View style={{ flex: 1 }}>
      <View style={{ paddingHorizontal: 24 }}>
        <SText style={{ fontFamily: 'bold', fontSize: 28 }}>Search</SText>

        <View
          style={{
            borderColor: hexToRgba(colors.outline_variant, 0.2),
            borderWidth: 1,
            borderRadius: 12,
            backgroundColor: colors.surface_container_low,
            flexDirection: 'row',
            alignItems: 'center',
            paddingHorizontal: 15,
            paddingVertical: 5,
            gap: 5,
          }}
        >
          <SIcon name="search" color={colors.on_surface_variant} size={26} type="outlined" />
          <TextInput
            placeholder="Search by title, author or genere..."
            onChangeText={setQuery}
            style={{
              fontSize: 16,
              flex: 1,
            }}
          />
        </View>
      </View>
      <ScrollView style={{ paddingHorizontal: 24, marginTop: 10 }} bounces>
        {isSearching
          ? Array.from({ length: 3 }).map((_, i) => <SearchedBookCardSkeleton key={i} />)
          : results.map((e) => (
              <SearchedBookCard
                key={e.md5}
                book={e}
                onSelect={() => selectBook(e)}
                selected={!!selectedBooks[e.md5]}
              />
            ))}
      </ScrollView>

      {selectedCant !== 0 && (
        <SButton
          onPress={() => router.navigate('/send-libgen')}
          style={{
            position: 'absolute',
            bottom: 20,
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
            Send books ({selectedCant})
          </SText>
          <SIcon name="upload_file" color={colors.on_primary} size={26} />
        </SButton>
      )}
    </View>
  );
}
