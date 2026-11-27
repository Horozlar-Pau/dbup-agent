// Electron main process: opens the management window, relays dbupd data to it,
// and drives the overlay windows (wait card, VIP alert, confetti) from what dbupd reports.
'use strict';

const { app, BrowserWindow, ipcMain } = require('electron');
const path = require('node:path');
const dbupd = require('./dbupd');
const { WaitCard } = require('./waitcard');
const { TrayMenu } = require('./tray');
const { Alerts } = require('./alerts');
const { cardStep } = require('./watch');
const { openCompose } = require('./compose');

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

// Opens (or brings back) the management window, optionally on a page.
function showMainWindow(page) {
  if (!mainWin) createWindow();
  else mainWin.show();
  if (page) mainWin.webContents.executeJavaScript(`location.hash = ${JSON.stringify(page)}`).catch(() => {});
  app.focus({ steal: true });
}

let tray = null;

// Like relay, with a longer timeout for calls that go on to the backend.
function relayWith(type, timeoutMs) {
  return async (_event, data) => {
    try {
      return { ok: true, data: await dbupd.send(type, data, { timeoutMs }) };
    } catch (err) {
      return { ok: false, error: err.message };
    }
  };
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

const cardState = { lastId: null, dismissed: false };

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

// ---------- VIP and moral alerts ----------

const alerts = new Alerts();
const markSeen = (id, reply) => dbupd.send('alert.ack', { id, reply }).catch((err) => console.error('ack failed:', err.message));
ipcMain.on('alert:ack', (_e, { reply }) => alerts.ackVip(reply));

// ---------- watch dbupd ----------

let watching = false;
async function watch() {
  if (watching) return;
  watching = true;
  try {
    const [events, inbox, pending] = await Promise.all([
      dbupd.send('events', { limit: 30 }),
      dbupd.send('nudges'),
      dbupd.send('alerts'),
    ]);
    alerts.handle(pending || [], markSeen);
    const queued = inbox?.queued || [];
    const step = cardStep(cardState, queued, card.open);
    if (step === 'show') card.show([...queued].reverse()); // oldest first on the card
    if (step === 'close') await closeCard();
    tray?.update({ events, pending: queued.length, connected: true });
  } catch {
    // dbupd not running: leave the overlays as they are
    tray?.update({ events: [], pending: 0, connected: false });
  } finally {
    watching = false;
  }
}

app.whenReady().then(() => {
  ipcMain.handle('dbupd:status', relay('status'));
  ipcMain.handle('dbupd:events', relay('events'));
  ipcMain.handle('dbupd:nudges', relay('nudges'));
  ipcMain.handle('compose:circle', relayWith('circle', 15000));
  ipcMain.handle('compose:send', relayWith('message.send', 15000));
  createWindow();
  tray = new TrayMenu({ onOpen: showMainWindow, onCompose: openCompose, onQuit: () => app.quit() });
  setInterval(watch, WATCH_MS);
  app.on('activate', () => showMainWindow());
});

app.on('window-all-closed', () => {
  // Keep running in the menu bar; quit from its menu or with Cmd+Q.
  if (process.platform !== 'darwin') app.quit();
});

module.exports = { showMainWindow, getTray: () => tray };
