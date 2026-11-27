// Moral message: a short confetti burst and the message in the middle of the screen.
'use strict';

const COLORS = ['#6E44C9', '#1E7F46', '#2F5BEA', '#C2410C', '#F4BF4F', '#EC6A5E'];

function confetti() {
  const cv = document.getElementById('confetti');
  const ctx = cv.getContext('2d');
  const dpr = window.devicePixelRatio || 1;
  cv.width = innerWidth * dpr;
  cv.height = innerHeight * dpr;
  ctx.scale(dpr, dpr);
  const parts = Array.from({ length: 180 }, () => ({
    x: innerWidth / 2 + (Math.random() - 0.5) * 200,
    y: innerHeight / 2,
    vx: (Math.random() - 0.5) * 18,
    vy: -Math.random() * 18 - 6,
    r: Math.random() * Math.PI,
    vr: (Math.random() - 0.5) * 0.3,
    w: 6 + Math.random() * 6,
    h: 10 + Math.random() * 8,
    c: COLORS[Math.floor(Math.random() * COLORS.length)],
  }));
  const start = performance.now();
  let prev = start;
  const frame = (now) => {
    const t = (now - start) / 1000;
    const k = Math.min(4, (now - prev) / 16.7); // steps of a 60 fps frame, so slow frames still move right
    prev = now;
    ctx.clearRect(0, 0, innerWidth, innerHeight);
    ctx.globalAlpha = Math.max(0, 1 - Math.max(0, t - 3) / 1.2);
    for (const p of parts) {
      p.vy += 0.45 * k;
      p.vx *= Math.pow(0.99, k);
      p.x += p.vx * k;
      p.y += p.vy * k;
      p.r += p.vr * k;
      ctx.save();
      ctx.translate(p.x, p.y);
      ctx.rotate(p.r);
      ctx.fillStyle = p.c;
      ctx.fillRect(-p.w / 2, -p.h / 2, p.w, p.h);
      ctx.restore();
    }
    if (t < 4.3) requestAnimationFrame(frame);
  };
  requestAnimationFrame(frame);
}

window.overlay.onData((a) => {
  document.getElementById('card').innerHTML = `
    ${avatar(a.from, 44)}
    <div>
      <div class="moral-from">${esc(a.from)} moral gönderdi</div>
      <div class="moral-note">${esc(a.note || 'Kolay gelsin!')}</div>
    </div>`;
  confetti();
});
