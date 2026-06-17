import Svg, { Path } from 'react-native-svg';

interface Props {
  size: string;
  color: string;
}

export default function OpenFolderIcon({ size, color }: Props) {
  return (
    <Svg height={size} viewBox="0 -960 960 960" width={size} fill={color}>
      <Path d="M160-160q-33 0-56.5-23.5T80-240v-480q0-33 23.5-56.5T160-800h240l80 80h320q33 0 56.5 23.5T880-640H160v400l96-320h684L837-217q-8 26-29.5 41.5T760-160H160Z" />
    </Svg>
  );
}
