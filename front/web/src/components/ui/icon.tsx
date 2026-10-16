import { IconName, icons } from '../icons/icons';
import { cx } from './cx';

interface Props {
  name: IconName;
  size?: number;
  className?: string;
}

/** Material Symbols Rounded (viewBox 0 -960 960 960), painted with currentColor. */
export function Icon({ name, size = 20, className }: Props) {
  return (
    <svg
      viewBox="0 -960 960 960"
      width={size}
      height={size}
      fill="currentColor"
      aria-hidden
      className={cx('shrink-0', className)}
    >
      <path d={icons[name]} />
    </svg>
  );
}
