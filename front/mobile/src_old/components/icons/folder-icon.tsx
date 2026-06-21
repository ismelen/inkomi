import Svg, { Path } from 'react-native-svg';

interface Props {
  size: string;
  color: string;
}

export default function FolderIcon({ size, color }: Props) {
  return (
    <Svg height={size} viewBox="0 -960 960 960" width={size} fill={color}>
      <Path d="M160-160q-33 0-56.5-23.5T80-240v-480q0-33 23.5-56.5T160-800h207q16 0 30.5 6t25.5 17l57 57h320q33 0 56.5 23.5T880-640v400q0 33-23.5 56.5T800-160H160Z" />
    </Svg>
  );
}
