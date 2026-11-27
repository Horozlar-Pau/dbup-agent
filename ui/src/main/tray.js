// Menu bar (tray) menu: status, pending nudges, Hidden mode, open window, quit.
// Manual status and Hidden mode only live here for now; the backend will own them later.
'use strict';

const { Tray, Menu, nativeImage } = require('electron');

const STATUS = {
  ai_working: { label: 'AI ile çalışıyor', color: [0x6e, 0x44, 0xc9] },
  focus: { label: 'Derin odak', color: [0x2f, 0x5b, 0xea] },
  available: { label: 'Müsait', color: [0x1e, 0x7f, 0x46] },
  busy: { label: 'Rahatsız etme', color: [0xb3, 0x26, 0x1e] },
  hidden: { label: 'Gizli mod', color: [0x5f, 0x5d, 0x57] },
};

// Local guess from the latest event, same rule as the management window.
const FROM_EVENT = {
  'ai.prompt_started': 'ai_working',
  'ai.prompt_finished': 'focus',
  'ai.session_ended': 'available',
  'git.commit': 'available',
};

// A filled circle as the tray icon, coloured by status (drawn as a BGRA bitmap).
function dot([r, g, b]) {
  const size = 32; // @2x for a 16pt icon
  const buf = Buffer.alloc(size * size * 4);
  const c = (size - 1) / 2;
  for (let y = 0; y < size; y++) {
    for (let x = 0; x < size; x++) {
      const d = Math.hypot(x - c, y - c);
      const a = Math.max(0, Math.min(1, 11 - d)); // radius 11 with a 1px soft edge
      const i = (y * size + x) * 4;
      // premultiplied alpha, as Electron expects
      buf[i] = Math.round(b * a); buf[i + 1] = Math.round(g * a); buf[i + 2] = Math.round(r * a); buf[i + 3] = Math.round(a * 255);
    }
  }
  return nativeImage.createFromBitmap(buf, { width: size, height: size, scaleFactor: 2 });
}

function since(ms) {
  const m = Math.max(1, Math.round(ms / 60000));
  return m < 60 ? `${m} dk'dır` : `${Math.floor(m / 60)} sa ${m % 60} dk'dır`;
}

class TrayMenu {
  constructor({ onOpen, onCompose, onQuit }) {
    this.onOpen = onOpen;
    this.onCompose = onCompose;
    this.onQuit = onQuit;
    this.mode = 'auto'; // auto | manual
    this.manual = 'available';
    this.hidden = false;
    this.auto = 'available';
    this.autoSince = Date.now();
    this.lastEventAt = null;
    this.pending = 0;
    this.connected = false;
    this.key = '';
    this.tray = new Tray(dot(STATUS.available.color));
    this.tray.setToolTip('dbup');
    this.render();
    setInterval(() => this.render(), 30000); // keep "x dk'dır" fresh
  }

  // Called by the dbupd watcher with the newest events and queued nudge count.
  update({ events, pending, connected }) {
    this.connected = connected;
    this.pending = pending;
    const last = [...(events || [])].reverse().find((e) => FROM_EVENT[e.type]);
    if (last && last.at !== this.lastEventAt) {
      const next = FROM_EVENT[last.type];
      if (next !== this.auto || !this.lastEventAt) this.autoSince = new Date(last.at).getTime();
      this.auto = next;
      this.lastEventAt = last.at;
    }
    this.render();
  }

  current() {
    if (this.hidden) return 'hidden';
    return this.mode === 'manual' ? this.manual : this.auto;
  }

  render() {
    const cur = this.current();
    const s = STATUS[cur];
    const how = this.hidden ? 'Takip durdu' : this.mode === 'manual' ? 'Elle · 1 saat' : 'Otomatik';
    const sub = this.connected ? `${since(Date.now() - this.autoSince)} · ${how}` : 'dbupd kapalı';
    const key = JSON.stringify([cur, sub, this.pending, this.mode, this.manual, this.hidden, this.connected]);
    if (key === this.key) return; // avoid rebuilding an open menu for nothing
    this.key = key;

    this.tray.setImage(dot(s.color));
    this.tray.setTitle(this.pending ? ` ${this.pending}` : '');
    const pick = (id) => ({
      label: id === 'auto' ? 'Otomatik' : `${STATUS[id].label}    1 saat`,
      type: 'radio',
      checked: id === 'auto' ? this.mode === 'auto' : this.mode === 'manual' && this.manual === id,
      click: () => {
        if (id === 'auto') this.mode = 'auto';
        else { this.mode = 'manual'; this.manual = id; }
        this.hidden = false;
        this.render();
      },
    });
    // _lastMenu is kept for checks that print the menu.
    this.tray.setContextMenu(this._lastMenu = Menu.buildFromTemplate([
      { label: s.label, enabled: false },
      { label: sub, enabled: false },
      { type: 'separator' },
      { label: `Bekleyen dürtme    ${this.pending}`, click: () => this.onOpen('durtmeler') },
      { type: 'separator' },
      { label: 'Durum', enabled: false },
      pick('auto'), pick('available'), pick('focus'), pick('busy'),
      { type: 'separator' },
      { label: 'Gizli mod', type: 'checkbox', checked: this.hidden, click: () => { this.hidden = !this.hidden; this.render(); } },
      { label: 'Bu repoyu takip etme', enabled: false, toolTip: 'Ayarlar > Takip dışı repolar' },
      { type: 'separator' },
      { label: 'Mesaj gönder…', click: () => this.onCompose() },
      { label: "dbup'ı aç…", click: () => this.onOpen() },
      { label: 'Çık', click: () => this.onQuit() },
    ]));
  }
}

module.exports = { TrayMenu };
