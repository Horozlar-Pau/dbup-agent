'use strict';

const test = require('node:test');
const assert = require('node:assert');
const fs = require('node:fs');
const net = require('node:net');
const os = require('node:os');
const path = require('node:path');
const { send, socketPath } = require('./dbupd');

function fakeDbupd(handler) {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'dbupui'));
  const sock = path.join(dir, 's.sock');
  const server = net.createServer((c) => {
    let buf = '';
    c.on('data', (d) => {
      buf += d;
      if (!buf.includes('\n')) return;
      c.end(JSON.stringify(handler(JSON.parse(buf))) + '\n');
    });
  });
  return new Promise((resolve) => server.listen(sock, () => resolve({ sock, server, dir })));
}

test('socketPath honours DBUP_HOME', () => {
  assert.strictEqual(socketPath({ DBUP_HOME: '/x' }), '/x/dbupd.sock');
});

test('send returns data on success', async (t) => {
  const { sock, server } = await fakeDbupd((req) => ({ ok: true, data: { echo: req.type, limit: req.data.limit } }));
  t.after(() => server.close());
  const data = await send('events', { limit: 5 }, { sock });
  assert.deepStrictEqual(data, { echo: 'events', limit: 5 });
});

test('send rejects on error response', async (t) => {
  const { sock, server } = await fakeDbupd(() => ({ ok: false, error: 'nope' }));
  t.after(() => server.close());
  await assert.rejects(send('status', undefined, { sock }), /nope/);
});

test('send rejects when dbupd is not running', async () => {
  await assert.rejects(send('status', undefined, { sock: '/tmp/does-not-exist.sock' }));
});
