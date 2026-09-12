import React from 'react';
import { Modal, View, StyleSheet } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useVersionChecker } from '../../hooks/useVersionhecker';
import { useShallow } from 'zustand/react/shallow';
import { colors, hexToRgba } from '../../theme/colors';
import SButton from '../shared/SButton';
import SText from '../shared/SText';

export default function NewVersionModal() {
  const insets = useSafeAreaInsets();
  const { install, show } = useVersionChecker(
    useShallow((s) => ({
      install: s.installNewVersion,
      show: s.showDialog,
    }))
  );

  return (
    <Modal transparent={true} visible={show} animationType="fade">
      <View
        style={[styles.overlay, { paddingTop: 24 + insets.top, paddingBottom: 24 + insets.bottom }]}
      >
        <View style={styles.modalContainer}>
          <SText style={styles.title}>Nueva versión disponible</SText>
          <SText style={styles.message}>
            Hay una nueva versión de la aplicación. Por favor, actualiza para disfrutar de las
            últimas mejoras y correcciones.
          </SText>
          <View style={styles.buttonContainer}>
            <SButton style={styles.installButton} onPress={install}>
              <SText style={styles.installButtonText}>Instalar</SText>
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
    fontWeight: '700',
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
    fontWeight: '600',
  },
  installButton: {
    backgroundColor: colors.primary,
    paddingVertical: 10,
    paddingHorizontal: 20,
    borderRadius: 12,
    justifyContent: 'center',
    alignItems: 'center',
  },
  installButtonText: {
    color: colors.on_primary,
    fontSize: 16,
    fontWeight: '600',
  },
});
