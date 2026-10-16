import { NavLink } from 'react-router';
import { useTranslation } from 'react-i18next';
import { IconName } from '../icons/icons';
import { useActivity } from '../../stores/use-activity';
import { useProfile } from '../../stores/use-profile';
import { useSession } from '../../stores/use-session';
import { cx } from '../ui/cx';
import { Icon } from '../ui/icon';

interface NavEntry {
  to: string;
  icon: IconName;
  label: string;
}

function useNavEntries(): NavEntry[] {
  const { t } = useTranslation();
  return [
    { to: '/', icon: 'attach_file', label: t('nav.list') },
    { to: '/activity', icon: 'arrow_circle_up', label: t('nav.activity') },
    { to: '/search', icon: 'search', label: t('nav.search') },
    { to: '/settings', icon: 'settings', label: t('nav.settings') },
  ];
}

function Badge({ count, className }: { count: number; className?: string }) {
  if (count === 0) return null;
  return (
    <span
      className={cx(
        'min-w-6 rounded-full bg-primary px-1.5 text-center font-mono text-xs leading-[18px] text-on-primary',
        className
      )}
    >
      {count}
    </span>
  );
}

/** The app icon (front/mobile/assets/icon.png) and the name. */
export function Brand({
  className,
  nameClassName,
}: {
  className?: string;
  nameClassName?: string;
}) {
  return (
    <div className={cx('flex items-center gap-2.5', className)}>
      <img src="/icon.png" alt="" className="size-8 shrink-0 rounded-md" />
      <span className={cx('text-xl font-bold tracking-tight text-primary', nameClassName)}>
        Inkomi
      </span>
    </div>
  );
}

/** Sidebar of the "Web" frame: 220px with labels on desktop, icons only on tablets. */
export function Sidebar({ disabled }: { disabled?: boolean }) {
  const { t } = useTranslation();
  const entries = useNavEntries();
  const queueCount = useActivity((s) => s.queue.length);
  const profile = useProfile((s) => s.profile);
  const signOut = useSession((s) => s.signOut);
  const linked = profile?.dropbox.linked ?? false;

  return (
    <aside
      className={cx(
        'z-10 hidden w-[72px] shrink-0 flex-col bg-card px-3 py-6 shadow-card md:flex lg:w-[232px] lg:px-4',
        disabled && 'pointer-events-none'
      )}
    >
      <Brand
        className="mb-8 justify-center lg:justify-start lg:px-2"
        nameClassName="max-lg:hidden"
      />
      <nav aria-label={t('nav.menu')} className="flex flex-col gap-1.5">
        {entries.map((entry) => (
          <NavLink
            key={entry.to}
            to={entry.to}
            end={entry.to === '/'}
            title={entry.label}
            className={({ isActive }) =>
              cx(
                'relative flex h-11 items-center justify-center gap-3 rounded-input px-3 text-base transition lg:justify-start',
                disabled
                  ? 'text-faint'
                  : isActive
                    ? 'bg-accent font-semibold text-on-accent shadow-card'
                    : 'font-medium text-muted hover:bg-hover hover:text-primary'
              )
            }
          >
            <Icon name={entry.icon} size={22} />
            <span className="hidden flex-1 lg:inline">{entry.label}</span>
            {entry.to === '/activity' && !disabled && (
              <Badge
                count={queueCount}
                className="absolute -top-1 -right-1 min-w-4 lg:static lg:min-w-6"
              />
            )}
          </NavLink>
        ))}
      </nav>
      {!disabled && (
        <div className="mt-auto hidden rounded-lg bg-subtle p-3 lg:block">
          <p className="truncate text-sm font-medium">{profile?.email ?? '…'}</p>
          <p className="mt-1 flex items-center gap-1.5 text-xs text-muted">
            <span className={cx('size-2 rounded-full', linked ? 'bg-success' : 'bg-line')} />
            {linked ? t('shell.dropboxConnected') : t('shell.dropboxDisconnected')}
          </p>
          <button
            type="button"
            onClick={() => void signOut()}
            className="mt-3 cursor-pointer text-sm font-semibold text-primary hover:underline"
          >
            {t('settings.signOut')}
          </button>
        </div>
      )}
    </aside>
  );
}
