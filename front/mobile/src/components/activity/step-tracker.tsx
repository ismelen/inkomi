import React from 'react';
import { View } from 'react-native';
import { useTranslation } from 'react-i18next';
import { colors, radii, spacing } from '../../theme';
import { ActivityStep, Tone } from '../../models/activity-view';
import { SIcon } from '../ui/s-icon';
import { SText } from '../ui/s-text';

interface Props {
  steps: ActivityStep[];
  tone: Tone;
}

const dot = 26;
const line = 2;

/** Numbered steps joined by a line: done ones get a check, the current one is filled. */
export function StepTracker({ steps, tone }: Props) {
  const { t } = useTranslation();
  const active = colors[tone];

  return (
    <View style={{ flexDirection: 'row' }}>
      {steps.map((step, i) => {
        const done = step.state === 'done';
        const current = step.state === 'active';
        const fill = done ? colors.success : current ? active : colors.background;
        const lineColor = (reached: boolean) => (reached ? colors.success : colors.border);
        return (
          <View key={step.key} style={{ flex: 1, alignItems: 'center', gap: spacing.xs }}>
            <View
              style={{
                height: dot,
                alignSelf: 'stretch',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              {i > 0 && (
                <View
                  style={{
                    position: 'absolute',
                    left: 0,
                    right: '50%',
                    height: line,
                    backgroundColor: lineColor(done || current),
                  }}
                />
              )}
              {i < steps.length - 1 && (
                <View
                  style={{
                    position: 'absolute',
                    left: '50%',
                    right: 0,
                    height: line,
                    backgroundColor: lineColor(done),
                  }}
                />
              )}
              <View
                style={{
                  width: dot,
                  height: dot,
                  borderRadius: radii.pill,
                  borderWidth: line,
                  borderColor: done ? colors.success : current ? active : colors.border,
                  backgroundColor: fill,
                  alignItems: 'center',
                  justifyContent: 'center',
                }}
              >
                {done ? (
                  <SIcon name="check" size={16} color={colors.onPrimary} />
                ) : (
                  <SText
                    variant="small"
                    style={{
                      color: current ? colors.onPrimary : colors.textSecondary,
                      fontFamily: 'semibold',
                    }}
                  >
                    {i + 1}
                  </SText>
                )}
              </View>
            </View>
            <SText
              variant="small"
              numberOfLines={1}
              adjustsFontSizeToFit
              tone={current ? 'text' : 'textSecondary'}
              style={current ? { fontFamily: 'semibold' } : undefined}
            >
              {t(`activity.steps.${step.key}`)}
            </SText>
          </View>
        );
      })}
    </View>
  );
}
