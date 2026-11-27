'use strict';

const $ = (id) => document.getElementById(id);
const to = $('to');
const text = $('text');
const send = $('send');
const status = $('status');

function setStatus(msg, kind) {
  status.textContent = msg;
  status.className = kind || '';
}

function refresh() {
  $('count').textContent = `${text.value.length} / 500`;
  send.disabled = !to.value || !text.value.trim();
}

async function loadCircle() {
  const res = await window.compose.circle();
  to.innerHTML = '';
  if (!res.ok) {
    to.append(new Option('Kişiler yüklenemedi', ''));
    setStatus(res.error, 'error');
    return;
  }
  const people = (res.data || []).filter((m) => !m.disabled);
  if (people.length === 0) {
    to.append(new Option('Yakın çevrende kimse yok', ''));
    return;
  }
  to.append(new Option('Birini seç', ''));
  for (const m of people) to.append(new Option(`${m.displayName} (@${m.username})`, m.id));
  if (people.length === 1) to.value = people[0].id;
  refresh();
}

$('form').addEventListener('submit', async (e) => {
  e.preventDefault();
  if (send.disabled) return;
  send.disabled = true;
  setStatus('Gönderiliyor…');
  const res = await window.compose.send(to.value, text.value.trim());
  if (res.ok) {
    setStatus(`Gönderildi: ${to.selectedOptions[0].textContent.split(' (')[0]} sayfasında görecek.`, 'ok');
    text.value = '';
  } else {
    setStatus(res.error, 'error');
  }
  refresh();
});
to.addEventListener('change', refresh);
text.addEventListener('input', refresh);
loadCircle();
