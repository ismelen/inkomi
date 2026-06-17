import Svg, { Path } from 'react-native-svg';

interface Props {
  size: string;
  color: string;
}

export default function CloseIcon({ size, color }: Props) {
  return (
    <Svg height={size} viewBox="0 -960 960 960" width={size} fill={color}>
      <Path d="m256-200-56-56 224-224-224-224 56-56 224 224 224-224 56 56-224 224 224 224-56 56-224-224-224 224Z" />
    </Svg>
  );
}
