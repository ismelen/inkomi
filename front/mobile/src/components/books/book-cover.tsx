import React, { useEffect, useState } from 'react';
import { Image, View } from 'react-native';
import { colors, radii } from '../../theme';
import { backendUrl } from '../../config';
import { SIcon } from '../ui/s-icon';

interface Props {
  /** Absolute url, or a path of the api (/books/cover?url=...). */
  url?: string;
  width?: number;
  height?: number;
  accessibilityLabel?: string;
}

/** Cover of a book, or a placeholder without one or when it doesn't load. */
export function BookCover({ url, width = 44, height = 62, accessibilityLabel }: Props) {
  const [failed, setFailed] = useState(false);
  useEffect(() => setFailed(false), [url]);
  const uri = url?.startsWith('/') ? backendUrl + url : url;

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
      <SIcon name="menu_book" size={22} color={colors.textSecondary} />
      {uri && !failed && (
        <Image
          source={{ uri }}
          accessibilityLabel={accessibilityLabel}
          onError={() => setFailed(true)}
          style={{ position: 'absolute', width, height }}
          resizeMode="cover"
        />
      )}
    </View>
  );
}
