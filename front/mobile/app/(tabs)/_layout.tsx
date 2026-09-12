import { Tabs } from 'expo-router';
import React from 'react';
import { colors } from '../../src/theme/colors';
import SIcon from '../../src/components/icons/SIcon';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useTranslation } from 'react-i18next';

export default function TabsLayout() {
  const insets = useSafeAreaInsets();
  const { t } = useTranslation();

  return (
    <Tabs
      screenOptions={{
        sceneStyle: {
          backgroundColor: colors.background,
          paddingTop: insets.top,
        },
        headerShadowVisible: false,
        headerShown: false,

        headerStyle: {
          backgroundColor: colors.background,
          elevation: 0,
          borderBottomWidth: 0,
        },

        tabBarActiveBackgroundColor: colors.secondary_container,
        tabBarInactiveBackgroundColor: 'transparent',

        tabBarItemStyle: {
          overflow: 'hidden',
          borderRadius: 12,
          marginHorizontal: 15,
          marginVertical: 8,
          maxWidth: 80,
        },

        tabBarStyle: {
          alignItems: 'center',
          justifyContent: 'space-evenly',
          height: 70 + insets.bottom,
          paddingBottom: insets.bottom,
          backgroundColor: colors.background,
          borderTopWidth: 0,
          elevation: 0,
          shadowOpacity: 0,
        },

        tabBarActiveTintColor: colors.on_secondary_container,
        tabBarInactiveTintColor: colors.on_surface,

        tabBarLabelStyle: {
          fontSize: 12,
          fontWeight: '900',
        },
      }}
    >
      <Tabs.Screen
        name="index"
        options={{
          title: t('tabs.home'),
          tabBarIcon: ({ color, size, focused }) => (
            <SIcon color={color} size={size} name="home" type={focused ? 'filled' : 'outlined'} />
          ),
        }}
      />
      <Tabs.Screen
        name="queue"
        options={{
          title: t('tabs.queue'),
          tabBarIcon: ({ color, size, focused }) => (
            <SIcon
              color={color}
              size={size}
              name="database"
              type={focused ? 'filled' : 'outlined'}
            />
          ),
        }}
      />
      <Tabs.Screen
        name="search"
        options={{
          title: t('tabs.search'),
          tabBarIcon: ({ color, size }) => (
            <SIcon color={color} size={size} name="search" type="outlined" />
          ),
        }}
      />
      <Tabs.Screen
        name="settings"
        options={{
          title: t('tabs.settings'),
          tabBarIcon: ({ color, size, focused }) => (
            <SIcon
              color={color}
              size={size}
              name="settings"
              type={focused ? 'filled' : 'outlined'}
            />
          ),
        }}
      />
    </Tabs>
  );
}
