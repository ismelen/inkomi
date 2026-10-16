import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { pickBookFiles, pickBookFolder } from '../../services/file-picker';
import { useList } from '../../stores/use-list';
import { showError, useToast } from '../../stores/use-toast';
import { Button } from '../ui/button';
import { cx } from '../ui/cx';
import { Icon } from '../ui/icon';
import { IconButton } from '../ui/icon-button';
import { Menu, MenuItem } from '../ui/menu';
import { Md5Modal } from './md5-modal';

/**
 * "+ Añadir": files, a folder (a group) or a library book by md5. `compact` is the app's attach
 * button next to the search box (phones).
 */
export function AddMenu({ className, compact }: { className?: string; compact?: boolean }) {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);
  const [md5Open, setMd5Open] = useState(false);

  const addFiles = async () => {
    setOpen(false);
    try {
      useList.getState().addFiles(await pickBookFiles());
    } catch (err) {
      showError(err);
    }
  };

  const addFolder = async () => {
    setOpen(false);
    try {
      const folder = await pickBookFolder();
      if (!folder) return;
      if (folder.files.length === 0) useToast.getState().show(t('add.emptyFolder'), 'danger');
      else useList.getState().addFolder(folder);
    } catch (err) {
      showError(err);
    }
  };

  return (
    <div className={cx('relative', className)}>
      {compact ? (
        <IconButton
          icon="attach_file"
          label={t('add.title')}
          size={22}
          aria-haspopup="menu"
          aria-expanded={open}
          className="size-11 rounded-input bg-card text-fg shadow-card"
          onClick={() => setOpen((o) => !o)}
        />
      ) : (
        <Button
          variant="secondary"
          aria-haspopup="menu"
          aria-expanded={open}
          className="gap-3"
          onClick={() => setOpen((o) => !o)}
        >
          <Icon name="add" />
          {t('list.add')}
          <Icon name="expand_more" />
        </Button>
      )}
      <Menu open={open} onClose={() => setOpen(false)}>
        <MenuItem
          icon="attach_file"
          label={t('add.files')}
          hint={t('add.filesHint')}
          onSelect={() => void addFiles()}
        />
        <MenuItem
          icon="drive_file_move"
          label={t('add.folder')}
          hint={t('add.folderHint')}
          onSelect={() => void addFolder()}
        />
        <MenuItem
          icon="search"
          label={t('list.md5')}
          hint={t('list.md5Hint')}
          onSelect={() => {
            setOpen(false);
            setMd5Open(true);
          }}
        />
      </Menu>
      <Md5Modal open={md5Open} onClose={() => setMd5Open(false)} />
    </div>
  );
}
