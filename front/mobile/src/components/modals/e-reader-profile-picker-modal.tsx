import { Pressable, useWindowDimensions, View } from 'react-native';
import { ScrollView } from 'react-native-gesture-handler';
import { create } from 'zustand';
import { eReaderProfiles } from '../../constants';
import { colors } from '../../theme/colors';
import ChevronRightIcon from '../icons/chevron-right-icon';
import CloseIcon from '../icons/close-icon';
import SColumn from '../shared/SColumn';
import SDivider from '../shared/SDivider';
import SText from '../shared/SText';

interface State {
  resolve?: (value?: string) => void;
  showing: boolean;
  show(): Promise<string | undefined>;
  close(model?: string): void;
}

export const useEReaderModelPicker = create<State>((set, get) => ({
  showing: false,

  show(): Promise<string | undefined> {
    return new Promise<string | undefined>((resolve) => {
      set({ resolve: resolve, showing: true });
    });
  },

  close(model?: string) {
    const resolve = get().resolve;
    set({ showing: false });

    resolve?.(model);
  },
}));

export default function EReaderProfilePickerModalRoot() {
  const showing = useEReaderModelPicker((s) => s.showing);
  const close = useEReaderModelPicker((s) => s.close);
  const { height, width } = useWindowDimensions();

  const entries = Object.entries(eReaderProfiles);

  if (!showing) return null;

  return (
    <View
      style={{ backgroundColor: colors.background, width: width, height: height, paddingTop: 50 }}
    >
      <View
        style={{
          paddingBottom: 15,
          paddingHorizontal: 20,
          flexDirection: 'row',
          justifyContent: 'space-between',
        }}
      >
        <SText style={{ fontWeight: 600, fontSize: 20 }}>Select Device</SText>
        <Pressable onPress={() => close()}>
          <CloseIcon size="28px" color={colors.onCard} />
        </Pressable>
      </View>
      <SDivider />
      <ScrollView
        style={{
          paddingHorizontal: 20,
          paddingTop: 10,
        }}
      >
        <SColumn>
          {entries.map(([model, name]) => (
            <Pressable
              key={model}
              style={{ flexDirection: 'row', justifyContent: 'space-between' }}
              onPress={() => close(model)}
            >
              <SText style={{ fontWeight: 500, fontSize: 16 }}>{name}</SText>
              <ChevronRightIcon size="26px" color={colors.onCard} />
            </Pressable>
          ))}
        </SColumn>
        <View style={{ height: 100 }} />
      </ScrollView>
    </View>
  );
}
