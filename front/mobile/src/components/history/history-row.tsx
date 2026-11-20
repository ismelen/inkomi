import React from 'react';
import { Pressable, View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { colors, radii, spacing } from '../../theme';
import {
  downloadablesOf,
  isCancelled,
  isDownloadable,
  isFolder,
  treeError,
  treeFinishedAt,
  treeSize,
  UploadTree,
} from '../../models/upload-tree';
import { formatBytes, formatWhen } from '../../utils/format';
import { SIconButton } from '../ui/s-icon-button';
import { SText } from '../ui/s-text';
import { ErrorText } from './error-text';

interface Props {
  tree: UploadTree;
  busy: boolean;
  onDownload: (sourceId: string) => void;
  onShare: (sourceId: string) => void;
  onOpenResults: (tree: UploadTree) => void;
  onShowError: (error: string) => void;
}

export function HistoryRow({ tree, busy, onDownload, onShare, onOpenResults, onShowError }: Props) {
  const { t, i18n } = useTranslation();
  const error = treeError(tree);
  const files = downloadablesOf(tree);
  const multiple = isFolder(tree);
  const single = !multiple && isDownloadable(tree.root) ? tree.root : null;

  const summary = [
    tree.root.join && multiple
      ? t('history.joined')
      : t('history.books', { count: Math.max(files.length, 1) }),
    formatBytes(treeSize(tree)),
    formatWhen(treeFinishedAt(tree), i18n.language),
  ]
    .filter(Boolean)
    .join(' · ');

  return (
    <Pressable
      onPress={multiple ? () => onOpenResults(tree) : undefined}
      disabled={!multiple}
      android_ripple={multiple ? { color: colors.ripple } : undefined}
      style={{
        flexDirection: 'row',
        alignItems: 'center',
        gap: spacing.sm,
        paddingVertical: spacing.sm,
        paddingHorizontal: spacing.sm,
        borderRadius: radii.md,
      }}
    >
      <View style={{ flex: 1, gap: spacing.xxs }}>
        <SText variant="label" numberOfLines={1}>
          {tree.root.title}
        </SText>
        <SText variant="small" tone="textSecondary" numberOfLines={1}>
          {summary}
        </SText>
        {isCancelled(tree) ? (
          <SText variant="small" tone="textSecondary">
            {t('history.cancelled')}
          </SText>
        ) : error !== null ? (
          <ErrorText error={error} onPress={onShowError} />
        ) : (
          <View style={{ flexDirection: 'row', alignItems: 'center', gap: spacing.xs }}>
            <View
              style={{
                width: 6,
                height: 6,
                borderRadius: radii.pill,
                backgroundColor: colors.success,
              }}
            />
            <SText variant="small" tone="success">
              {t('history.sent')}
            </SText>
          </View>
        )}
      </View>
      {single && (
        <>
          <SIconButton
            icon="download"
            label={t('history.download')}
            size={20}
            disabled={busy}
            onPress={() => onDownload(single.id)}
          />
          <SIconButton
            icon="share"
            label={t('history.share')}
            size={20}
            disabled={busy}
            onPress={() => onShare(single.id)}
          />
        </>
      )}
      {multiple && <SIconButton icon="chevron_right" label={t('history.results')} size={20} />}
    </Pressable>
  );
}
