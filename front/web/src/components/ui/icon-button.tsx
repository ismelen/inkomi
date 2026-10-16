import { ButtonHTMLAttributes } from 'react';
import { IconName } from '../icons/icons';
import { cx } from './cx';
import { Icon } from './icon';

interface Props extends ButtonHTMLAttributes<HTMLButtonElement> {
  icon: IconName;
  /** Accessible name, also the tooltip. */
  label: string;
  size?: number;
}

export function IconButton({ icon, label, size = 20, className, ...rest }: Props) {
  return (
    <button
      type="button"
      aria-label={label}
      title={label}
      className={cx(
        'relative inline-flex size-9 shrink-0 cursor-pointer items-center justify-center rounded-md text-fg transition hover:bg-hover disabled:cursor-default disabled:opacity-40',
        className
      )}
      {...rest}
    >
      <Icon name={icon} size={size} />
    </button>
  );
}
