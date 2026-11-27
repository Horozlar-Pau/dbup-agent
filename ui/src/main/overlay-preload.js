// Preload for the small always-on-top windows (wait card, VIP alert, confetti).
// Only the listed channels can be used.
'use strict';

const { contextBridge, ipcRenderer } = require('electron');

const SEND = new Set(['card:reply', 'card:dismiss', 'card:height', 'alert:ack', 'moral:done']);

contextBridge.exposeInMainWorld('overlay', {
  onData: (cb) => ipcRenderer.on('overlay:data', (_e, data) => cb(data)),
  send: (channel, payload) => {
    if (SEND.has(channel)) ipcRenderer.send(channel, payload);
  },
});
