import { ReactNode } from 'react';

/** Title of a section (24px in the design) with its actions on the right. */
export function PageHeader({ title, children }: { title: string; children?: ReactNode }) {
  return (
    <header className="flex min-h-14 shrink-0 items-center gap-3 px-4 pt-3 md:px-8 md:pt-6 lg:min-h-[60px]">
      <h1 className="min-w-0 flex-1 truncate text-xl md:text-2xl">{title}</h1>
      {children}
    </header>
  );
}
