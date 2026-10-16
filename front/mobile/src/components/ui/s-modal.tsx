import React from 'react';
import { KeyboardAvoidingView, Modal, Platform, Pressable, View } from 'react-native';
import { colors, radii, spacing } from '../../theme';
import { SButton } from './s-button';
import { SText } from './s-text';

interface Action {
  label: string;
  onPress: () => void;
  variant?: 'primary' | 'secondary' | 'danger';
  disabled?: boolean;
  loading?: boolean;
}

interface Props {
  visible: boolean;
  title: string;
  message?: string;
  onClose: () => void;
  /** Rendered side by side at the bottom; the last one is the main action. */
  actions?: Action[];
  children?: React.ReactNode;
  /** The scrim doesn't close it (e.g. while sending). */
  persistent?: boolean;
}

/** Centered dialog over a scrim, as in the design. */
export function SModal({ visible, title, message, onClose, actions, children, persistent }: Props) {
  return (
    <Modal visible={visible} transparent animationType="fade" onRequestClose={onClose}>
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
        style={{ flex: 1 }}
      >
        <Pressable
          onPress={persistent ? undefined : onClose}
          style={{
            flex: 1,
            backgroundColor: colors.scrim,
            justifyContent: 'center',
            padding: spacing.xl,
          }}
        >
          <Pressable
            style={{
              backgroundColor: colors.background,
              borderRadius: radii.lg,
              padding: spacing.lg + 4,
              gap: spacing.md,
              maxHeight: '85%',
            }}
          >
            <SText variant="title" style={{ textAlign: 'center' }}>
              {title}
            </SText>
            {!!message && (
              <SText tone="textSecondary" style={{ textAlign: 'center' }}>
                {message}
              </SText>
            )}
            {children}
            {!!actions?.length && (
              <View style={{ flexDirection: 'row', gap: spacing.sm, marginTop: spacing.xs }}>
                {actions.map((action) => (
                  <SButton
                    key={action.label}
                    label={action.label}
                    onPress={action.onPress}
                    variant={action.variant ?? 'primary'}
                    disabled={action.disabled}
                    loading={action.loading}
                    style={{ flex: 1 }}
                  />
                ))}
              </View>
            )}
          </Pressable>
        </Pressable>
      </KeyboardAvoidingView>
    </Modal>
  );
}
