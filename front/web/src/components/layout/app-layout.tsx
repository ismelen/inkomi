import { useEffect } from 'react';
import { Outlet } from 'react-router';
import { startSync, stopSync } from '../../services/sync-service';
import { useActivity } from '../../stores/use-activity';
import { useProfile } from '../../stores/use-profile';
import { SendingModal } from '../upload/sending-modal';
import { Sidebar } from './nav';

/** Mounted only with a session: realtime activity and the profile live here. */
export function AppLayout() {
  useEffect(() => {
    void useActivity.getState().reload();
    startSync();
    useProfile
      .getState()
      .load()
      .catch((err) => console.warn('profile', err));
    return () => stopSync();
  }, []);

  return (
    <div className="flex h-dvh">
      <Sidebar />
      <div className="flex min-w-0 flex-1 flex-col">
        <main className="flex min-h-0 flex-1 flex-col">
          <Outlet />
        </main>
      </div>
      <SendingModal />
    </div>
  );
}
