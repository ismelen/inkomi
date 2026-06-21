import Svg, { Path } from 'react-native-svg';

interface Props {
  size: string;
  color: string;
}

export default function BookIcon({ size, color }: Props) {
  return (
    <Svg height={size} viewBox="0 -960 960 960" width={size} fill={color}>
      <Path d="M240-80q-33 0-56.5-23.5T160-160v-640q0-33 23.5-56.5T240-880h480q33 0 56.5 23.5T800-800v640q0 33-23.5 56.5T720-80H240Zm200-440 100-60 100 60v-280H440v280Z" />
    </Svg>
  );
}
