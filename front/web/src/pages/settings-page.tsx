import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { appVersion } from '../config';
import { PageHeader } from '../components/layout/page-header';
import { defaultOptionsId, OptionsPanel } from '../components/options/options-panel';
import { Button } from '../components/ui/button';
import { Card } from '../components/ui/card';
import { cx } from '../components/ui/cx';
import { Select, TextInput } from '../components/ui/field';
import { Modal } from '../components/ui/modal';
import { Spinner } from '../components/ui/spinner';
import { Switch } from '../components/ui/switch';
import { supportedLanguages, SupportedLanguage } from '../i18n/i18n';
import { ereaders, findEReader } from '../models/ereaders';
import { notificationsSupported, requestNotifications } from '../services/notifications';
import { useProfile } from '../stores/use-profile';
import { useSession } from '../stores/use-session';
import { Theme, useSettings } from '../stores/use-settings';
import { showError, useToast } from '../stores/use-toast';

function Row({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex min-h-10 items-center gap-4">
      <span className="flex-1 text-base">{label}</span>
      {children}
    </div>
  );
}

export function SettingsPage() {
  const { t, i18n } = useTranslation();
  const signOut = useSession((s) => s.signOut);
  const { profile, loading, load, linkDropbox, unlinkDropbox, setDropboxFolder } = useProfile();
  const settings = useSettings();
  const ereader = findEReader(settings.ereaderKey);
  const [folder, setFolder] = useState(profile?.dropbox.folder ?? '');
  const [busy, setBusy] = useState<string | null>(null);
  const [mangaOpen, setMangaOpen] = useState(false);

  useEffect(() => {
    setFolder(profile?.dropbox.folder ?? '');
  }, [profile?.dropbox.folder]);

  const attempt = async (key: string, action: () => Promise<unknown>) => {
    setBusy(key);
    try {
      await action();
    } catch (err) {
      showError(err);
    } finally {
      setBusy(null);
    }
  };

  const saveFolder = () => {
    const next = folder.trim();
    if (!profile || next === profile.dropbox.folder) return;
    void attempt('folder', async () => {
      try {
        await setDropboxFolder(next);
        useToast.getState().show(t('history.saved'));
      } catch (err) {
        setFolder(profile.dropbox.folder);
        throw err;
      }
    });
  };

  const toggleNotifications = async (value: boolean) => {
    if (!value) return settings.setNotifications(false);
    if (!notificationsSupported()) {
      useToast.getState().show(t('settings.notificationsUnsupported'), 'danger');
    } else if (await requestNotifications()) {
      settings.setNotifications(true);
    } else {
      useToast.getState().show(t('settings.notificationsDenied'), 'danger');
    }
  };

  const providers = [
    profile?.hasPassword ? t('settings.providers.email') : null,
    ...(profile?.providers ?? []).map((p) => t(`settings.providers.${p}`)),
  ].filter(Boolean);
  const dropbox = profile?.dropbox;

  return (
    <>
      <PageHeader back title={t('settings.title')} />
      <div className="min-h-0 flex-1 overflow-y-auto px-4 pt-4 pb-6 md:px-8 md:pt-5">
        <div className="grid max-w-[1000px] grid-cols-[minmax(0,1fr)] items-start gap-4 lg:grid-cols-2 lg:gap-7">
          <div className="flex flex-col gap-4">
            <Card title={t('settings.account')}>
              <div className="flex items-center gap-4">
                <div className="min-w-0 flex-1">
                  <p className="truncate text-base font-medium">
                    {profile?.email ?? (loading ? '…' : t('settings.noEmail'))}
                  </p>
                  {providers.length > 0 && (
                    <p className="text-xs text-muted">
                      {t('settings.accountHint', { providers: providers.join(', ') })}
                    </p>
                  )}
                </div>
                <Button
                  variant="secondary"
                  size="sm"
                  loading={busy === 'signOut'}
                  onClick={() => void attempt('signOut', signOut)}
                >
                  {t('settings.signOut')}
                </Button>
              </div>
            </Card>

            <Card title={t('settings.reader')}>
              <Select
                aria-label={t('settings.reader')}
                value={settings.ereaderKey}
                onChange={(e) => settings.setEReader(e.target.value)}
                options={ereaders.map((e) => ({ value: e.key, label: e.name }))}
              />
              <p className="mt-3 text-xs text-muted">
                {t(ereader.kepub ? 'settings.readerKepub' : 'settings.readerSize', {
                  width: ereader.width,
                  height: ereader.height,
                })}
              </p>
            </Card>

            <Card title={t('settings.dropbox')}>
              {!profile ? (
                loading ? (
                  <Spinner />
                ) : (
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={() => void attempt('profile', load)}
                  >
                    {t('common.retry')}
                  </Button>
                )
              ) : dropbox?.linked ? (
                <div className="flex flex-col gap-4">
                  <div className="flex items-center gap-4">
                    <p className="flex min-w-0 flex-1 items-center gap-2 text-base">
                      <span className="size-2 shrink-0 rounded-full bg-success" />
                      <span className="truncate">
                        {dropbox.email
                          ? t('settings.connectedAs', { email: dropbox.email })
                          : t('settings.connected')}
                      </span>
                    </p>
                    <Button
                      variant="secondary"
                      size="sm"
                      loading={busy === 'unlink'}
                      onClick={() => void attempt('unlink', unlinkDropbox)}
                    >
                      {t('settings.disconnect')}
                    </Button>
                  </div>
                  <form
                    className="flex items-end gap-3"
                    onSubmit={(e) => {
                      e.preventDefault();
                      saveFolder();
                    }}
                  >
                    <TextInput
                      label={t('settings.folder')}
                      className="flex-1"
                      value={folder}
                      placeholder={t('settings.folderPlaceholder')}
                      spellCheck={false}
                      autoComplete="off"
                      onChange={(e) => setFolder(e.target.value)}
                      right={busy === 'folder' ? <Spinner /> : undefined}
                    />
                    <Button
                      type="submit"
                      variant="secondary"
                      disabled={folder.trim() === dropbox.folder || busy === 'folder'}
                    >
                      {t('settings.change')}
                    </Button>
                  </form>
                  <Row label={t('settings.toCloudDefault')}>
                    <Switch
                      label={t('settings.toCloudDefault')}
                      checked={settings.toCloudDefault}
                      onChange={settings.setToCloudDefault}
                    />
                  </Row>
                </div>
              ) : (
                <div className="flex items-center gap-4">
                  <p className="flex flex-1 items-center gap-2 text-base text-muted">
                    <span className="size-2 rounded-full bg-line" />
                    {t('settings.disconnected')}
                  </p>
                  <Button
                    size="sm"
                    icon="cloud"
                    loading={busy === 'link'}
                    onClick={() => void attempt('link', linkDropbox)}
                  >
                    {t('settings.connect')}
                  </Button>
                </div>
              )}
            </Card>
          </div>

          <div className="flex flex-col gap-4">
            <Card title={t('settings.preferences')}>
              <div className="flex flex-col gap-1">
                <Row label={t('settings.browserNotifications')}>
                  <Switch
                    label={t('settings.browserNotifications')}
                    checked={settings.notifications}
                    onChange={(v) => void toggleNotifications(v)}
                  />
                </Row>
                <Row label={t('settings.language')}>
                  <Select
                    aria-label={t('settings.language')}
                    className="w-40"
                    value={settings.language ?? i18n.language}
                    onChange={(e) => void settings.setLanguage(e.target.value as SupportedLanguage)}
                    options={supportedLanguages.map((lang) => ({
                      value: lang,
                      label: t(`languages.${lang}`),
                    }))}
                  />
                </Row>
                <Row label={t('settings.theme')}>
                  <Select
                    aria-label={t('settings.theme')}
                    className="w-40"
                    value={settings.theme}
                    onChange={(e) => settings.setTheme(e.target.value as Theme)}
                    options={(['system', 'light', 'dark'] as Theme[]).map((theme) => ({
                      value: theme,
                      label: t(`settings.themes.${theme}`),
                    }))}
                  />
                </Row>
              </div>
            </Card>

            <Card title={t('settings.defaultManga')}>
              <p className="mb-4 text-sm text-muted">{t('options.defaultHint')}</p>
              <Button variant="secondary" icon="tune" onClick={() => setMangaOpen(true)}>
                {t('options.edit')}
              </Button>
            </Card>
          </div>
        </div>
        <p className={cx('mt-8 text-center text-xs text-muted lg:max-w-[1000px]')}>
          {t('settings.version', { version: appVersion })}
        </p>
      </div>

      <Modal open={mangaOpen} onClose={() => setMangaOpen(false)} size="md" tall focusField={false}>
        {mangaOpen && (
          <div className="-mx-5 h-[min(720px,80dvh)] md:-mx-6">
            <OptionsPanel
              targetId={defaultOptionsId}
              onClose={() => setMangaOpen(false)}
              onSaved={() => setMangaOpen(false)}
            />
          </div>
        )}
      </Modal>
    </>
  );
}
