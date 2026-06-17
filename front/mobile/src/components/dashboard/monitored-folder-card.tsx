import { Pressable, StyleSheet, View } from 'react-native';
import { MonitoredFolder } from '../../models/monitored-folder';
import { colors } from '../../theme/colors';
import OpenFolderIcon from '../icons/open-folder-icon';
import SText from '../shared/SText';

interface Props {
  folder: MonitoredFolder;
  onPress(): void;
}

export default function MonitoredFolderCard({ folder, onPress }: Props) {
  return (
    <Pressable onPress={onPress} style={styles.folder}>
      <OpenFolderIcon size="20px" color={colors.onCard} />
      <SText style={{ flex: 1 }}>{folder.source.name}</SText>
      <View style={[styles.folder, { gap: 8 }]}>
        {folder.source.children?.length !== 0 && <View style={styles.dot} />}
        <SText
          style={{
            color: folder.source.children?.length === 0 ? colors.onCard : colors.primary,
          }}
        >
          {folder.source.children?.length} new
        </SText>
      </View>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  folder: {
    flexDirection: 'row',
    gap: 12,
    alignItems: 'center',
  },
  dot: {
    backgroundColor: colors.primary,
    width: 7,
    height: 7,
    borderRadius: 14,
  },
});
