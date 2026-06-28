import React from 'react';
import { Upload } from '../../models/upload';
import { ActivityIndicator, StyleSheet, View } from 'react-native';
import { colors, hexToRgba } from '../../theme/colors';
import SText from '../shared/SText';
import SButton from '../shared/SButton';
import SIcon from '../icons/SIcon';

interface Props {
  data: Upload;
  onRetry(): void;
}

export default function UploadCard({ data, onRetry }: Props) {
  const sources =
    data.request.mode === 'files' ? data.request.sources : (data.request.sources[0].children ?? []);

  return (
    <SButton
      disabled={data.error === undefined}
      style={{
        backgroundColor: data.error
          ? hexToRgba(colors.error_container, 0.2)
          : colors.surface_container_lowest,
        borderRadius: 12,
        boxShadow: colors.boxShadow,
        padding: 10,
        borderColor: colors.error_container,
        borderWidth: data.error ? 1 : 0,
      }}
    >
      <SText style={{ flexShrink: 1, fontSize: 16 }} numberOfLines={2} ellipsizeMode="tail">
        {sources
          .slice(0, 3)
          .map((e) => e.name)
          .join(', ')}
      </SText>
      <View style={{ flexDirection: 'row', alignItems: 'flex-end', marginTop: 3, gap: 10 }}>
        <View style={{ flex: 1 }}>
          <SText style={styles.timestamp}>{new Date(data.timestamp).toLocaleString()}</SText>
          {data.error && <SText style={styles.error}>{data.error?.message}</SText>}
        </View>
        {data.error ? (
          <View style={styles.retry}>
            <SIcon name="sync" color={colors.on_primary} size={24} type="outlined" />
            <SText style={{ color: colors.on_primary, fontFamily: 'semibold' }}>Retry</SText>
          </View>
        ) : (
          <ActivityIndicator color={colors.primary} />
        )}
      </View>
    </SButton>
  );
}

const styles = StyleSheet.create({
  retry: {
    backgroundColor: colors.primary_container,
    paddingVertical: 5,
    paddingLeft: 5,
    paddingRight: 10,
    borderRadius: 6,
    flexDirection: 'row',
    gap: 3,
    alignItems: 'center',
  },
  timestamp: {
    fontSize: 12,
    color: hexToRgba(colors.on_surface_variant, 0.5),
    marginBottom: 4,
  },
  error: {
    backgroundColor: colors.error_container,
    color: colors.error,
    paddingVertical: 2,
    paddingHorizontal: 5,
    borderRadius: 5,
    flexShrink: 1,
    alignSelf: 'flex-start',
    fontSize: 12,
  },
});
