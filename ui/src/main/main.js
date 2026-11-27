// Electron main process: opens the management window and relays dbupd data to it.
// The UI decides nothing on its own; it only shows what dbupd reports.
'use strict';

const { app, BrowserWindow, ipcMain } = require('electron');
const path = require('node:path');
const dbupd = require('./dbupd');

function createWindow() {
  const win = new BrowserWindow({
    width: 1120,
    height: 740,
    minWidth: 900,
    minHeight: 600,
    title: 'dbup',
    titleBarStyle: 'hiddenInset',
    backgroundColor: '#f6f4f0',
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: true,
    },
  });
  win.loadFile(path.join(__dirname, '..', 'renderer', 'index.html'));
}

// Each handler answers { ok, data } or { ok: false, error } so the page can show
// "dbupd is not running" instead of throwing.
function relay(type) {
  return async (_event, data) => {
    try {
      return { ok: true, data: await dbupd.send(type, data) };
    } catch (err) {
      return { ok: false, error: err.message };
    }
  };
}

app.whenReady().then(() => {
  ipcMain.handle('dbupd:status', relay('status'));
  ipcMain.handle('dbupd:events', relay('events'));
  createWindow();
  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) createWindow();
  });
});

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') app.quit();
});
