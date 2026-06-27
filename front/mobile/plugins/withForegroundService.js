const { withAndroidManifest } = require('@expo/config-plugins');

module.exports = (config) =>
  withAndroidManifest(config, (config) => {
    const manifest = config.modResults.manifest;

    // Añadir namespace tools si no existe
    if (!manifest.$['xmlns:tools']) {
      manifest.$['xmlns:tools'] = 'http://schemas.android.com/tools';
    }

    const app = manifest.application[0];
    app.service = app.service ?? [];

    const serviceName = 'com.asterinet.react.bgactions.RNBackgroundActionsTask';

    // Buscar si ya existe el service declarado por la librería
    const existing = app.service.find(
      (s) => s.$['android:name'] === serviceName
    );

    if (existing) {
      // Modificar el existente
      existing.$['android:foregroundServiceType'] = 'dataSync';
      existing.$['tools:replace'] = 'android:foregroundServiceType';
    } else {
      // Añadir nuevo
      app.service.push({
        $: {
          'android:name': serviceName,
          'android:foregroundServiceType': 'dataSync',
          'tools:replace': 'android:foregroundServiceType',
        },
      });
    }

    return config;
  });