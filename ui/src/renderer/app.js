// Management window: sidebar navigation and five pages.
// Pages use fake data (data.js) except Teşhis and the status card, which read dbupd.
'use strict';

const D = window.FAKE;
const POLL_MS = 1000;

// ---------- helpers ----------

const esc = (s) => String(s ?? '').replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
const person = (id) => D.circle.find((p) => p.id === id) || { name: id, color: '#999' };
const avatar = (p) => `<div class="avatar" style="background:${esc(p.color)}">${esc(p.name[0])}</div>`;
const inviteURL = (p) => `https://${D.host}/u/${D.username}?davet=${p.code}`;
const randomCode = () => Array.from(crypto.getRandomValues(new Uint8Array(6)), (b) => 'abcdefghjkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789'[b % 56]).join('');
const clock = (iso) => new Date(iso).toLocaleTimeString('tr-TR', { hour: '2-digit', minute: '2-digit', second: '2-digit' });

const STATUSES = {
  ai_working: { label: 'AI ile çalışıyor', cls: 'ai', detail: 'Claude yanıt üretiyor' },
  focus: { label: 'Derin odak', cls: 'focus', detail: 'Dürtmeler bekleme penceresine kadar bekler' },
  available: { label: 'Müsait', cls: 'ok', detail: 'Dürtmeler hemen iletilir' },
  busy: { label: 'Rahatsız etme', cls: 'busy', detail: 'Süre bitene kadar her şey bekler. VIP yine gelir' },
  hidden: { label: 'Gizli', cls: 'hidden', detail: 'Takip duraklatıldı' },
};

// Local guess from the latest dbupd event; the real status will come from dbup-backend.
const STATUS_FROM_EVENT = {
  'ai.prompt_started': 'ai_working',
  'ai.prompt_finished': 'focus',
  'ai.session_ended': 'available',
  'git.commit': 'available',
};

const EVENT_TEXT = {
  'ai.prompt_started': 'Claude çalışmaya başladı',
  'ai.prompt_finished': 'Claude yanıtı bitti',
  'ai.session_ended': 'Claude oturumu kapandı',
  'git.commit': 'commit atıldı',
};

// ---------- state ----------

const state = {
  page: 'durum',
  mode: 'auto', // auto | manual
  manualStatus: 'focus',
  manualFor: '1 saat',
  hidden: false,
  autoStatus: 'ai_working',
  autoSince: Date.now() - 42 * 60 * 1000, // fake until the first dbupd event
  lastEventAt: null,
  nudgeFilter: 'all',
  daemon: null, // dbupd status, null when unreachable
  daemonError: '',
  events: [],
  newInvite: null,
};

function currentStatus() {
  if (state.hidden) return 'hidden';
  return state.mode === 'manual' ? state.manualStatus : state.autoStatus;
}

// ---------- pages ----------

const pages = {
  durum() {
    const key = currentStatus();
    const s = STATUSES[key];
    const mins = Math.max(1, Math.round((Date.now() - state.autoSince) / 60000));
    const dur = mins < 60 ? `${mins} dk'dır` : `${Math.floor(mins / 60)} sa ${mins % 60} dk'dır`;
    const how = state.hidden ? 'Gizli mod açık' : state.mode === 'manual' ? `Elle seçildi · ${state.manualFor}` : 'Otomatik: ajan belirliyor';
    const pick = (k) => `
      <button class="card pick ${state.mode === 'manual' && state.manualStatus === k ? 'on' : ''}" data-act="manual" data-status="${k}">
        <b><span class="dot ${STATUSES[k].cls}"></span>${STATUSES[k].label}</b>
        <div class="muted small">${STATUSES[k].detail}.</div>
      </button>`;
    const chip = (v) => `<button class="chip ${state.manualFor === v ? 'on' : ''}" data-act="for" data-value="${v}">${v}</button>`;
    const queued = D.nudges.queued;
    return `
      <div class="page-head">
        <h1>Durum</h1>
        <div class="seg">
          <button class="${state.mode === 'auto' ? 'on' : ''}" data-act="mode" data-value="auto">Otomatik</button>
          <button class="${state.mode === 'manual' ? 'on' : ''}" data-act="mode" data-value="manual">Elle</button>
        </div>
      </div>
      <div class="card status-hero">
        <div class="status-orb ${s.cls}"></div>
        <div>
          <div class="muted small">Yakın çevre şu an bunu görüyor</div>
          <div class="status-title">${s.label}</div>
          <div class="muted small">${s.detail} · ${how}</div>
        </div>
        <div class="status-dur"><div class="muted small">Süre</div><b>${dur}</b></div>
      </div>

      <div class="section-title"><h3>Elle durum seç</h3><span class="muted small">Seçince Elle moda geçilir</span></div>
      <div class="grid g3">${pick('available')}${pick('focus')}${pick('busy')}</div>
      <div class="chips" style="margin-top:10px"><span class="muted small">Ne kadar süre?</span>${['30 dk', '1 saat', 'Akşama kadar', 'Ben kapatana kadar'].map(chip).join('')}</div>

      <div class="grid g3" style="margin-top:16px">
        <div class="card">
          <div class="setting" style="padding-top:0;border:none"><h3 style="margin:0">Gizli mod</h3><button class="toggle ${state.hidden ? 'on' : ''}" data-act="hidden" aria-label="Gizli mod"></button></div>
          <div class="muted small">Takip tamamen durur. Yakın çevre "Gizli" yerine son durumu ve saatini görür.</div>
        </div>
        <div class="card">
          <h3>Bekleyen dürtmeler <span class="badge queued">${queued.length} sırada</span></h3>
          <div class="muted small">${queued.map((n) => `${esc(person(n.from).name)}: "${esc(n.kind)}"`).join(' · ')}. Claude yanıt üretmeye başlayınca gösterilecek.</div>
          <button class="link" data-go="durtmeler" style="margin-top:8px">Dürtmeleri gör</button>
        </div>
        <div class="card">
          <h3>Bugün</h3>
          <div class="stat-row"><span>AI ile çalışma</span><b>${D.today.aiMinutes}</b></div>
          <div class="stat-row"><span>Mola</span><b>${D.today.breakMinutes}</b></div>
          <div class="stat-row"><span>Cevaplanan dürtme</span><b>${D.today.replies}</b></div>
        </div>
      </div>`;
  },

  'yakin-cevre'() {
    const rows = D.circle.map((p) => `
      <tr>
        <td><div class="person">${avatar(p)}<div><b>${esc(p.name)}</b><div class="muted small">@${esc(p.id)}</div></div></div></td>
        <td class="muted small">${esc(p.lastSeen)}</td>
        <td><span class="badge ${p.vip ? 'on' : 'off'}">${p.vip ? 'Açık' : 'Kapalı'}</span></td>
        <td><span class="badge ${p.invite === 'active' ? 'on' : 'revoked'}">${p.invite === 'active' ? 'Etkin' : 'İptal edildi'}</span></td>
        <td><div class="actions">${p.invite === 'active'
          ? `<button data-act="copy-invite" data-id="${p.id}">Linki kopyala</button><button data-act="revoke" data-id="${p.id}">İptal et</button>`
          : `<button data-act="renew" data-id="${p.id}">Yeni link</button>`}</div></td>
      </tr>`).join('');
    const inv = state.newInvite;
    return `
      <div class="page-head"><h1>Yakın çevre</h1></div>
      <p class="sub">Bu kişiler sana özel davet linkiyle durum sayfanı açar. Hesap veya şifre gerekmez.</p>
      <div class="two-col">
        <div class="card" style="padding:6px 8px">
          <table class="table">
            <thead><tr><th>Kişi</th><th>Son giriş</th><th>VIP</th><th>Davet linki</th><th style="text-align:right">İşlemler</th></tr></thead>
            <tbody>${rows}</tbody>
          </table>
          <div class="muted small" style="padding:8px">Link iptal edilince kişinin açık web oturumları kapanır. Yeni link verene kadar sayfanı göremez.</div>
        </div>
        <div class="card">
          <h3>Kişi ekle</h3>
          <label class="field" for="new-name">Görünen ad</label>
          <input type="text" id="new-name" placeholder="Zeynep" value="${esc(inv?.name || '')}">
          <label class="check"><input type="checkbox" id="new-vip" checked> VIP uyarısı gönderebilsin</label>
          <button class="primary block" data-act="create-invite">Davet linki oluştur</button>
          ${inv ? `
            <div class="invite"><input type="text" readonly value="${esc(inv.url)}"><button data-act="copy-new">Kopyala</button></div>
            <div class="warn">Bu linki yalnızca ${esc(inv.name)} ile paylaş. Linki bilen herkes durum sayfanı açabilir; istediğin an iptal edebilirsin.</div>` : ''}
        </div>
      </div>`;
  },

  durtmeler() {
    const f = state.nudgeFilter;
    const match = (n) => f === 'all' || n.from === f;
    const chip = (id, label) => `<button class="chip ${f === id ? 'on' : ''}" data-act="filter" data-value="${id}">${label}</button>`;
    const row = (n, badge) => {
      const p = person(n.from);
      const second = n.note ? `"${esc(n.note)}"` : n.reply ? `Cevabın: ${esc(n.reply)}` : '';
      return `<div class="row">${avatar(p)}<div class="grow"><b>${esc(n.kind)}</b> <span class="muted">· ${esc(p.name)}</span>${second ? `<div class="muted small">${second}</div>` : ''}</div><span class="muted small">${esc(n.at)}</span>${badge}</div>`;
    };
    const queued = D.nudges.queued.filter(match);
    const history = D.nudges.history.filter(match);
    const BADGES = { replied: '<span class="badge replied">Cevaplandı</span>', delivered: '<span class="badge delivered">İletildi</span>' };
    return `
      <div class="page-head">
        <h1>Dürtmeler</h1>
        <div class="chips">${chip('all', 'Hepsi')}${D.circle.map((p) => chip(p.id, esc(p.name))).join('')}</div>
      </div>
      <div class="label-caps">Sırada</div>
      <div class="card list" style="padding:0">${queued.map((n) => row(n, '<span class="badge queued">Sırada</span>')).join('') || '<div class="row muted">Sırada dürtme yok</div>'}</div>
      <div class="muted small" style="margin-top:6px">Claude yanıt üretmeye başlayınca, commit atınca veya Müsait olunca bekleme kartında gösterilir.</div>
      <div class="label-caps">Geçmiş</div>
      <div class="card list" style="padding:0">${history.map((n) => row(n, BADGES[n.state])).join('') || '<div class="row muted">Geçmiş yok</div>'}</div>`;
  },

  ayarlar() {
    const d = state.daemon;
    const setting = (s) => `
      <div class="setting"><div><b>${esc(s.title)}</b><div class="muted small">${esc(s.desc)}</div></div>
      <button class="toggle ${s.on ? 'on' : ''}" data-act="privacy" data-id="${s.id}" aria-label="${esc(s.title)}"></button></div>`;
    return `
      <div class="page-head"><h1>Ayarlar</h1></div>
      <div class="two-col even">
        <div>
          <div class="card"><h3>Gizlilik</h3>${D.privacy.map(setting).join('')}
            <div class="muted small" style="margin-top:8px">Kod, prompt metni, dosya ve klasör adları hiçbir zaman bilgisayardan çıkmaz.</div>
          </div>
          <div class="card" style="margin-top:12px">
            <div class="setting" style="padding-top:0;border:none"><h3 style="margin:0">Takip dışı repolar</h3><button class="link" disabled>+ Klasör ekle</button></div>
            ${D.ignoredRepos.map((r) => `<div class="repo">${esc(r)}<span class="muted">×</span></div>`).join('')}
          </div>
        </div>
        <div>
          <div class="card"><h3>Sunucu ve cihaz</h3>
            <div class="kv">
              <span>Sunucu</span><span>${esc(d?.backend_url || '—')}</span>
              <span>Hesap</span><span>${esc(D.username)}</span>
              <span>Token</span><span>${d ? (d.token_set ? 'Girildi' : 'Girilmedi') : '—'}</span>
            </div>
            <div class="chips"><button disabled>Sunucuyu değiştir</button><button disabled>Çıkış yap</button></div>
          </div>
          <div class="card" style="margin-top:12px"><h3>Hook'lar</h3>
            <div class="setting"><span>Claude Code</span><span class="badge on">Kurulu</span></div>
            <div class="setting"><span>Git post-commit</span><span class="badge on">Kurulu</span></div>
            <div class="chips" style="margin-top:8px"><button disabled>Yeniden kur</button><button disabled>Kaldır</button></div>
          </div>
          <div class="card" style="margin-top:12px"><h3>dbup'ı kaldır</h3>
            <div class="muted small">Hook'ları temizler, token'ı siler ve servisi durdurur.</div>
            <button class="danger" style="margin-top:10px" disabled>Kaldır</button>
          </div>
        </div>
      </div>`;
  },

  teshis() {
    const d = state.daemon;
    const up = d ? `${Math.floor(d.uptime_seconds / 60)} dk açık` : '';
    const metric = (title, dotCls, value, sub) => `<div class="card metric"><div class="muted small">${title}</div><b><span class="dot ${dotCls}"></span>${value}</b><div class="muted small">${sub}</div></div>`;
    const rows = [...state.events].reverse().map((e) => `
      <div class="log-row"><span class="t">${clock(e.at)}</span><span class="tag-olay">olay</span><span>${esc(e.type)} alındı · ${esc(EVENT_TEXT[e.type] || '')}</span></div>`).join('');
    const empty = d
      ? '<div class="log-empty">Henüz olay yok. Claude\'a bir prompt gönderin veya commit atın.</div>'
      : `<div class="log-empty">dbupd'ye ulaşılamıyor. Terminalde <b>dbupd</b> komutuyla başlatın.<br>${esc(state.daemonError)}</div>`;
    return `
      <div class="page-head">
        <h1>Teşhis</h1>
        <div class="chips"><button data-act="refresh">Yeniden test et</button><button class="primary" data-act="export">Log'u dışa aktar</button></div>
      </div>
      <div class="grid g4">
        ${metric('Ajan servisi', d ? 'ok' : 'busy', d ? 'Çalışıyor' : 'Kapalı', d ? `dbupd ${esc(d.version)} · ${up}` : 'dbupd başlatılmadı')}
        ${metric('Sunucu bağlantısı', '', 'Henüz yok', 'Backend bağlantısı sonra eklenecek')}
        ${metric('Alınan olay', d ? 'ai' : '', d ? d.event_count : '—', 'Son 200 olay tutulur')}
        ${metric('macOS izinleri', 'ok', 'Gerekmiyor', 'Odak takibi henüz yok')}
      </div>
      <div class="log">
        <div class="log-head"><b>Son olaylar</b><span>Kod ve prompt metni log'a yazılmaz</span></div>
        ${rows || empty}
      </div>`;
  },
};

// ---------- rendering ----------

function render() {
  document.getElementById('page').innerHTML = pages[state.page]();
  document.querySelectorAll('#nav a').forEach((a) => a.classList.toggle('active', a.dataset.page === state.page));
  renderSidebar();
}

function renderSidebar() {
  const s = STATUSES[currentStatus()];
  document.getElementById('me-status').textContent = s.label;
  document.getElementById('me-dot').className = `dot ${s.cls}`;
  document.getElementById('count-circle').textContent = D.circle.filter((p) => p.invite === 'active').length;
  document.getElementById('count-nudges').textContent = D.nudges.queued.length;
  const d = state.daemon;
  document.getElementById('conn-dot').className = `dot ${d ? 'ok' : 'busy'}`;
  document.getElementById('conn-label').textContent = d ? 'dbupd çalışıyor' : 'dbupd kapalı';
  document.getElementById('conn-host').textContent = d ? d.backend_url.replace(/^https?:\/\//, '') : 'Terminalde dbupd ile başlatın';
}

function go(page) {
  if (!pages[page]) page = 'durum';
  state.page = page;
  if (location.hash !== `#${page}`) history.replaceState(null, '', `#${page}`);
  render();
}

// ---------- actions ----------

async function copy(text) {
  try { await navigator.clipboard.writeText(text); } catch { /* clipboard may be unavailable */ }
}

const actions = {
  mode: (el) => { state.mode = el.dataset.value; },
  manual: (el) => { state.mode = 'manual'; state.manualStatus = el.dataset.status; },
  for: (el) => { state.manualFor = el.dataset.value; state.mode = 'manual'; },
  hidden: () => { state.hidden = !state.hidden; },
  filter: (el) => { state.nudgeFilter = el.dataset.value; },
  privacy: (el) => { const s = D.privacy.find((x) => x.id === el.dataset.id); s.on = !s.on; },
  'copy-invite': (el) => copy(inviteURL(person(el.dataset.id))),
  revoke: (el) => { person(el.dataset.id).invite = 'revoked'; },
  renew: (el) => { const p = person(el.dataset.id); p.invite = 'active'; p.code = randomCode(); },
  'create-invite': () => {
    const name = document.getElementById('new-name').value.trim();
    if (!name) return document.getElementById('new-name').focus();
    const id = name.toLowerCase().normalize('NFD').replace(/[^a-z0-9]/g, '') || randomCode().toLowerCase();
    const p = { id, name, color: '#8a6d3b', lastSeen: 'Hiç girmedi', vip: document.getElementById('new-vip').checked, invite: 'active', code: randomCode() };
    D.circle.push(p);
    state.newInvite = { name, url: inviteURL(p) };
  },
  'copy-new': () => copy(state.newInvite.url),
  refresh: () => poll(),
  export: () => {
    const lines = state.events.map((e) => `${e.at}\t${e.source}\t${e.type}`).join('\n');
    const a = document.createElement('a');
    a.href = URL.createObjectURL(new Blob([lines + '\n'], { type: 'text/plain' }));
    a.download = 'dbup-log.txt';
    a.click();
    URL.revokeObjectURL(a.href);
  },
};

document.addEventListener('click', (ev) => {
  const goEl = ev.target.closest('[data-go]');
  if (goEl) return go(goEl.dataset.go);
  const el = ev.target.closest('[data-act]');
  if (!el || el.disabled || !actions[el.dataset.act]) return;
  const result = actions[el.dataset.act](el);
  if (result instanceof Promise) return;
  render();
});

window.addEventListener('hashchange', () => go(location.hash.slice(1)));

// ---------- dbupd polling ----------

async function poll() {
  if (!window.dbup) return; // opened outside Electron (plain browser preview)
  const [st, ev] = await Promise.all([window.dbup.status(), window.dbup.events(50)]);
  const before = JSON.stringify([state.daemon?.event_count, !!state.daemon]);
  state.daemon = st.ok ? st.data : null;
  state.daemonError = st.ok ? '' : st.error;
  state.events = ev.ok ? ev.data || [] : [];

  const last = state.events[state.events.length - 1];
  const next = last && STATUS_FROM_EVENT[last.type];
  // A new status starts its timer at the event that caused it; repeats keep the timer.
  if (next && last.at !== state.lastEventAt) {
    if (next !== state.autoStatus || !state.lastEventAt) state.autoSince = new Date(last.at).getTime();
    state.autoStatus = next;
    state.lastEventAt = last.at;
  }

  const after = JSON.stringify([state.daemon?.event_count, !!state.daemon]);
  // Re-render only when something changed and the user is not typing.
  if (before !== after && document.activeElement?.tagName !== 'INPUT') render();
  else renderSidebar();
}

go(location.hash.slice(1) || 'durum');
poll();
setInterval(poll, POLL_MS);
