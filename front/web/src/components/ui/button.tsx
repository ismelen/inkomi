import { ButtonHTMLAttributes } from 'react';
import { IconName } from '../icons/icons';
import { cx } from './cx';
import { Icon } from './icon';
import { Spinner } from './spinner';

type Variant = 'primary' | 'secondary' | 'ghost' | 'danger';
type Size = 'sm' | 'md' | 'lg';

interface Props extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
  size?: Size;
  icon?: IconName;
  loading?: boolean;
}

const variants: Record<Variant, string> = {
  primary: 'bg-primary text-on-primary border border-primary hover:opacity-90',
  secondary: 'bg-bg text-fg border border-strong hover:bg-hover',
  ghost: 'text-fg border border-transparent hover:bg-hover',
  danger: 'bg-danger text-on-primary border border-danger hover:opacity-90',
};

const sizes: Record<Size, string> = {
  sm: 'h-8 px-3 text-sm gap-1.5',
  md: 'h-10 px-4 text-base gap-2',
  lg: 'h-11 px-5 text-md gap-2',
};

export function Button({
  variant = 'primary',
  size = 'md',
  icon,
  loading,
  disabled,
  className,
  children,
  type = 'button',
  ...rest
}: Props) {
  return (
    <button
      type={type}
      disabled={disabled || loading}
      className={cx(
        'inline-flex shrink-0 cursor-pointer items-center justify-center rounded-md font-medium whitespace-nowrap transition disabled:cursor-default disabled:opacity-40',
        variants[variant],
        sizes[size],
        className
      )}
      {...rest}
    >
      {loading ? <Spinner /> : icon && <Icon name={icon} size={size === 'sm' ? 16 : 20} />}
      {children}
    </button>
  );
}
