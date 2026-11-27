// Shared helpers for overlay windows.
'use strict';

const esc = (s) => String(s ?? '').replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));

// Same colours as the design; unknown names get a stable colour from the palette.
const KNOWN = { elif: '#C2410C', can: '#2F5BEA', annem: '#1E7F46', anne: '#1E7F46', mert: '#6E44C9' };
const PALETTE = ['#C2410C', '#2F5BEA', '#1E7F46', '#6E44C9', '#B45309', '#0E7490'];
function colorFor(name) {
  const key = String(name).toLocaleLowerCase('tr-TR');
  if (KNOWN[key]) return KNOWN[key];
  let h = 0;
  for (const ch of key) h = (h * 31 + ch.codePointAt(0)) >>> 0;
  return PALETTE[h % PALETTE.length];
}
const avatar = (name, size) =>
  `<div class="avatar" style="width:${size}px;height:${size}px;background:${colorFor(name)}">${esc(String(name).charAt(0).toLocaleUpperCase('tr-TR'))}</div>`;
const hhmm = (iso) => new Date(iso).toLocaleTimeString('tr-TR', { hour: '2-digit', minute: '2-digit' });
