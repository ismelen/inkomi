import { cx } from './cx';

interface Props {
  checked: boolean;
  onChange: (value: boolean) => void;
  label: string;
  disabled?: boolean;
}

export function Switch({ checked, onChange, label, disabled }: Props) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      aria-label={label}
      disabled={disabled}
      onClick={(e) => {
        e.stopPropagation();
        onChange(!checked);
      }}
      className={cx(
        'relative inline-flex h-6 w-11 shrink-0 cursor-pointer items-center rounded-full transition-colors duration-200 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary disabled:cursor-default disabled:opacity-40',
        checked ? 'bg-primary' : 'bg-variant'
      )}
    >
      <span
        className={cx(
          'absolute size-5 rounded-full bg-thumb shadow-raised transition-all duration-200',
          checked ? 'left-[22px]' : 'left-0.5'
        )}
      />
    </button>
  );
}
