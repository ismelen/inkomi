import { Tabs } from 'expo-router';
import DatabaseIcon from '../../src/components/icons/databse-icon';
import HomeIcon from '../../src/components/icons/home-icon';
import SettingsIcon from '../../src/components/icons/settings-icon';
import { colors } from '../../src/theme/colors';

export default function TabsLayout() {
  return (
    <Tabs
      screenOptions={{
        headerShown: false,
        headerStyle: {
          backgroundColor: colors.background,
        },
        headerShadowVisible: false,
        sceneStyle: {
          paddingTop: 50,
          backgroundColor: colors.background,
        },
        tabBarActiveTintColor: colors.primary,
        tabBarInactiveTintColor: colors.onCard,
        tabBarStyle: {
          backgroundColor: colors.background,
          opacity: 90,
          paddingTop: 10,
          height: 70,
          borderTopColor: colors.gray,
          borderTopWidth: 1,
        },
        tabBarLabelStyle: {
          fontSize: 10,
          fontWeight: 'medium',
        },
      }}
    >
      <Tabs.Screen
        name="index"
        options={{
          title: 'Home',
          tabBarIcon: ({ color, size }) => <HomeIcon size={`${size}px`} color={color} />,
        }}
      />
      <Tabs.Screen
        name="queue"
        options={{
          title: 'Queue',
          tabBarIcon: ({ color, size }) => <DatabaseIcon size={`${size}px`} color={color} />,
        }}
      />
      <Tabs.Screen
        name="settings"
        options={{
          title: 'Settings',
          tabBarIcon: ({ color, size }) => <SettingsIcon size={`${size}px`} color={color} />,
        }}
      />
    </Tabs>
  );
}
