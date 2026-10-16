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
    <section className={cx('rounded-lg bg-card p-4 shadow-card md:p-5', className)}>
      {title && <h2 className="mb-3 text-lg font-semibold">{title}</h2>}
      {children}
    </section>
  );
}
