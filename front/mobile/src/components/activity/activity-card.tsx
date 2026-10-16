import React from 'react';
import { View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { colors, radii, spacing } from '../../theme';
import { ActivityLine, activityView } from '../../models/activity-view';
import { UploadTree } from '../../models/upload-tree';
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
  const lineText = useLineText();
  const view = activityView(tree);
  const toneColor = colors[view.tone];

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
      <StepChips steps={view.steps} tone={view.tone} />
    </View>
  );
}
