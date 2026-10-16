export function formatBytes(bytes: number | null | undefined): string {
  if (!bytes || bytes <= 0) return '';
  const units = ['B', 'KB', 'MB', 'GB'];
  let value = bytes;
  let unit = 0;
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024;
    unit++;
  }
  const digits = value >= 10 || unit === 0 ? 0 : 1;
  return `${value.toFixed(digits)} ${units[unit]}`;
}

export function formatNumber(value: number, step: number): string {
  const decimals = step >= 1 ? 0 : Math.min(2, String(step).split('.')[1]?.length ?? 2);
  return value.toFixed(decimals);
}

function startOfDay(date: Date): number {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate()).getTime();
}

export type DayBucket = 'today' | 'yesterday' | 'week' | 'older';

export function dayBucket(iso: string, now = new Date()): DayBucket {
  const day = startOfDay(new Date(iso));
  const today = startOfDay(now);
  const dayMs = 86_400_000;
  if (day >= today) return 'today';
  if (day >= today - dayMs) return 'yesterday';
  if (day > today - 7 * dayMs) return 'week';
  return 'older';
}

/** 10:32 for today and yesterday, the weekday this week, the date before. */
export function formatWhen(iso: string, locale: string, now = new Date()): string {
  const date = new Date(iso);
  switch (dayBucket(iso, now)) {
    case 'today':
    case 'yesterday':
      return date.toLocaleTimeString(locale, { hour: '2-digit', minute: '2-digit' });
    case 'week':
      return date.toLocaleDateString(locale, { weekday: 'short' });
    default:
      return date.toLocaleDateString(locale, { day: 'numeric', month: 'short' });
  }
}
