import React, { useState } from 'react';
import { View } from 'react-native';
import { colors } from '../../theme/colors';
import SButton from '../shared/SButton';
import SText from '../shared/SText';
import { Destination } from '../../models/transaction-request';

const destinations: Destination[] = ['local', 'cloud'];

interface Props {
  initDestination: Destination;
  onChange(destination: Destination): void;
}

export default function DestinationSelector({ initDestination, onChange }: Props) {
  const [destination, setDestination] = useState(initDestination);

  return (
    <View
      style={{
        borderRadius: 12,
        backgroundColor: colors.surface_container_lowest,
        boxShadow: colors.boxShadow,
        padding: 5,
        flexDirection: 'row',
        gap: 5,
      }}
    >
      {destinations.map((dest) => (
        <SButton
          key={dest}
          onPress={() => setDestination(dest)}
          style={{
            flex: 1,
            alignItems: 'center',
            justifyContent: 'center',
            paddingVertical: 10,
            borderRadius: 7,
            backgroundColor: dest === destination ? colors.primary_container : 'transparent',
          }}
        >
          <SText
            style={{
              fontFamily: 'semibold',
              color: dest === destination ? colors.on_primary : colors.on_surface,
            }}
          >
            {dest[0].toUpperCase() + dest.slice(1)}
          </SText>
        </SButton>
      ))}
    </View>
  );
}
