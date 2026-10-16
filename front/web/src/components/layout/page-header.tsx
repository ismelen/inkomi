import { ReactNode } from 'react';
import { useNavigate } from 'react-router';
import { useTranslation } from 'react-i18next';
import { cx } from '../ui/cx';
import { IconButton } from '../ui/icon-button';

/** Back like the app's header: to the previous page, or to the list when opened directly. */
export function useBack(): () => void {
  const navigate = useNavigate();
  return () => {
    if ((window.history.state?.idx ?? 0) > 0) void navigate(-1);
    else void navigate('/', { replace: true });
  };
}

interface Props {
  title: string;
  children?: ReactNode;
  /** Phones: back chevron and a compact title, like the app's subscreens. */
  back?: boolean;
  className?: string;
}

/** Title of a section (display-lg of DESIGN.md) with its actions on the right. */
export function PageHeader({ title, children, back, className }: Props) {
  const { t } = useTranslation();
  const goBack = useBack();
  return (
    <header
      className={cx(
        'flex min-h-14 shrink-0 items-center gap-3 px-4 pt-3 md:px-8 md:pt-6 lg:min-h-[60px]',
        back && 'max-md:gap-1 max-md:px-2 max-md:pt-0',
        className
      )}
    >
      {back && (
        <IconButton
          icon="chevron_left"
          label={t('common.back')}
          size={24}
          className="text-fg md:hidden"
          onClick={goBack}
        />
      )}
      <h1
        className={cx(
          'min-w-0 flex-1 truncate text-2xl font-bold',
          back && 'max-md:text-lg max-md:font-semibold max-md:tracking-normal'
        )}
      >
        {title}
      </h1>
      {children}
    </header>
  );
}
