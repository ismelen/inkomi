import React, { useCallback } from 'react';
import { ScrollView, StyleSheet, View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { useFocusEffect } from 'expo-router';
import { useShallow } from 'zustand/react/shallow';
import { colors } from '../../theme/colors';
import { Transaction } from '../../models/transaction';
import { useQueue } from '../../hooks/useQueue';

import SText from '../shared/SText';

import { CardHeader } from './card-header';
import { UploadRow } from './upload-row';
import { ItemRow } from './item-row';
import { ResultRow } from './result-row';
import {
  CancelButton,
  DownloadAllButton,
  RedoButton,
  ShareAllButton,
  StartUploadsButton,
} from './queue-buttons';
import { useObjectNavigation } from '../../hooks/useObjectNavigation';
import { TransactionMode } from '../../models/transaction-config';

const SECTION_MAX_HEIGHT = 160;

function getPath(mode: TransactionMode): string {
  switch (mode) {
    case 'cbz':
      return '/send-comic';
    case 'epub':
      return '/send-book';
    case 'md5':
      return '/send-libgen';
  }
}

interface Props {
  data: Transaction;
  idx: number;
}

export default React.memo(QueueItemCard);

function QueueItemCard({ data, idx }: Props) {
  const hasUploads = data.uploads.length > 0;
  const hasItems = (data.items?.length ?? 0) > 0;
  const hasResults = (data.results?.length ?? 0) > 0;
  const isError = data.status === 'error';
  const isDone = data.status === 'done';
  const isCancellable = data.status === 'processing' || data.status === 'waiting';

  const navigate = useObjectNavigation((s) => s.navigate);
  const isFinished =
    data.status === 'done' || data.status === 'error' || data.status === 'canceled';

  const { checkProgress, cancel, startUploads, retryUpload, retryItem } = useQueue(
    useShallow((s) => ({
      checkProgress: s.checkProgress,
      cancel: s.cancel,
      startUploads: s.startUploads,
      retryUpload: s.retryUpload,
      retryItem: s.retryItem,
    }))
  );

  useFocusEffect(
    useCallback(() => {
      const interval = setInterval(() => checkProgress(idx), 2000);
      return () => clearInterval(interval);
    }, [idx, checkProgress])
  );

  const handleRetryUpload = (uploadIdx: number, newFile?: any) => {
    retryUpload(idx, uploadIdx, newFile);
  };

  const handleRetryItem = (itemId: string) => {
    retryItem(idx, itemId);
  };

  const handleRedo = () => {
    const path = getPath(data.config.mode);
    navigate(path, data.config);
  };

  const canStartUploads =
    data.status === 'waiting' && data.uploads.some((u) => u.status === 'pending');

  return (
    <View style={[s.card, isError && s.cardError]}>
      {/* Header */}
      <CardHeader data={data} />

      {/* Uploads */}
      {hasUploads && (
        <Section sectionKey="queue.uploads">
          {data.uploads.map((u, i) => (
            <UploadRow
              key={i}
              upload={u}
              idx={i}
              onRetry={handleRetryUpload}
              allowedTypes={
                data.config.mode === 'cbz'
                  ? ['application/pdf', 'application/x-cbz', 'application/zip', '.cbz', '.pdf']
                  : undefined
              }
            />
          ))}
        </Section>
      )}

      {/* Items */}
      {hasItems && (
        <Section sectionKey="queue.items">
          {data.items.map((item) => (
            <ItemRow key={item.id} item={item} onRetry={handleRetryItem} />
          ))}
        </Section>
      )}

      {/* Results */}
      {hasResults && (
        <Section sectionKey="queue.results">
          <View style={{ gap: 4, flexDirection: 'column' }}>
            {data.results.map((result) => (
              <ResultRow key={result.id} result={result} tran={data} idx={idx} />
            ))}
          </View>
        </Section>
      )}

      {/* Download All & Share buttons */}
      {!data.config.toCloud && isDone && hasResults && (
        <View style={{ flexDirection: 'row', gap: 8 }}>
          <View style={{ flex: 1 }}>
            <DownloadAllButton tran={data} idx={idx} />
          </View>
          <ShareAllButton tran={data} />
        </View>
      )}

      {/* Start Uploads button */}
      {canStartUploads && <StartUploadsButton onPress={() => startUploads?.(idx)} />}

      {/* Redo button */}
      {isFinished && <RedoButton onPress={handleRedo} />}

      {/* Cancel button */}
      {isCancellable && <CancelButton onPress={() => cancel(idx)} />}
    </View>
  );
}

function Section({ sectionKey, children }: { sectionKey: string; children: React.ReactNode }) {
  const { t } = useTranslation();
  return (
    <View style={{ gap: 6 }}>
      <SText style={s.sectionTitle}>{t(sectionKey)}</SText>
      <ScrollView style={{ maxHeight: SECTION_MAX_HEIGHT }} nestedScrollEnabled>
        {children}
      </ScrollView>
    </View>
  );
}

const s = StyleSheet.create({
  card: {
    backgroundColor: colors.surface_container_lowest,
    borderRadius: 16,
    boxShadow: colors.boxShadow,
    padding: 16,
    gap: 14,
  },
  cardError: {
    borderColor: colors.error,
    borderWidth: 1,
  },
  sectionTitle: {
    fontSize: 13,
    fontFamily: 'medium',
    color: colors.on_surface_variant,
  },
});
