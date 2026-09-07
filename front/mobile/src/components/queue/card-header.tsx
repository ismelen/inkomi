import React from 'react';
import { StyleSheet, View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { Transaction } from '../../models/transaction';
import { colors } from '../../theme/colors';
import SIcon from '../icons/SIcon';
import SText from '../shared/SText';

type TransactionStatus = Transaction['status'];

export const statusConfig: Record<TransactionStatus, { labelKey: string; bg: string; fg: string }> =
  {
    processing: {
      labelKey: 'queue.status.processing',
      bg: colors.primary_fixed,
      fg: colors.on_primary,
    },
    merging: {
      labelKey: 'queue.status.merging',
      bg: colors.secondary_fixed,
      fg: colors.on_secondary_fixed,
    },
    error: { labelKey: 'queue.status.error', bg: colors.error_container, fg: colors.error },
    done: { labelKey: 'queue.status.done', bg: colors.ok, fg: '#ffffff' },
    waiting: {
      labelKey: 'queue.status.waiting',
      bg: colors.surface_container_high,
      fg: colors.on_surface_variant,
    },
    canceled: {
      labelKey: 'queue.status.canceled',
      bg: colors.outline_variant,
      fg: colors.on_surface_variant,
    },
    unknown: {
      labelKey: 'queue.status.unknown',
      bg: colors.surface_container_high,
      fg: colors.on_surface_variant,
    },
    enqueued: {
      labelKey: 'queue.status.enqueued',
      bg: colors.surface_container_high,
      fg: colors.outline,
    },
  };

export function CardHeader({ data }: { data: Transaction }) {
  const { t } = useTranslation();
  const title = data.config.title || data.id.slice(0, 8);
  const cfg = statusConfig[data.status] || statusConfig['unknown'];
  const isProcessing = data.status === 'processing';

  return (
    <View style={{ gap: 6 }}>
      <View style={s.row}>
        <View style={[s.row, { flex: 1, gap: 8 }]}>
          {isProcessing && <SIcon name="autorenew" color={colors.primary} size={20} />}
          <SText style={s.title} numberOfLines={2}>
            {title}
          </SText>
        </View>
        <StatusBadge label={t(cfg.labelKey)} bg={cfg.bg} fg={cfg.fg} />
      </View>
      <View style={[s.row, { justifyContent: 'space-between' }]}>
        <View style={[s.row, { gap: 12 }]}>
          <FormatBadge mode={data.config.mode} />
          <DestinationBadge toCloud={data.config.toCloud} />
        </View>
        <SText style={s.meta}>
          {new Date(data.timestamp * 1000).toLocaleString(undefined, {
            day: 'numeric',
            month: 'short',
            hour: '2-digit',
            minute: '2-digit',
          })}
        </SText>
      </View>
    </View>
  );
}

function StatusBadge({ label, bg, fg }: { label: string; bg: string; fg: string }) {
  return (
    <View style={[s.badge, { backgroundColor: bg }]}>
      <SText style={{ color: fg, fontSize: 10, fontFamily: 'semibold' }}>{label}</SText>
    </View>
  );
}

function FormatBadge({ mode }: { mode: string }) {
  const { t } = useTranslation();
  return (
    <View style={[s.row, { gap: 4 }]}>
      <SText style={s.meta}>{t('queue.format')}</SText>
      <SText style={[s.meta, { fontFamily: 'medium' }]}>{mode}</SText>
    </View>
  );
}

function DestinationBadge({ toCloud }: { toCloud: boolean }) {
  const { t } = useTranslation();
  return (
    <View style={[s.row, { gap: 4 }]}>
      <SIcon name="cloud" color={colors.outline} size={14} type="outlined" />
      <SText style={s.meta}>{toCloud ? t('queue.cloud') : t('queue.local')}</SText>
    </View>
  );
}

const s = StyleSheet.create({
  row: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  title: {
    fontSize: 16,
    fontFamily: 'semibold',
    color: colors.on_surface,
    flexShrink: 1,
  },
  badge: {
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 12,
  },
  meta: {
    fontSize: 12,
    color: colors.outline,
  },
});
