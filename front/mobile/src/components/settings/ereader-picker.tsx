import React from 'react';
import { FlatList, Pressable } from 'react-native';
import { colors, radii, spacing } from '../../theme';
import { ereaders } from '../../models/ereaders';
import { SIcon } from '../ui/s-icon';
import { SSheet } from '../ui/s-sheet';
import { SText } from '../ui/s-text';

interface Props {
  visible: boolean;
  title: string;
  value: string;
  onChange: (key: string) => void;
  onClose: () => void;
}

export function EReaderPicker({ visible, title, value, onChange, onClose }: Props) {
  return (
    <SSheet visible={visible} title={title} onClose={onClose}>
      <FlatList
        data={ereaders}
        keyExtractor={(e) => e.key}
        initialScrollIndex={Math.max(
          0,
          ereaders.findIndex((e) => e.key === value)
        )}
        getItemLayout={(_, index) => ({ length: 48, offset: 48 * index, index })}
        renderItem={({ item }) => {
          const selected = item.key === value;
          return (
            <Pressable
              onPress={() => {
                onChange(item.key);
                onClose();
              }}
              style={{
                height: 48,
                flexDirection: 'row',
                alignItems: 'center',
                paddingHorizontal: spacing.md,
                borderRadius: radii.md,
                backgroundColor: selected ? colors.surface : 'transparent',
              }}
            >
              <SText variant="body" style={{ flex: 1 }}>
                {item.name}
              </SText>
              <SText variant="small" tone="textSecondary" style={{ marginRight: spacing.sm }}>
                {item.width} × {item.height}
              </SText>
              {selected && <SIcon name="check" size={18} />}
            </Pressable>
          );
        }}
      />
    </SSheet>
  );
}
