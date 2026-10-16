import React, { useEffect, useMemo, useState } from 'react';
import { ActivityIndicator, Alert, SectionList, View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { errorKey } from '../../src/api/api-error';
import { HistoryRow } from '../../src/components/history/history-row';
import { ResultsSheet } from '../../src/components/history/results-sheet';
import { Header } from '../../src/components/ui/header';
import { Screen } from '../../src/components/ui/screen';
import { SText } from '../../src/components/ui/s-text';
import { treeFinishedAt, UploadTree } from '../../src/models/upload-tree';
import { saveSource, shareSource } from '../../src/services/download-service';
import { useActivity } from '../../src/stores/use-activity';
import { colors, spacing } from '../../src/theme';
import { dayBucket, DayBucket } from '../../src/utils/format';

const bucketOrder: DayBucket[] = ['today', 'yesterday', 'week', 'older'];

export default function HistoryScreen() {
  const { t } = useTranslation();
  const { history, historyEnd, loadingMore, loadMoreHistory } = useActivity();
  const [results, setResults] = useState<UploadTree | null>(null);
  const [busyId, setBusyId] = useState<string | null>(null);

  useEffect(() => {
    if (history.length === 0) void loadMoreHistory();
  }, []);

  const sections = useMemo(() => {
    const byBucket = new Map<DayBucket, UploadTree[]>();
    history.forEach((tree) => {
      const bucket = dayBucket(treeFinishedAt(tree));
      byBucket.set(bucket, [...(byBucket.get(bucket) ?? []), tree]);
    });
    return bucketOrder
      .filter((b) => byBucket.has(b))
      .map((b) => ({ key: b, title: t(`history.${b}`), data: byBucket.get(b)! }));
  }, [history, t]);

  // The presigned link is asked on every tap: it expires and must never be cached.
  const withDownload = async (sourceId: string, action: (id: string) => Promise<unknown>) => {
    setBusyId(sourceId);
    try {
      await action(sourceId);
    } catch (err) {
      Alert.alert(t('errors.title'), t(errorKey(err)));
    } finally {
      setBusyId(null);
    }
  };

  const download = (id: string) => withDownload(id, saveSource);
  const share = (id: string) => withDownload(id, shareSource);

  return (
    <Screen bottomInset>
      <Header
        title={t('history.title')}
        right={busyId ? <ActivityIndicator color={colors.primary} /> : undefined}
      />
      <SectionList
        sections={sections}
        keyExtractor={(tree) => tree.root.id}
        stickySectionHeadersEnabled={false}
        contentContainerStyle={{
          paddingHorizontal: spacing.md,
          paddingBottom: spacing.lg,
          flexGrow: 1,
        }}
        renderSectionHeader={({ section }) => (
          <SText
            variant="section"
            style={{
              paddingHorizontal: spacing.sm,
              paddingTop: spacing.lg,
              paddingBottom: spacing.xs,
            }}
          >
            {section.title}
          </SText>
        )}
        renderItem={({ item }) => (
          <HistoryRow
            tree={item}
            busy={busyId !== null}
            onDownload={download}
            onShare={share}
            onOpenResults={setResults}
          />
        )}
        onEndReached={() => void loadMoreHistory()}
        onEndReachedThreshold={0.5}
        ListFooterComponent={
          loadingMore && !historyEnd ? (
            <ActivityIndicator color={colors.primary} style={{ margin: spacing.lg }} />
          ) : null
        }
        ListEmptyComponent={
          loadingMore ? null : (
            <View style={{ flex: 1, alignItems: 'center', justifyContent: 'center' }}>
              <SText tone="textSecondary">{t('history.empty')}</SText>
            </View>
          )
        }
      />
      <ResultsSheet
        tree={results}
        busyId={busyId}
        onClose={() => setResults(null)}
        onDownload={download}
        onShare={share}
      />
    </Screen>
  );
}
