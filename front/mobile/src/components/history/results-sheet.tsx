import React from 'react';
import { FlatList, View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { spacing } from '../../theme';
import { downloadablesOf, isDownloadable, UploadTree } from '../../models/upload-tree';
import { formatBytes } from '../../utils/format';
import { SIconButton } from '../ui/s-icon-button';
import { SSheet } from '../ui/s-sheet';
import { SText } from '../ui/s-text';

interface Props {
  tree: UploadTree | null;
  busyId: string | null;
  onClose: () => void;
  onDownload: (sourceId: string) => void;
  onShare: (sourceId: string) => void;
}

/** Files of a folder upload: join results, or the items of a folder without join. */
export function ResultsSheet({ tree, busyId, onClose, onDownload, onShare }: Props) {
  const { t } = useTranslation();
  const files = tree ? downloadablesOf(tree) : [];

  return (
    <SSheet visible={!!tree} title={tree?.root.title ?? ''} onClose={onClose}>
      <FlatList
        data={files}
        keyExtractor={(f) => f.id}
        ItemSeparatorComponent={() => <View style={{ height: spacing.xs }} />}
        renderItem={({ item }) => {
          const ready = isDownloadable(item);
          return (
            <View style={{ flexDirection: 'row', alignItems: 'center', gap: spacing.sm }}>
              <View style={{ flex: 1, gap: spacing.xxs }}>
                <SText variant="label" numberOfLines={1}>
                  {item.filename || item.title}
                </SText>
                {item.status === 'failed' ? (
                  <SText variant="small" tone="danger" numberOfLines={2}>
                    {item.errorMsg
                      ? t('history.error', { error: item.errorMsg })
                      : t('history.errorUnknown')}
                  </SText>
                ) : (
                  <SText variant="small" tone="textSecondary">
                    {formatBytes(item.size)}
                  </SText>
                )}
              </View>
              {ready && (
                <>
                  <SIconButton
                    icon="download"
                    label={t('history.download')}
                    size={20}
                    disabled={busyId !== null}
                    onPress={() => onDownload(item.id)}
                  />
                  <SIconButton
                    icon="share"
                    label={t('history.share')}
                    size={20}
                    disabled={busyId !== null}
                    onPress={() => onShare(item.id)}
                  />
                </>
              )}
            </View>
          );
        }}
      />
    </SSheet>
  );
}
