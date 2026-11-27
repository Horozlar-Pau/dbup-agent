// Preload for the "Mesaj gönder" window. Only these two calls are exposed.
'use strict';

const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('compose', {
  circle: () => ipcRenderer.invoke('compose:circle'),
  send: (memberId, text) => ipcRenderer.invoke('compose:send', { memberId, text }),
});
