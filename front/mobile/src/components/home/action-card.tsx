import React from 'react';
import { BackHandler, Pressable, View } from 'react-native';
import SText from '../shared/SText';
import { colors, hexToRgba } from '../../theme/colors';
import SIcon from '../icons/SIcon';
import SButton from '../shared/SButton';

interface Props {
  title: string;
  subtitle: string;
  icon: string;
  tag: string;
  onClick(): void;
}

export default function ActionCard({ title, subtitle, icon, tag, onClick }: Props) {
  return (
    <SButton
      onPress={onClick}
      style={{
        flexDirection: 'row',
        borderRadius: 12,
        backgroundColor: colors.surface_container_lowest,
        boxShadow: `0px 4px 12px 0px ${hexToRgba(colors.on_surface, 0.04)}`,
        padding: 16,
        overflow: 'hidden',
      }}
    >
      <View style={{ flex: 1 }}>
        <SText style={{ fontFamily: 'semibold', fontSize: 20, color: colors.on_surface }}>
          {title}
        </SText>
        <SText style={{ fontSize: 14, color: colors.on_surface_variant, marginBottom: 12 }}>
          {subtitle}
        </SText>
        <Tag text={tag} />
      </View>
      <View
        style={{
          opacity: 0.2,
          width: 128,
          height: 128,
          borderRadius: 9999,
          backgroundColor: colors.primary_fixed,
          position: 'absolute',
          top: -32,
          right: -32,
        }}
      ></View>
      <View style={{ width: 96, alignItems: 'flex-end' }}>
        <View
          style={{
            backgroundColor: colors.primary_container,
            width: 48,
            height: 48,
            alignItems: 'center',
            justifyContent: 'center',
            borderRadius: 9999,
          }}
        >
          <SIcon name={icon} color={colors.on_primary} size={22} type="outlined" />
        </View>
      </View>
    </SButton>
  );
}

function Tag({ text }: { text: string }) {
  return (
    <SText
      style={{
        fontFamily: 'medium',
        fontSize: 12,
        padding: 4,
        paddingHorizontal: 8,
        borderRadius: 8,
        backgroundColor: colors.surface_variant,
        alignSelf: 'flex-start',
      }}
    >
      {text}
    </SText>
  );
}
