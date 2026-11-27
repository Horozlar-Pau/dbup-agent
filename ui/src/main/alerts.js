// VIP alert (full screen, sound until "Gördüm") and moral confetti (short, click-through).
'use strict';

const { BrowserWindow, screen } = require('electron');
const path = require('node:path');

const MORAL_MS = 4500;
const { newestFirstSplit } = require('./watch');

function overlayWindow(opts) {
  const { bounds } = screen.getPrimaryDisplay();
  const win = new BrowserWindow({
    ...bounds,
    frame: false,
    resizable: false,
    movable: false,
    minimizable: false,
    maximizable: false,
    fullscreenable: false,
    skipTaskbar: true,
    hasShadow: false,
    show: false,
    webPreferences: {
      preload: path.join(__dirname, 'overlay-preload.js'),
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: true,
      autoplayPolicy: 'no-user-gesture-required', // the VIP sound starts without a click
      backgroundThrottling: false, // keep animations smooth in a window without focus
    },
    ...opts,
  });
  win.setAlwaysOnTop(true, 'screen-saver');
  win.setVisibleOnAllWorkspaces(true, { visibleOnFullScreen: true });
  return win;
}

class Alerts {
  constructor() {
    this.vip = null; // { win, alert }
    this.moral = null;
    this.shown = new Set(); // alert ids already on screen or done
  }

  // Shows new alerts one at a time per kind. onSeen(id, reply) is called when one is done.
  handle(pending, onSeen) {
    this.onSeen = onSeen;
    // Only the newest few cheers are shown; older ones are marked seen and skipped.
    const { stale } = newestFirstSplit(pending.filter((a) => a.kind === 'moral' && !this.shown.has(a.id)));
    for (const a of stale) {
      this.shown.add(a.id);
      onSeen(a.id, '');
    }
    for (const a of pending) {
      if (this.shown.has(a.id)) continue;
      if (a.kind === 'vip' && !this.vip) this.showVip(a);
      else if (a.kind === 'moral' && !this.moral) this.showMoral(a);
    }
  }

  showVip(alert) {
    this.shown.add(alert.id);
    const win = overlayWindow({ backgroundColor: '#5C120D' });
    this.vip = { win, alert };
    win.loadFile(path.join(__dirname, '..', 'overlay', 'vip.html'));
    win.webContents.on('did-finish-load', () => {
      win.webContents.send('overlay:data', alert);
      win.show();
      win.focus();
    });
  }

  // Called when the developer presses one of the VIP buttons.
  ackVip(reply) {
    if (!this.vip) return;
    const { win, alert } = this.vip;
    this.vip = null;
    if (!win.isDestroyed()) win.destroy();
    this.onSeen?.(alert.id, reply);
  }

  showMoral(alert) {
    this.shown.add(alert.id);
    const win = overlayWindow({ transparent: true, focusable: false });
    win.setIgnoreMouseEvents(true); // never blocks clicks
    this.moral = { win, alert };
    win.loadFile(path.join(__dirname, '..', 'overlay', 'moral.html'));
    win.webContents.on('did-finish-load', () => {
      win.webContents.send('overlay:data', alert);
      win.showInactive();
    });
    setTimeout(() => this.doneMoral(), MORAL_MS);
  }

  doneMoral() {
    if (!this.moral) return;
    const { win, alert } = this.moral;
    this.moral = null;
    if (!win.isDestroyed()) win.destroy();
    this.onSeen?.(alert.id, '');
  }
}

module.exports = { Alerts };
