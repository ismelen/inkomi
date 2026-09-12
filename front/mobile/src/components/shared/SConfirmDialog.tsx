import React from 'react';
import { Modal, View, StyleSheet } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { colors, hexToRgba } from '../../theme/colors';
import SButton from './SButton';
import SText from './SText';

interface Props {
  visible: boolean;
  title: string;
  message: string;
  onConfirm: () => void;
  onCancel?: () => void;
  confirmText?: string;
  cancelText?: string;
}

export default function SConfirmDialog({
  visible,
  title,
  message,
  onConfirm,
  onCancel,
  confirmText = 'OK',
  cancelText = 'Cancel',
}: Props) {
  const insets = useSafeAreaInsets();

  return (
    <Modal transparent={true} visible={visible} animationType="fade" onRequestClose={onCancel}>
      <View
        style={[styles.overlay, { paddingTop: 24 + insets.top, paddingBottom: 24 + insets.bottom }]}
      >
        <View style={styles.modalContainer}>
          <SText style={styles.title}>{title}</SText>
          <SText style={styles.message}>{message}</SText>
          <View style={styles.buttonContainer}>
            {onCancel && (
              <SButton style={styles.cancelButton} onPress={onCancel}>
                <SText style={styles.cancelButtonText}>{cancelText}</SText>
              </SButton>
            )}
            <SButton style={styles.confirmButton} onPress={onConfirm}>
              <SText style={styles.confirmButtonText}>{confirmText}</SText>
            </SButton>
          </View>
        </View>
      </View>
    </Modal>
  );
}

const styles = StyleSheet.create({
  overlay: {
    flex: 1,
    backgroundColor: hexToRgba(colors.on_surface, 0.5),
    justifyContent: 'center',
    alignItems: 'center',
    padding: 24,
  },
  modalContainer: {
    width: '100%',
    backgroundColor: colors.surface,
    borderRadius: 24,
    padding: 24,
    elevation: 8,
    shadowColor: colors.on_surface,
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.15,
    shadowRadius: 12,
  },
  title: {
    fontSize: 22,
    fontFamily: 'semibold',
    color: colors.on_surface,
    marginBottom: 12,
  },
  message: {
    fontSize: 16,
    color: colors.on_surface_variant,
    marginBottom: 24,
    lineHeight: 24,
  },
  buttonContainer: {
    flexDirection: 'row',
    justifyContent: 'flex-end',
    gap: 8,
  },
  cancelButton: {
    paddingVertical: 10,
    paddingHorizontal: 16,
    borderRadius: 12,
    justifyContent: 'center',
    alignItems: 'center',
  },
  cancelButtonText: {
    color: colors.primary,
    fontSize: 16,
    fontFamily: 'semibold',
  },
  confirmButton: {
    backgroundColor: colors.primary,
    paddingVertical: 10,
    paddingHorizontal: 20,
    borderRadius: 12,
    justifyContent: 'center',
    alignItems: 'center',
  },
  confirmButtonText: {
    color: colors.on_primary,
    fontSize: 16,
    fontFamily: 'semibold',
  },
});
