import { useState } from 'react';
import { Pressable, View } from 'react-native';
import { colors } from '../../theme/colors';
import SText from '../shared/SText';

interface Props {
  toCloud?: boolean;
  onChange(toCloud: boolean): void;
}

export default function CloudField({ toCloud, onChange }: Props) {
  const [cloudDestination, setCloudDestination] = useState(toCloud ?? false);
  const options = ['local', 'cloud'];

  return (
    <View
      style={{
        flexDirection: 'row',
        paddingLeft: 16,
        alignItems: 'center',
      }}
    >
      <SText style={{ flex: 1, fontWeight: '500', fontSize: 16 }}>Destination</SText>
      <View style={{ flexDirection: 'row' }}>
        {options.map((e, i) => {
          const value = e === 'cloud';
          return (
            <View key={i} style={{ flexDirection: 'row' }}>
              <Pressable
                onPress={() => {
                  setCloudDestination(value);
                  onChange(value);
                }}
                style={{
                  paddingVertical: 14,
                  paddingHorizontal: 16,
                }}
              >
                <SText
                  style={{
                    color: cloudDestination === value ? colors.white : colors.onCard,
                    fontSize: 16,
                  }}
                >
                  {e.at(0)?.toUpperCase() + e.substring(1)}
                </SText>
              </Pressable>
              {i < options.length - 1 && (
                <View style={{ width: 1, backgroundColor: colors.onCard, marginVertical: 12 }} />
              )}
            </View>
          );
        })}
      </View>
    </View>
  );
}
