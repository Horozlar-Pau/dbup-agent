// Client for dbupd's local Unix socket (see internal/ipc in the Go code).
// One JSON request line out, one JSON response line back, per connection.
'use strict';

const net = require('node:net');
const os = require('node:os');
const path = require('node:path');

const TIMEOUT_MS = 1000;

function socketPath(env = process.env) {
  const home = env.DBUP_HOME || path.join(os.homedir(), '.dbup');
  return path.join(home, 'dbupd.sock');
}

/**
 * Sends one request to dbupd.
 * Resolves with the response's data, rejects if dbupd is unreachable or answers with an error.
 */
function send(type, data, { sock = socketPath(), timeoutMs = TIMEOUT_MS } = {}) {
  return new Promise((resolve, reject) => {
    const conn = net.createConnection(sock);
    let buf = '';
    let done = false;
    const finish = (err, value) => {
      if (done) return;
      done = true;
      conn.destroy();
      err ? reject(err) : resolve(value);
    };

    conn.setTimeout(timeoutMs, () => finish(new Error('dbupd did not answer in time')));
    conn.on('error', (err) => finish(err));
    conn.on('connect', () => {
      const req = data === undefined ? { type } : { type, data };
      conn.write(JSON.stringify(req) + '\n');
    });
    conn.on('data', (chunk) => {
      buf += chunk;
      const nl = buf.indexOf('\n');
      if (nl === -1) return;
      let resp;
      try {
        resp = JSON.parse(buf.slice(0, nl));
      } catch (err) {
        return finish(new Error('bad response from dbupd'));
      }
      if (!resp.ok) return finish(new Error(resp.error || 'dbupd error'));
      finish(null, resp.data ?? null);
    });
    conn.on('end', () => finish(new Error('dbupd closed the connection')));
  });
}

module.exports = { send, socketPath };
