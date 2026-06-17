import Svg, { Path } from 'react-native-svg';

interface Props {
  size: string;
  color: string;
}

export default function ChevronRightIcon({ size, color }: Props) {
  return (
    <Svg height={size} viewBox="0 -960 960 960" width={size} fill={color}>
      <Path d="M504-480 320-664l56-56 240 240-240 240-56-56 184-184Z" />
    </Svg>
  );
}
