import { ReactNode } from 'react';
import { cx } from './cx';

export function Card({
  title,
  children,
  className,
}: {
  title?: string;
  children: ReactNode;
  className?: string;
}) {
  return (
    <section className={cx('rounded-lg border border-strong p-4 md:p-5', className)}>
      {title && <h2 className="mb-3 text-lg">{title}</h2>}
      {children}
    </section>
  );
}
