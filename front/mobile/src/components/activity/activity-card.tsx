import React, { useState } from 'react';
import { Alert, View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { colors, radii, spacing } from '../../theme';
import { errorKey } from '../../api/api-error';
import { uploadsApi } from '../../api/uploads-api';
import { ActivityLine, activityView } from '../../models/activity-view';
import { unfinishedIds, UploadTree } from '../../models/upload-tree';
import { advanceStatus } from '../../services/activity-db';
import { SButton } from '../ui/s-button';
import { SModal } from '../ui/s-modal';
import { SText } from '../ui/s-text';
import { StepChips } from './step-chips';

function useLineText() {
  const { t } = useTranslation();
  return (line: ActivityLine) => {
    switch (line.key) {
      case 'processing':
        return line.total
          ? t('activity.line.processingCount', { done: line.done, total: line.total })
          : t('activity.line.processing');
      case 'waitingJoin':
        return t('activity.line.waitingJoin', { done: line.done, total: line.total });
      case 'failed':
        return t('activity.line.failed', { error: line.error });
      default:
        return t(`activity.line.${line.key}`);
    }
  };
}

export function ActivityCard({ tree }: { tree: UploadTree }) {
  const { t } = useTranslation();
  const lineText = useLineText();
  const view = activityView(tree);
  const toneColor = colors[view.tone];
  const [confirming, setConfirming] = useState(false);
  const [cancelling, setCancelling] = useState(false);

  const cancel = async () => {
    setConfirming(false);
    setCancelling(true);
    try {
      await uploadsApi.cancel(tree.root.id);
      // the websocket brings the same statuses; this shows them right away
      await advanceStatus(unfinishedIds(tree), 'cancelled');
    } catch (err) {
      Alert.alert(t('errors.title'), t(errorKey(err)));
    } finally {
      setCancelling(false);
    }
  };

  return (
    <View
      style={{
        borderWidth: 1,
        borderColor: colors.border,
        borderRadius: radii.lg,
        padding: spacing.md,
        gap: spacing.sm,
      }}
    >
      <View style={{ flexDirection: 'row', alignItems: 'center', gap: spacing.sm }}>
        <SText variant="label" numberOfLines={1} style={{ flex: 1 }}>
          {tree.root.title}
        </SText>
        {view.counter && (
          <SText variant="small" tone="textSecondary">
            {view.counter.done}/{view.counter.total}
          </SText>
        )}
      </View>
      <SText variant="small" style={{ color: toneColor }} numberOfLines={2}>
        {lineText(view.line)}
      </SText>
      {!!view.current && (
        <SText variant="small" tone="textSecondary" numberOfLines={1}>
          {view.current}
        </SText>
      )}
      <View style={{ height: 4, borderRadius: radii.pill, backgroundColor: colors.surface }}>
        <View
          style={{
            width: `${Math.round(view.progress * 100)}%`,
            height: 4,
            borderRadius: radii.pill,
            backgroundColor: toneColor,
          }}
        />
      </View>
      <View style={{ flexDirection: 'row', alignItems: 'center', gap: spacing.sm }}>
        <View style={{ flex: 1 }}>
          <StepChips steps={view.steps} tone={view.tone} />
        </View>
        <SButton
          variant="secondary"
          compact
          icon="close"
          label={t('activity.cancel')}
          loading={cancelling}
          onPress={() => setConfirming(true)}
        />
      </View>
      <SModal
        visible={confirming}
        title={t('activity.cancelTitle')}
        message={t('activity.cancelMessage')}
        onClose={() => setConfirming(false)}
        actions={[
          {
            label: t('activity.cancelKeep'),
            onPress: () => setConfirming(false),
            variant: 'secondary',
          },
          { label: t('activity.cancel'), onPress: () => void cancel(), variant: 'danger' },
        ]}
      />
    </View>
  );
}
