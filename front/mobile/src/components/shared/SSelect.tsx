import React, { useRef, useState } from 'react';
import { Animated, FlatList, Modal, Pressable, StyleSheet, Text, View } from 'react-native';
import { colors, hexToRgba } from '../../theme/colors';
import SIcon from '../icons/SIcon';

export interface SelectOption<T = string> {
  label: string;
  value: T;
}

interface Props<T = string> {
  label?: string;
  options: SelectOption<T>[];
  value?: T;
  onOptionChange: (option: SelectOption<T>) => void;
  placeholder?: string;
}

export default function SSelect<T = string>({
  label,
  options,
  value,
  onOptionChange,
  placeholder = 'Seleccionar...',
}: Props<T>) {
  const [open, setOpen] = useState(false);
  const fadeAnim = useRef(new Animated.Value(0)).current;
  const scaleAnim = useRef(new Animated.Value(0.95)).current;
  const flatListRef = useRef<FlatList<SelectOption<T>>>(null);

  // Altura de cada fila: paddingVertical(10*2) + fontSize(15) * lineHeight ≈ 41
  const ITEM_HEIGHT = 41;
  const SEPARATOR_HEIGHT = 1;
  const ROW_HEIGHT = ITEM_HEIGHT + SEPARATOR_HEIGHT;

  const selectedOption = options.find((o) => o.value === value);
  const selectedIndex = options.findIndex((o) => o.value === value);

  const openModal = () => {
    setOpen(true);
    Animated.parallel([
      Animated.timing(fadeAnim, {
        toValue: 1,
        duration: 180,
        useNativeDriver: true,
      }),
      Animated.spring(scaleAnim, {
        toValue: 1,
        tension: 180,
        friction: 12,
        useNativeDriver: true,
      }),
    ]).start();
  };

  const closeModal = () => {
    Animated.parallel([
      Animated.timing(fadeAnim, {
        toValue: 0,
        duration: 140,
        useNativeDriver: true,
      }),
      Animated.timing(scaleAnim, {
        toValue: 0.95,
        duration: 140,
        useNativeDriver: true,
      }),
    ]).start(() => setOpen(false));
  };

  const handleSelect = (option: SelectOption<T>) => {
    onOptionChange(option);
    closeModal();
  };

  return (
    <View>
      {label && <Text style={styles.label}>{label}</Text>}

      <Pressable
        onPress={openModal}
        android_ripple={{ color: hexToRgba(colors.primary, 0.08), foreground: true }}
        style={[styles.trigger, open && styles.triggerPressed]}
      >
        <Text
          style={[styles.triggerText, !selectedOption && styles.triggerPlaceholder]}
          numberOfLines={1}
        >
          {selectedOption ? selectedOption.label : placeholder}
        </Text>

        <View style={styles.chevronWrapper}>
          <SIcon
            name={open ? 'arrow_up' : 'arrow_down'}
            color={colors.primary}
            size={30}
            type="outlined"
          />
        </View>
      </Pressable>

      <Modal
        visible={open}
        transparent
        animationType="none"
        onRequestClose={closeModal}
        statusBarTranslucent
      >
        <Pressable style={styles.backdrop} onPress={closeModal}>
          <Animated.View
            style={[styles.sheet, { opacity: fadeAnim, transform: [{ scale: scaleAnim }] }]}
          >
            {label && <Text style={styles.sheetTitle}>{label}</Text>}

            <FlatList
              ref={flatListRef}
              data={options}
              onLayout={() => {
                if (selectedIndex > 0) {
                  flatListRef.current?.scrollToOffset({
                    offset: ROW_HEIGHT * selectedIndex,
                    animated: false,
                  });
                }
              }}
              getItemLayout={(_, index) => ({
                length: ROW_HEIGHT,
                offset: ROW_HEIGHT * index,
                index,
              })}
              keyExtractor={(_, i) => String(i)}
              bounces={false}
              ItemSeparatorComponent={() => (
                <View
                  style={{ height: 1, backgroundColor: colors.outline_variant, opacity: 0.3 }}
                />
              )}
              renderItem={({ item }) => {
                const isSelected = item.value === value;
                return (
                  <Pressable
                    onPress={() => handleSelect(item)}
                    android_ripple={{
                      color: hexToRgba(colors.primary, 0.1),
                      foreground: true,
                    }}
                    style={[styles.option, isSelected && styles.optionSelected]}
                  >
                    <Text style={[styles.optionText, isSelected && styles.optionTextSelected]}>
                      {item.label}
                    </Text>
                    {isSelected && (
                      <SIcon name="check" size={20} color={colors.primary} type="outlined" />
                    )}
                  </Pressable>
                );
              }}
            />
          </Animated.View>
        </Pressable>
      </Modal>
    </View>
  );
}

const styles = StyleSheet.create({
  label: {
    fontFamily: 'regular',
    fontSize: 13,
    color: colors.on_surface_variant,
    marginBottom: 6,
    marginLeft: 2,
  },

  // ── Trigger ──────────────────────────────────────────────────────────────
  trigger: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: colors.surface_container_lowest,
    boxShadow: colors.boxShadow,
    borderRadius: 12,
    paddingHorizontal: 14,
    paddingVertical: 13,
    overflow: 'hidden',
  },
  triggerPressed: {
    borderColor: colors.primary,
    backgroundColor: colors.surface_container,
  },
  triggerText: {
    flex: 1,
    fontFamily: 'regular',
    fontSize: 15,
    color: colors.on_surface,
  },
  triggerPlaceholder: {
    color: colors.outline,
  },

  // ── Chevron ───────────────────────────────────────────────────────────────
  chevronWrapper: {
    width: 16,
    height: 10,
    justifyContent: 'center',
    alignItems: 'center',
    marginLeft: 8,
  },
  chevronLine: {
    position: 'absolute',
    width: 9,
    height: 2,
    borderRadius: 1,
    backgroundColor: colors.outline,
  },
  chevronLeft: {
    transform: [{ rotate: '45deg' }, { translateX: -3 }],
  },
  chevronRight: {
    transform: [{ rotate: '-45deg' }, { translateX: 3 }],
  },

  // ── Modal sheet ───────────────────────────────────────────────────────────
  backdrop: {
    flex: 1,
    backgroundColor: hexToRgba('#131b2e', 0.4),
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: 24,
  },
  sheet: {
    width: '100%',
    maxHeight: 400,
    backgroundColor: colors.surface_container_lowest,
    borderRadius: 20,
    overflow: 'hidden',
    paddingVertical: 8,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 8 },
    shadowOpacity: 0.12,
    shadowRadius: 24,
    elevation: 12,
  },
  sheetTitle: {
    fontFamily: 'regular',
    fontSize: 12,
    color: colors.on_surface_variant,
    textTransform: 'uppercase',
    letterSpacing: 1,
    paddingHorizontal: 16,
    paddingTop: 8,
    paddingBottom: 12,
  },

  // ── Options ───────────────────────────────────────────────────────────────
  option: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 10,
    overflow: 'hidden',
  },
  optionSelected: {
    backgroundColor: hexToRgba(colors.primary, 0.07),
  },
  optionText: {
    flex: 1,
    fontFamily: 'regular',
    fontSize: 15,
    color: colors.on_surface,
  },
  optionTextSelected: {
    color: colors.primary,
    fontFamily: 'regular',
  },

  // ── Checkmark (two lines forming a tick) ──────────────────────────────────
  checkWrapper: {
    width: 18,
    height: 18,
    justifyContent: 'center',
    alignItems: 'center',
  },
  checkShort: {
    position: 'absolute',
    width: 5,
    height: 2,
    borderRadius: 1,
    backgroundColor: colors.primary,
    transform: [{ rotate: '45deg' }, { translateX: -4 }, { translateY: 2 }],
  },
  checkLong: {
    position: 'absolute',
    width: 10,
    height: 2,
    borderRadius: 1,
    backgroundColor: colors.primary,
    transform: [{ rotate: '-50deg' }, { translateX: 2 }],
  },
});
