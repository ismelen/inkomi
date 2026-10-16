import { forwardRef, InputHTMLAttributes, ReactNode, SelectHTMLAttributes, useId } from 'react';
import { cx } from './cx';
import { Icon } from './icon';

export function SectionLabel({ children, className }: { children: ReactNode; className?: string }) {
  return <div className={cx('text-section text-muted uppercase', className)}>{children}</div>;
}

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  hint?: string;
  error?: string;
  left?: ReactNode;
  right?: ReactNode;
}

export const TextInput = forwardRef<HTMLInputElement, InputProps>(function TextInput(
  { label, hint, error, left, right, className, id, ...rest },
  ref
) {
  const fallbackId = useId();
  const inputId = id ?? fallbackId;
  return (
    <div className={className}>
      {label && (
        <label htmlFor={inputId} className="mb-1.5 block text-section text-muted uppercase">
          {label}
        </label>
      )}
      <div
        className={cx(
          'flex h-11 items-center gap-2 rounded-input border bg-subtle px-3 transition focus-within:border-primary focus-within:bg-card focus-within:ring-1 focus-within:ring-primary',
          error ? 'border-danger' : 'border-transparent'
        )}
      >
        {left}
        <input
          ref={ref}
          id={inputId}
          aria-invalid={!!error}
          className="h-full min-w-0 flex-1 bg-transparent text-base text-fg outline-none placeholder:text-faint"
          {...rest}
        />
        {right}
      </div>
      {(error || hint) && (
        <p className={cx('mt-1 text-sm', error ? 'text-danger' : 'text-muted')}>{error ?? hint}</p>
      )}
    </div>
  );
});

interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  options: { value: string; label: string }[];
}

export function Select({ options, className, ...rest }: SelectProps) {
  return (
    <div className={cx('relative', className)}>
      <select
        className="h-11 w-full cursor-pointer appearance-none rounded-input border border-transparent bg-subtle pr-10 pl-3 text-base text-fg outline-none transition focus:border-primary focus:bg-card focus:ring-1 focus:ring-primary"
        {...rest}
      >
        {options.map((o) => (
          <option key={o.value} value={o.value}>
            {o.label}
          </option>
        ))}
      </select>
      <Icon
        name="expand_more"
        className="pointer-events-none absolute top-1/2 right-3 -translate-y-1/2 text-primary"
      />
    </div>
  );
}
