// Copyright 2017 The Free Chess Club.

import { app, BrowserWindow, dialog, Menu, session, shell } from 'electron';
import * as Electron from 'electron';
import * as path from 'path';
import * as url from 'url';

let mainWindow = null;

const template = [{
  label: 'Edit',
  submenu: [{
    label: 'Undo',
    accelerator: 'CmdOrCtrl+Z',
    role: 'undo',
  }, {
    label: 'Redo',
    accelerator: 'Shift+CmdOrCtrl+Z',
    role: 'redo',
  }, {
    type: 'separator',
  }, {
    label: 'Cut',
    accelerator: 'CmdOrCtrl+X',
    role: 'cut',
  }, {
    label: 'Copy',
    accelerator: 'CmdOrCtrl+C',
    role: 'copy',
  }, {
    label: 'Paste',
    accelerator: 'CmdOrCtrl+V',
    role: 'paste',
  }, {
    label: 'Select All',
    accelerator: 'CmdOrCtrl+A',
    role: 'selectall',
  }],
}, {
  label: 'View',
  submenu: [{
    label: 'Reload',
    accelerator: 'CmdOrCtrl+R',
    click: (item, focusedWindow) => {
      if (focusedWindow) {
        if (focusedWindow.id === 1) {
          BrowserWindow.getAllWindows().forEach((win) => {
            if (win.id > 1) {
              win.close();
            }
          });
        }
        focusedWindow.reload();
      }
    },
  }, {
    label: 'Toggle Full Screen',
    accelerator: (() => {
      if (process.platform === 'darwin') {
        return 'Ctrl+Command+F';
      } else {
        return 'F11';
      }
    })(),
    click: (item, focusedWindow) => {
      if (focusedWindow) {
        focusedWindow.setFullScreen(!focusedWindow.isFullScreen());
      }
    },
  }, {
    type: 'separator',
  }, {
    label: 'App Menu Demo',
    click: (item, focusedWindow) => {
      if (focusedWindow) {
        const options = {
          type: 'info',
          title: 'Application Menu Demo',
          buttons: ['Ok'],
          message: 'This demo is for the Menu section.',
        };
        dialog.showMessageBox(focusedWindow, options, () => undefined);
      }
    },
  }],
}, {
  label: 'Window',
  role: 'window',
  submenu: [{
    label: 'Minimize',
    accelerator: 'CmdOrCtrl+M',
    role: 'minimize',
  }, {
    label: 'Close',
    accelerator: 'CmdOrCtrl+W',
    role: 'close',
  }, {
    type: 'separator',
  }, {
    label: 'Reopen Window',
    accelerator: 'CmdOrCtrl+Shift+T',
    enabled: false,
    key: 'reopenMenuItem',
    click: () => {
      app.emit('activate');
    },
  }],
}, {
  label: 'Help',
  role: 'help',
  submenu: [{
    label: 'Learn More',
    click: () => {
      shell.openExternal('http://www.freechess.club');
    },
  }],
}];

function addUpdateMenuItems(items, position) {
  if (process.mas) {
      return;
  }

  const version = app.getVersion();
  const updateItems = [{
    label: `Version ${version}`,
    enabled: false,
  }, {
    label: 'Checking for Update',
    enabled: false,
    key: 'checkingForUpdate',
  }, {
    label: 'Check for Update',
    visible: false,
    key: 'checkForUpdate',
    click: () => {
      Electron.autoUpdater.checkForUpdates();
    },
  }, {
    label: 'Restart and Install Update',
    enabled: true,
    visible: false,
    key: 'restartToUpdate',
    click: () => {
      Electron.autoUpdater.quitAndInstall();
    },
  }];

  items.splice.apply(items, [position, 0].concat(updateItems));
}

if (process.platform === 'darwin') {
  const name = app.getName();
  template.unshift({
    label: name,
    submenu: [{
      label: `About ${name}`,
      role: 'about',
    }, {
      type: 'separator',
    }, {
      label: 'Services',
      role: 'services',
      submenu: [],
    }, {
      type: 'separator',
    }, {
      label: `Hide ${name}`,
      accelerator: 'Command+H',
      role: 'hide',
    }, {
      label: 'Hide Others',
      accelerator: 'Command+Alt+H',
      role: 'hideothers',
    }, {
      label: 'Show All',
      role: 'unhide',
    }, {
      type: 'separator',
    }, {
      label: 'Quit',
      accelerator: 'Command+Q',
      click: () => {
        app.quit();
      },
    }],
  } as any);

  // Window menu.
  (template[3].submenu as any).push({
    type: 'separator',
  }, {
    label: 'Bring All to Front',
    role: 'front',
  });

  addUpdateMenuItems(template[0].submenu, 1);
}

if (process.platform === 'win32') {
  const helpMenu = template[template.length - 1].submenu;
  addUpdateMenuItems(helpMenu, 0);
}

function findReopenMenuItem() {
  const menu = Menu.getApplicationMenu();
  if (!menu) {
    return;
  }

  let reopenMenuItem;
  menu.items.forEach((item: any) => {
    if (item.submenu) {
      item.submenu.items.forEach((subitem: any) => {
        if (subitem.key === 'reopenMenuItem') {
          reopenMenuItem = subitem;
        }
      });
    }
  });
  return reopenMenuItem;
}

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1200,
    height: 713,
    center: true,
    resizable: false,
    title: app.getName(),
    icon: path.join(__dirname, '../assets/img/tfcc-small.png'),
    webPreferences: {
      devTools: false,
    },
  });

  const ur = url.format({
    protocol: 'file',
    slashes: true,
    pathname: path.join(__dirname, '../app.html'),
  });

  mainWindow.loadURL(ur, {
    userAgent: 'The Free Chess Club',
  });

  mainWindow.on('closed', () => {
    mainWindow = null;
  });

  const menu = Menu.buildFromTemplate(template as any);
  Menu.setApplicationMenu(menu);
  mainWindow.show();
}

app.on('browser-window-created', () => {
  const reopenMenuItem = findReopenMenuItem();
  if (reopenMenuItem) {
    reopenMenuItem.enabled = false;
  }
});

app.setName('The Free Chess Club');

if (process.platform === 'darwin') {
  app.setAboutPanelOptions({
    applicationName: app.getName(),
    applicationVersion: app.getVersion(),
    copyright: 'Released under the MIT license',
    credits: 'The Free Chess Club Author(s)',
  });
}

app.on('ready', createWindow);

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit();
  }

  const reopenMenuItem = findReopenMenuItem();
  if (reopenMenuItem) {
    reopenMenuItem.enabled = true;
  }
});

app.on('activate', () => {
  if (mainWindow === null) {
    createWindow();
  }
});
