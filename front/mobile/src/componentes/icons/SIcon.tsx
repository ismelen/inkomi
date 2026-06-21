import Svg, { Path } from 'react-native-svg';

import React from 'react';
import { icons } from './icons';

interface BaseIconProps {
  size: number;
  color: string;
}

interface SIconProps extends BaseIconProps {
  name: string;
  type?: 'filled' | 'outlined';
}

interface ConcreteIconProps extends BaseIconProps {
  path: string;
}

export default function SIcon({ type = 'filled', name, ...props }: SIconProps) {
  const path = icons[type][name] as string;
  return <ConcreteIcon path={path} {...props} />;
}

function ConcreteIcon({ path, size, color }: ConcreteIconProps) {
  return (
    <Svg height={`${size}px`} viewBox="0 -960 960 960" width={`${size}px`} fill={color}>
      <Path d={path} />
    </Svg>
  );
}
