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
          'flex h-10 items-center gap-2 rounded-md border bg-bg px-2.5 focus-within:ring-2 focus-within:ring-line',
          error ? 'border-danger' : 'border-strong'
        )}
      >
        {left}
        <input
          ref={ref}
          id={inputId}
          aria-invalid={!!error}
          className="h-full min-w-0 flex-1 bg-transparent text-base text-fg outline-none placeholder:text-muted"
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
        className="h-9 w-full cursor-pointer appearance-none rounded-md border border-strong bg-bg pr-9 pl-2.5 text-base text-fg outline-none focus:ring-2 focus:ring-line"
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
        className="pointer-events-none absolute top-1/2 right-2.5 -translate-y-1/2"
      />
    </div>
  );
}
