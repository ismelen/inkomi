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
          'inline-flex size-3.5 items-center justify-center rounded-[3px] border border-strong',
          checked && 'bg-primary text-on-primary'
        )}
      >
        {checked && <Icon name="check" size={12} />}
      </span>
    </button>
  );
}
