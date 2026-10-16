import { ButtonHTMLAttributes } from 'react';
import { IconName } from '../icons/icons';
import { cx } from './cx';
import { Icon } from './icon';

interface Props extends ButtonHTMLAttributes<HTMLButtonElement> {
  icon: IconName;
  /** Accessible name, also the tooltip. */
  label: string;
  size?: number;
  /** Small counter on the top right corner (the app's SIconButton). */
  badge?: number;
}

export function IconButton({ icon, label, size = 20, badge, className, ...rest }: Props) {
  return (
    <button
      type="button"
      aria-label={label}
      title={label}
      className={cx(
        'relative inline-flex size-9 shrink-0 cursor-pointer items-center justify-center rounded-full text-muted transition hover:bg-hover hover:text-primary focus-visible:outline-2 focus-visible:outline-primary disabled:cursor-default disabled:opacity-40',
        className
      )}
      {...rest}
    >
      <Icon name={icon} size={size} />
      {!!badge && (
        <span className="absolute top-0.5 right-0.5 min-w-4 rounded-full bg-primary px-1 text-center font-mono text-[10px] leading-4 text-on-primary">
          {badge > 99 ? '99+' : badge}
        </span>
      )}
    </button>
  );
}
