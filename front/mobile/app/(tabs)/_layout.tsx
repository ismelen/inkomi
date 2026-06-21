import { Tabs } from 'expo-router';
import React from 'react';
import { colors } from '../../src/theme/colors';
import SIcon from '../../src/componentes/icons/SIcon';
import { Image, Text, View } from 'react-native';
import logo from '../../assets/icon.png';

export default function TabsLayout() {
  return (
    <Tabs
      screenOptions={{
        sceneStyle: {
          backgroundColor: colors.background,
        },
        headerShadowVisible: false,

        header: () => (
          <View
            style={{
              alignItems: 'center',
              justifyContent: 'center',
              flexDirection: 'row',
              gap: '8',
            }}
          >
            <View style={{}}>
              <Image source={logo} style={{ width: 38, height: 38 }} resizeMode="contain" />
            </View>
            <Text
              style={{
                fontSize: 32,
                color: colors.primary,
                fontFamily: 'Inter_700Bold',
              }}
            >
              Inkomi
            </Text>
          </View>
        ),

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
          height: 70,
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
          title: 'Home',
          tabBarIcon: ({ color, size, focused }) => (
            <SIcon color={color} size={size} name="home" type={focused ? 'filled' : 'outlined'} />
          ),
        }}
      />
      <Tabs.Screen
        name="queue"
        options={{
          title: 'Queue',
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
        name="settings"
        options={{
          title: 'Settings',
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
