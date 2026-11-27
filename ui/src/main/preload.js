// Exposes a small, read-only dbupd API to the page.
'use strict';

const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('dbup', {
  status: () => ipcRenderer.invoke('dbupd:status'),
  events: (limit = 50) => ipcRenderer.invoke('dbupd:events', { limit }),
});
