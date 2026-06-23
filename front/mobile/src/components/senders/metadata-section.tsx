import React, { useEffect, useState } from 'react';
import { StyleSheet, View } from 'react-native';
import { colors, hexToRgba } from '../../theme/colors';
import { BookMetadata } from '../../models/book-metadata';
import SText from '../shared/SText';
import { TextInput } from 'react-native-gesture-handler';

interface Props {
  initialMetadata?: BookMetadata;
  onChange(metadata: BookMetadata): void;
}

export default function MetadataSection({ initialMetadata, onChange }: Props) {
  const [metadata, setMetadata] = useState(initialMetadata ?? {});

  useEffect(() => {
    onChange(metadata)
  }, [metadata])

  return (
    <View style={{ boxShadow: colors.boxShadow, borderRadius: 12, padding: 15, gap: 8 }}>
      <View style={styles.section}>
        <SText style={styles.label}>Title</SText>
        <TextInput
          style={styles.textInput}
          onChangeText={(e) => setMetadata((s) => ({ ...s, title: e }))}
        />
      </View>

      <View style={styles.section}>
        <SText style={styles.label}>Author</SText>
        <TextInput
          style={styles.textInput}
          onChangeText={(e) => setMetadata((s) => ({ ...s, title: e }))}
        />
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  section: {
    gap: 5
  },
  label: {
    fontFamily: 'semibold'
  },
  textInput: {
    borderColor: hexToRgba(colors.outline_variant, 0.2),
    borderWidth: 1,
    borderRadius: 8,
    backgroundColor: colors.surface_container_low,
  },
});
