// VIP alert page: plays an alarm until one of the buttons is pressed.
'use strict';

let audio = null;

// Two-tone alarm with Web Audio, repeating every 1.2 s.
function startAlarm() {
  audio = new AudioContext();
  const beep = (t, f) => {
    const o = audio.createOscillator();
    const g = audio.createGain();
    o.frequency.value = f;
    o.type = 'sine';
    g.gain.setValueAtTime(0.0001, t);
    g.gain.exponentialRampToValueAtTime(0.35, t + 0.02);
    g.gain.exponentialRampToValueAtTime(0.0001, t + 0.28);
    o.connect(g).connect(audio.destination);
    o.start(t);
    o.stop(t + 0.3);
  };
  const loop = () => {
    if (!audio) return;
    const t = audio.currentTime;
    beep(t, 880);
    beep(t + 0.35, 660);
    setTimeout(loop, 1200);
  };
  loop();
}

function stopAlarm() {
  if (audio) audio.close();
  audio = null;
}

window.overlay.onData((a) => {
  document.getElementById('avatar').innerHTML = avatar(a.from, 84);
  document.getElementById('title').textContent = `${a.from} sana ulaşmaya çalışıyor`;
  document.getElementById('note').textContent = a.note ? `"${a.note}"` : '';
  document.getElementById('meta').textContent = `${hhmm(a.at)} · Bugünkü VIP hakkını kullandı`;
  document.getElementById('foot').textContent = `${a.from} "Gördüm" bilgisini görecek. Ses, sen basana kadar çalar.`;
  startAlarm();
});

document.addEventListener('click', (ev) => {
  const b = ev.target.closest('button[data-reply]');
  if (!b) return;
  stopAlarm();
  window.overlay.send('alert:ack', { reply: b.dataset.reply });
});
