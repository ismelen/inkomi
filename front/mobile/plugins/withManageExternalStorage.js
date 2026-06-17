const { withAndroidManifest } = require('@expo/config-plugins');

const withManageExternalStorage = (config) => {
  return withAndroidManifest(config, async (config) => {
    const androidManifest = config.modResults.manifest;

    // Agregar el permiso MANAGE_EXTERNAL_STORAGE
    if (!androidManifest['uses-permission']) {
      androidManifest['uses-permission'] = [];
    }

    const manageStoragePermission = {
      $: {
        'android:name': 'android.permission.MANAGE_EXTERNAL_STORAGE',
      },
    };

    // Verificar si ya existe
    const hasPermission = androidManifest['uses-permission'].some(
      (permission) => permission.$['android:name'] === 'android.permission.MANAGE_EXTERNAL_STORAGE'
    );

    if (!hasPermission) {
      androidManifest['uses-permission'].push(manageStoragePermission);
    }

    return config;
  });
};

module.exports = withManageExternalStorage;
