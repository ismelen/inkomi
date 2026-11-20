import React, { useState } from 'react';
import { Alert, View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { colors, radii, spacing } from '../../theme';
import { errorKey } from '../../api/api-error';
import { uploadsApi } from '../../api/uploads-api';
import { ActivityLine, activityView } from '../../models/activity-view';
import { unfinishedIds, UploadTree } from '../../models/upload-tree';
import { advanceStatus } from '../../services/activity-db';
import { IconName } from '../icons/icons';
import { SButton } from '../ui/s-button';
import { SIcon } from '../ui/s-icon';
import { SModal } from '../ui/s-modal';
import { SText } from '../ui/s-text';
import { StepTracker } from './step-tracker';

const lineIcon: Record<ActivityLine['key'], IconName> = {
  pendingUpload: 'attach_file',
  queued: 'progress_activity',
  downloading: 'download',
  processing: 'menu_book',
  waitingJoin: 'checklist',
  joining: 'checklist',
  kepub: 'tune',
  uploading: 'check_circle',
  failed: 'error',
  cancelled: 'close',
};

export function ActivityCard({ tree }: { tree: UploadTree }) {
  const { t } = useTranslation();
  const view = activityView(tree);
  const toneColor = colors[view.tone];
  const cancellable = unfinishedIds(tree).length > 0;
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

  const step = view.stepIndex >= 0 ? view.steps[view.stepIndex] : undefined;

  return (
    <View
      style={{
        borderWidth: 1,
        borderColor: colors.border,
        borderRadius: radii.lg,
        padding: spacing.lg,
        gap: spacing.md,
      }}
    >
      <SText variant="title" numberOfLines={2}>
        {tree.root.title}
      </SText>

      <View style={{ flexDirection: 'row', alignItems: 'center', gap: spacing.sm }}>
        <SIcon name={lineIcon[view.line.key]} size={22} color={toneColor} />
        <SText variant="label" style={{ flex: 1, color: toneColor }}>
          {t(`activity.line.${view.line.key}`)}
        </SText>
      </View>
      {view.line.key === 'failed' && (
        <SText variant="small" tone="textSecondary">
          {t('activity.failedHint')}
        </SText>
      )}

      {view.steps.length > 1 && (
        <View style={{ gap: spacing.sm }}>
          {step && (
            <SText variant="small" tone="textSecondary">
              {t('activity.stepOf', {
                current: view.stepIndex + 1,
                total: view.steps.length,
                step: t(`activity.steps.${step.key}`),
              })}
            </SText>
          )}
          <StepTracker steps={view.steps} tone={view.tone} />
        </View>
      )}

      {view.counter && view.counter.total > 0 && (
        <View style={{ gap: spacing.xs }}>
          <SText variant="small" tone="textSecondary">
            {t('activity.booksReady', {
              count: view.counter.total,
              done: view.counter.done,
            })}
          </SText>
          <View
            accessibilityRole="progressbar"
            accessibilityValue={{ min: 0, max: view.counter.total, now: view.counter.done }}
            style={{ height: 8, borderRadius: radii.pill, backgroundColor: colors.surface }}
          >
            <View
              style={{
                width: `${Math.round((view.counter.done / view.counter.total) * 100)}%`,
                height: 8,
                borderRadius: radii.pill,
                backgroundColor: colors.success,
              }}
            />
          </View>
          {!!view.current && (
            <SText variant="small" tone="textSecondary" numberOfLines={2}>
              {t('activity.now', { titles: view.current })}
            </SText>
          )}
        </View>
      )}

      {cancellable && (
        <SButton
          variant="dangerOutline"
          icon="close"
          label={t('activity.cancel')}
          loading={cancelling}
          onPress={() => setConfirming(true)}
          style={{ alignSelf: 'flex-end' }}
        />
      )}

      <SModal
        visible={confirming}
        title={t('activity.cancelTitle')}
        message={t('activity.cancelMessage', { title: tree.root.title })}
        onClose={() => setConfirming(false)}
        actions={[
          {
            label: t('activity.cancelKeep'),
            onPress: () => setConfirming(false),
            variant: 'secondary',
          },
          {
            label: t('activity.cancelConfirm'),
            onPress: () => void cancel(),
            variant: 'danger',
          },
        ]}
      />
    </View>
  );
}
