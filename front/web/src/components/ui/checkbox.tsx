import { cx } from './cx';
import { Icon } from './icon';

interface Props {
  checked: boolean;
  onChange: () => void;
  label: string;
}

export function Checkbox({ checked, onChange, label }: Props) {
  return (
    <button
      type="button"
      role="checkbox"
      aria-checked={checked}
      aria-label={label}
      onClick={(e) => {
        e.stopPropagation();
        onChange();
      }}
      className="-m-2 inline-flex shrink-0 cursor-pointer p-2"
    >
      <span
        className={cx(
          'inline-flex size-[18px] items-center justify-center rounded-sm border-2 transition',
          checked ? 'border-primary bg-primary text-on-primary' : 'border-strong bg-card'
        )}
      >
        {checked && <Icon name="check" size={14} />}
      </span>
    </button>
  );
}
