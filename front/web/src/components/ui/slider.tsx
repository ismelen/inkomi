import { ReactNode, useId } from 'react';
import { Range } from '../../models/manga-config';
import { formatNumber } from '../../utils/format';
import { cx } from './cx';

interface Props {
  label: string;
  value: number;
  range: Range;
  onChange: (value: number) => void;
  format?: (value: number) => string;
  disabled?: boolean;
  accessory?: ReactNode;
}

export function Slider({ label, value, range, onChange, format, disabled, accessory }: Props) {
  const id = useId();
  return (
    <div className="py-1">
      <div className="flex items-center gap-2">
        <label htmlFor={id} className={cx('flex-1 text-sm', disabled && 'text-faint')}>
          {label}
        </label>
        {accessory}
        <span className="rounded-md bg-surface px-1.5 font-mono text-xs text-primary tabular-nums">
          {format ? format(value) : formatNumber(value, range.step)}
        </span>
      </div>
      <input
        id={id}
        type="range"
        min={range.min}
        max={range.max}
        step={range.step}
        value={value}
        disabled={disabled}
        onChange={(e) => onChange(Number(e.target.value))}
        className="h-5 w-full cursor-pointer disabled:cursor-default disabled:opacity-40"
      />
    </div>
  );
}
