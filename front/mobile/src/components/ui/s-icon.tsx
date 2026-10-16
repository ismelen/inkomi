import React from 'react';
import Svg, { Path } from 'react-native-svg';
import { colors } from '../../theme';
import { IconName, icons } from '../icons/icons';

interface Props {
  name: IconName;
  size?: number;
  color?: string;
}

/** Material Symbols Rounded (viewBox 0 -960 960 960). Add new paths to icons.ts. */
export function SIcon({ name, size = 22, color = colors.text }: Props) {
  return (
    <Svg width={size} height={size} viewBox="0 -960 960 960" fill={color}>
      <Path d={icons[name]} />
    </Svg>
  );
}
