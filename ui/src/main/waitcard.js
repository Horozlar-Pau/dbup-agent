// Wait card: a small corner card that shows queued nudges while Claude is working.
// It never takes keyboard focus and makes no sound.
'use strict';

const { BrowserWindow, screen } = require('electron');
const path = require('node:path');

const WIDTH = 380;
const MARGIN = 16;

class WaitCard {
  constructor({ onReply, onClose }) {
    this.win = null;
    this.items = new Map(); // id -> nudge shown on this card (reply added when answered)
    this.onReply = onReply;
    this.onClose = onClose;
  }

  get open() {
    return !!this.win && !this.win.isDestroyed();
  }

  // Adds nudges to the card and shows it without stealing focus.
  show(nudges) {
    for (const n of nudges) if (!this.items.has(n.id)) this.items.set(n.id, { ...n });
    if (!this.open) this.create();
    else this.push();
  }

  // Marks a nudge as answered on the card.
  replied(id, reply) {
    const n = this.items.get(id);
    if (!n) return;
    n.reply = reply;
    this.push();
  }

  // Closes the card; returns the nudges that got no reply.
  close() {
    const unanswered = [...this.items.values()].filter((n) => !n.reply);
    this.items.clear();
    if (this.open) this.win.destroy();
    this.win = null;
    return unanswered;
  }

  create() {
    const area = screen.getPrimaryDisplay().workArea;
    this.win = new BrowserWindow({
      width: WIDTH,
      height: 200,
      x: area.x + area.width - WIDTH - MARGIN,
      y: area.y + MARGIN,
      frame: false,
      transparent: true,
      resizable: false,
      movable: true,
      minimizable: false,
      maximizable: false,
      fullscreenable: false,
      skipTaskbar: true,
      focusable: false, // never take keyboard focus
      alwaysOnTop: true,
      hasShadow: false,
      show: false,
      webPreferences: {
        preload: path.join(__dirname, 'overlay-preload.js'),
        contextIsolation: true,
        nodeIntegration: false,
        sandbox: true,
      },
    });
    this.win.setAlwaysOnTop(true, 'floating');
    this.win.setVisibleOnAllWorkspaces(true, { visibleOnFullScreen: true });
    this.win.loadFile(path.join(__dirname, '..', 'overlay', 'card.html'));
    this.win.webContents.on('did-finish-load', () => {
      this.push();
      this.win.showInactive();
    });
  }

  push() {
    if (!this.open) return;
    this.win.webContents.send('overlay:data', { nudges: [...this.items.values()] });
  }

  // Fits the window to the card's content height.
  fit(height) {
    if (!this.open) return;
    const [x, y] = this.win.getPosition();
    this.win.setBounds({ x, y, width: WIDTH, height: Math.ceil(height) });
  }
}

module.exports = { WaitCard };
