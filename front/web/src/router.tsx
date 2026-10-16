import { createBrowserRouter, Navigate } from 'react-router';
import { AppLayout } from './components/layout/app-layout';
import { ActivityPage } from './pages/activity-page';
import { ListPage } from './pages/list-page';
import { LoginPage } from './pages/login-page';
import { OAuthDropboxPage } from './pages/oauth-dropbox-page';
import { SearchPage } from './pages/search-page';
import { SettingsPage } from './pages/settings-page';
import { dropboxCallbackPath } from './services/dropbox-oauth';
import { useSession } from './stores/use-session';

/** Signed out, every path shows the login; signed in, the shell with its sections. */
function SessionGate() {
  const status = useSession((s) => s.status);
  return status === 'signedIn' ? <AppLayout /> : <LoginPage />;
}

export const router = createBrowserRouter([
  { path: dropboxCallbackPath, element: <OAuthDropboxPage /> },
  {
    path: '/',
    element: <SessionGate />,
    children: [
      { index: true, element: <ListPage /> },
      { path: 'activity', element: <ActivityPage /> },
      { path: 'search', element: <SearchPage /> },
      { path: 'settings', element: <SettingsPage /> },
      { path: '*', element: <Navigate to="/" replace /> },
    ],
  },
]);
