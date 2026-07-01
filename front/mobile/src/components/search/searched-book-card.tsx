import React from 'react';
import { Image, StyleSheet, View } from 'react-native';
import { LibgenBook } from '../../models/book';
import { colors, hexToRgba } from '../../theme/colors';
import SText from '../shared/SText';
import SButton from '../shared/SButton';
import SIcon from '../icons/SIcon';

interface Props {
  book: LibgenBook;
  selected?: boolean;
  onSelect(): void;
  deleteMode?: boolean;
}

export default function SearchedBookCard({
  book,
  selected = false,
  onSelect,
  deleteMode = false,
}: Props) {
  return (
    <SButton
      disabled={deleteMode}
      onPress={onSelect}
      style={{
        boxShadow: selected ? '' : colors.boxShadow,
        borderRadius: 12,
        backgroundColor: selected
          ? hexToRgba(colors.primary_fixed, 0.3)
          : colors.surface_container_lowest,
        padding: 10,
        marginTop: 10,
        flexDirection: 'row',
        gap: 10,
        outlineColor: colors.primary_fixed,
        outlineWidth: selected ? 1 : 0,
      }}
    >
      <Image
        source={{
          uri: book.cover_url,
          headers: {
            Referer: 'https://libgen.bz/',
            'User-Agent': 'Mozilla/5.0',
          },
        }}
        resizeMode="cover"
        style={{ height: 100, borderRadius: 8, aspectRatio: 0.67 }}
        onError={(err) => console.log('Image load error:', book.cover_url, err.nativeEvent.error)}
      />
      <View style={{ flex: 1, overflow: 'hidden' }}>
        <SText
          ellipsizeMode="tail"
          style={{ flex: 1, fontFamily: 'semibold', fontSize: 16, alignSelf: 'flex-start' }}
          numberOfLines={2}
        >
          {book.title}
        </SText>

        <View style={{ justifyContent: 'space-between' }}>
          <SText
            ellipsizeMode="tail"
            numberOfLines={1}
            style={[
              {
                backgroundColor: colors.tertiary_fixed_dim,
              },
              styles.label,
            ]}
          >
            {book.author}
          </SText>

          <View style={{ flexDirection: 'row', marginTop: 8, justifyContent: 'space-between' }}>
            <View style={{ flexDirection: 'row', gap: 8 }}>
              <SText
                style={[
                  {
                    backgroundColor: colors.tertiary_fixed,
                  },
                  styles.label,
                ]}
              >
                {book.extension}
              </SText>

              {book.language && (
                <SText
                  style={[
                    {
                      backgroundColor: colors.secondary_fixed,
                    },
                    styles.label,
                  ]}
                >
                  {book.language}
                </SText>
              )}
            </View>

            {!deleteMode && (
              <SIcon
                color={selected ? colors.primary : hexToRgba(colors.primary_fixed, 0.7)}
                size={24}
                name={selected ? 'check_circle' : 'circle'}
                type="outlined"
              />
            )}

            {deleteMode && (
              <SButton
                onPress={onSelect}
                style={{ backgroundColor: colors.error_container, padding: 1, borderRadius: 5 }}
              >
                <SIcon color={colors.error} size={24} name={'delete'} type="outlined" />
              </SButton>
            )}
          </View>
        </View>
      </View>
    </SButton>
  );
}

const styles = StyleSheet.create({
  label: {
    borderRadius: 8,
    paddingHorizontal: 4,
    paddingVertical: 1,
    alignSelf: 'flex-start',
  },
});
