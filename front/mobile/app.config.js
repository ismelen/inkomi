import 'dotenv/config';

export default ({ config }) => ({
  ...config,
  extra: {
    backendUrl: process.env.BACKEND_URL,
    webClientId: process.env.WEB_CLIENT_ID,
    webClientSecret: process.env.WEB_CLIENT_SECRET,
  },
});
