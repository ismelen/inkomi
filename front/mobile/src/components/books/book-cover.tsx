import React from 'react';
import { Image, View } from 'react-native';
import { colors, radii } from '../../theme';
import { SIcon } from '../ui/s-icon';

interface Props {
  url?: string;
  width?: number;
  height?: number;
}

/** Cover of a book, or a placeholder without one. */
export function BookCover({ url, width = 44, height = 62 }: Props) {
  return (
    <View
      style={{
        width,
        height,
        borderRadius: radii.sm,
        backgroundColor: colors.surface,
        overflow: 'hidden',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      {url ? (
        <Image source={{ uri: url }} style={{ width, height }} resizeMode="cover" />
      ) : (
        <SIcon name="menu_book" size={22} color={colors.textSecondary} />
      )}
    </View>
  );
}
