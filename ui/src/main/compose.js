// "Mesaj gönder" window: the developer picks someone from the close circle and
// writes a message. dbupd sends it to the backend; it shows on their page.
'use strict';

const { BrowserWindow } = require('electron');
const path = require('node:path');

let win = null;

function openCompose() {
  if (win && !win.isDestroyed()) {
    win.show();
    win.focus();
    return;
  }
  win = new BrowserWindow({
    width: 420,
    height: 380,
    resizable: false,
    minimizable: false,
    maximizable: false,
    title: 'Mesaj gönder',
    backgroundColor: '#f6f4f0',
    webPreferences: {
      preload: path.join(__dirname, 'compose-preload.js'),
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: true,
    },
  });
  win.loadFile(path.join(__dirname, '..', 'overlay', 'compose.html'));
  win.on('closed', () => { win = null; });
}

module.exports = { openCompose };
