import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { pickBookFiles, pickBookFolder } from '../../services/file-picker';
import { useList } from '../../stores/use-list';
import { showError, useToast } from '../../stores/use-toast';
import { Button } from '../ui/button';
import { cx } from '../ui/cx';
import { Icon } from '../ui/icon';
import { Menu, MenuItem } from '../ui/menu';
import { Md5Modal } from './md5-modal';

/** "+ Añadir": files, a folder (a group) or a library book by md5. */
export function AddMenu({ className }: { className?: string }) {
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
      <Button
        variant="secondary"
        aria-haspopup="menu"
        aria-expanded={open}
        className="w-full gap-3 md:w-auto"
        onClick={() => setOpen((o) => !o)}
      >
        + {t('list.add')}
        <Icon name="expand_more" />
      </Button>
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
