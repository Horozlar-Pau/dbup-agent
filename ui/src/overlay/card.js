// Wait card page: lists nudges with quick replies.
'use strict';

const REPLIES = ["10 dk'ya ararım", 'Akşam konuşalım', '❤️'];

function render({ nudges }) {
  document.getElementById('title').textContent = `${nudges.length} dürtme seni bekliyordu`;
  document.getElementById('list').innerHTML = nudges.map((n) => `
    <div class="wc-item">
      <div class="wc-row">
        ${avatar(n.from, 30)}
        <div class="wc-meta">
          <div class="wc-kind">${esc(n.kind)}</div>
          <div class="wc-from">${esc(n.from)} · ${hhmm(n.at)}</div>
        </div>
      </div>
      ${n.note ? `<div class="wc-note">"${esc(n.note)}"</div>` : ''}
      ${n.reply
        ? `<div class="wc-sent">Gönderildi: ${esc(n.reply)}</div>`
        : `<div class="wc-replies">${REPLIES.map((r) => `<button data-id="${n.id}" data-reply="${esc(r)}">${esc(r)}</button>`).join('')}</div>`}
    </div>`).join('');
  requestAnimationFrame(() => window.overlay.send('card:height', document.getElementById('card').getBoundingClientRect().height + 24));
}

document.addEventListener('click', (ev) => {
  const b = ev.target.closest('button[data-reply]');
  if (b) return window.overlay.send('card:reply', { id: Number(b.dataset.id), reply: b.dataset.reply });
  if (ev.target.closest('#close')) window.overlay.send('card:dismiss');
});

window.overlay.onData(render);
