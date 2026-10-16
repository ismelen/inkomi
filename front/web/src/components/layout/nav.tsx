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
        'min-w-6 rounded-sm bg-warning px-1.5 text-center text-xs leading-[18px] text-white',
        className
      )}
    >
      {count}
    </span>
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
        'hidden w-16 shrink-0 flex-col border-r border-surface bg-subtle px-2 py-6 md:flex lg:w-[220px] lg:px-3',
        disabled && 'pointer-events-none'
      )}
    >
      <div className="mb-6 flex items-center gap-2.5 px-2 lg:px-3">
        <span className="size-[26px] shrink-0 rounded-md bg-primary" />
        <span className="hidden text-lg tracking-wide lg:inline">INKOMI</span>
      </div>
      <nav aria-label={t('nav.menu')} className="flex flex-col gap-1.5">
        {entries.map((entry) => (
          <NavLink
            key={entry.to}
            to={entry.to}
            end={entry.to === '/'}
            title={entry.label}
            className={({ isActive }) =>
              cx(
                'relative flex h-[34px] items-center justify-center gap-3 rounded-md px-3 text-base lg:justify-start',
                disabled ? 'text-line' : isActive ? 'bg-surface' : 'hover:bg-hover'
              )
            }
          >
            <Icon name={entry.icon} />
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
        <div className="mt-auto hidden border-t border-surface px-3 pt-4 lg:block">
          <p className="truncate text-[13px]">{profile?.email ?? '…'}</p>
          <p className="mt-1 flex items-center gap-1.5 text-xs text-muted">
            <span className={cx('size-1.5 rounded-full', linked ? 'bg-success' : 'bg-line')} />
            {linked ? t('shell.dropboxConnected') : t('shell.dropboxDisconnected')}
          </p>
          <button
            type="button"
            onClick={() => void signOut()}
            className="mt-3 cursor-pointer text-sm text-info hover:underline"
          >
            {t('settings.signOut')}
          </button>
        </div>
      )}
    </aside>
  );
}

/** Phones: the sections at the bottom, like a native tab bar. */
export function BottomNav() {
  const { t } = useTranslation();
  const entries = useNavEntries();
  const queueCount = useActivity((s) => s.queue.length);
  return (
    <nav
      aria-label={t('nav.menu')}
      className="pb-safe flex shrink-0 border-t border-surface bg-bg md:hidden"
    >
      {entries.map((entry) => (
        <NavLink
          key={entry.to}
          to={entry.to}
          end={entry.to === '/'}
          className={({ isActive }) =>
            cx(
              'relative flex h-14 flex-1 flex-col items-center justify-center gap-0.5 text-xs',
              isActive ? 'text-fg' : 'text-muted'
            )
          }
        >
          <span className="relative">
            <Icon name={entry.icon} size={22} />
            {entry.to === '/activity' && (
              <Badge count={queueCount} className="absolute -top-1.5 left-4 min-w-4" />
            )}
          </span>
          {entry.label}
        </NavLink>
      ))}
    </nav>
  );
}
