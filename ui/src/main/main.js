// Electron main process: opens the management window, relays dbupd data to it,
// and drives the overlay windows (wait card) from what dbupd reports.
'use strict';

const { app, BrowserWindow, ipcMain } = require('electron');
const path = require('node:path');
const dbupd = require('./dbupd');
const { WaitCard } = require('./waitcard');
const { cardStep } = require('./watch');

const WATCH_MS = 700;

let mainWin = null;

function createWindow() {
  mainWin = new BrowserWindow({
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
  mainWin.loadFile(path.join(__dirname, '..', 'renderer', 'index.html'));
  mainWin.on('closed', () => { mainWin = null; });
}

function showMainWindow() {
  if (!mainWin) createWindow();
  else mainWin.show();
  app.focus({ steal: true });
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

// ---------- wait card ----------

const cardState = { promptAt: null, dismissed: false };

const card = new WaitCard({});

async function closeCard() {
  // Nudges the developer saw but did not answer count as delivered.
  for (const n of card.close()) await dbupd.send('nudge.deliver', { id: n.id }).catch(() => {});
}

ipcMain.on('card:reply', async (_e, { id, reply }) => {
  try {
    await dbupd.send('nudge.reply', { id, reply });
    card.replied(id, reply);
  } catch (err) {
    console.error('reply failed:', err.message);
  }
});
ipcMain.on('card:dismiss', () => {
  cardState.dismissed = true;
  closeCard();
});
ipcMain.on('card:height', (_e, h) => card.fit(h));

// ---------- watch dbupd ----------

let watching = false;
async function watch() {
  if (watching) return;
  watching = true;
  try {
    const [events, inbox] = await Promise.all([dbupd.send('events', { limit: 30 }), dbupd.send('nudges')]);
    const queued = inbox?.queued || [];
    const step = cardStep(cardState, events || [], queued, card.open);
    if (step === 'show') card.show([...queued].reverse()); // oldest first on the card
    if (step === 'close') await closeCard();
  } catch {
    // dbupd not running: leave the overlays as they are
  } finally {
    watching = false;
  }
}

app.whenReady().then(() => {
  ipcMain.handle('dbupd:status', relay('status'));
  ipcMain.handle('dbupd:events', relay('events'));
  ipcMain.handle('dbupd:nudges', relay('nudges'));
  createWindow();
  setInterval(watch, WATCH_MS);
  app.on('activate', showMainWindow);
});

app.on('window-all-closed', () => {
  // Keep running for the overlays; quit from the menu bar (DBUP-38) or Cmd+Q.
  if (process.platform !== 'darwin') app.quit();
});

module.exports = { showMainWindow };
