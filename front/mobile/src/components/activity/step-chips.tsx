import React from 'react';
import { View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { spacing } from '../../theme';
import { StepKey, stepOrder, StepState, Tone } from '../../models/activity-view';
import { SChip } from '../ui/s-chip';

interface Props {
  steps: Record<StepKey, StepState>;
  tone: Tone;
}

export function StepChips({ steps, tone }: Props) {
  const { t } = useTranslation();
  return (
    <View style={{ flexDirection: 'row', flexWrap: 'wrap', gap: spacing.xs }}>
      {stepOrder.map((step) => {
        const state = steps[step];
        return (
          <SChip
            key={step}
            label={t(`activity.steps.${step}`)}
            muted={state === 'skip'}
            filled={state === 'active'}
            tone={state === 'active' ? tone : state === 'done' ? 'success' : 'textSecondary'}
          />
        );
      })}
    </View>
  );
}
