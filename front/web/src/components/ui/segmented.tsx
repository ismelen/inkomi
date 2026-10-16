import { cx } from './cx';

interface Props<T extends string> {
  value: T;
  options: { value: T; label: string }[];
  onChange: (value: T) => void;
  /** "tabs": track with a raised tab (login); "outline": track with a filled active option. */
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
    <div role="radiogroup" aria-label={label} className="flex rounded-input bg-surface p-1">
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
              'h-8 flex-1 cursor-pointer rounded-md text-base font-medium transition',
              tabs
                ? active
                  ? 'bg-card text-primary shadow-card'
                  : 'text-muted hover:text-fg'
                : active
                  ? 'bg-primary text-on-primary shadow-card'
                  : 'text-muted hover:text-fg'
            )}
          >
            {option.label}
          </button>
        );
      })}
    </div>
  );
}
