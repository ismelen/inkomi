import { cx } from './cx';

interface Props<T extends string> {
  value: T;
  options: { value: T; label: string }[];
  onChange: (value: T) => void;
  /** "tabs": grey track with a raised tab (login); "outline": filled active option. */
  variant?: 'outline' | 'tabs';
  label?: string;
}

export function Segmented<T extends string>({
  value,
  options,
  onChange,
  variant = 'outline',
  label,
}: Props<T>) {
  const tabs = variant === 'tabs';
  return (
    <div
      role="radiogroup"
      aria-label={label}
      className={cx('flex rounded-md p-0.5', tabs ? 'bg-surface' : 'border border-strong')}
    >
      {options.map((option) => {
        const active = option.value === value;
        return (
          <button
            key={option.value}
            type="button"
            role="radio"
            aria-checked={active}
            onClick={() => onChange(option.value)}
            className={cx(
              'h-8 flex-1 cursor-pointer rounded-[6px] border text-base transition',
              tabs
                ? active
                  ? 'border-line bg-bg text-fg'
                  : 'border-transparent text-muted'
                : active
                  ? 'border-primary bg-primary text-on-primary'
                  : 'border-transparent text-fg hover:bg-hover'
            )}
          >
            {option.label}
          </button>
        );
      })}
    </div>
  );
}
