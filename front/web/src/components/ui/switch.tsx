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
        'relative inline-flex h-4 w-[30px] shrink-0 cursor-pointer items-center rounded-[5px] border transition disabled:cursor-default disabled:opacity-40',
        checked ? 'border-primary bg-primary' : 'border-line bg-surface'
      )}
    >
      <span
        className={cx(
          'absolute size-2.5 rounded-full transition-all',
          checked ? 'left-[16px] bg-on-primary' : 'left-[3px] bg-muted'
        )}
      />
    </button>
  );
}
