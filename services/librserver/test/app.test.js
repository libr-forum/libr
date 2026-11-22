const { test } = require('node:test');
const assert = require('node:assert');
const crypto = require('crypto');
const { createApp, registerMessage } = require('../src/app');
const { peerIdFromEd25519 } = require('../src/peerid');

// Start an app on a random port with a controllable clock.
async function serve(opts = {}) {
  const clock = { t: 1_000_000 };
  const server = createApp({ log: false, now: () => clock.t, ...opts }).listen(0);
  await new Promise((r) => server.once('listening', r));
  const base = `http://127.0.0.1:${server.address().port}`;
  return { base, clock, close: () => server.close() };
}

function keypair() {
  const { publicKey, privateKey } = crypto.generateKeyPairSync('ed25519');
  const raw = publicKey.export({ format: 'der', type: 'spki' }).subarray(12);
  return { raw, b64: raw.toString('base64'), peerId: peerIdFromEd25519(raw), privateKey };
}

async function register(base, k, { peerId = k.peerId, addrs = ['/ip4/1.2.3.4/tcp/4001'], tamper } = {}) {
  const { nonce } = await (await fetch(`${base}/auth/challenge?publicKey=${encodeURIComponent(k.b64)}`)).json();
  let signature = crypto.sign(null, Buffer.from(registerMessage(nonce, peerId, addrs)), k.privateKey).toString('base64');
  if (tamper) signature = tamper(signature);
  return fetch(`${base}/nodes/register`, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ publicKey: k.b64, peerId, addrs, nonce, signature }),
  });
}

test('peer ID matches go-libp2p (all-zero seed vector from librctl keygen)', () => {
  const raw = Buffer.from('O2onvM62pC1io6jQKm8Nc2UyFXcd4kOmOsBIoYtZ2ik=', 'base64');
  assert.strictEqual(peerIdFromEd25519(raw), '12D3KooWDpJ7As7BWAwRMfu1VU2WCqNjvq387JEYKDBj4kx6nXTN');
});

test('GET /health', async () => {
  const s = await serve();
  try {
    assert.deepStrictEqual(await (await fetch(`${s.base}/health`)).json(), { ok: true });
  } finally { s.close(); }
});

test('register then list; entries expire', async () => {
  const s = await serve({ ttlMs: 60_000 });
  try {
    const k = keypair();
    const res = await register(s.base, k);
    assert.strictEqual(res.status, 200, await res.text());
    let peers = await (await fetch(`${s.base}/peers`)).json();
    assert.deepStrictEqual(peers, [{ peerId: k.peerId, addrs: ['/ip4/1.2.3.4/tcp/4001'] }]);
    s.clock.t += 61_000;
    peers = await (await fetch(`${s.base}/peers`)).json();
    assert.deepStrictEqual(peers, []);
  } finally { s.close(); }
});

test('peerId not derived from the key is refused', async () => {
  const s = await serve();
  try {
    const k = keypair();
    const res = await register(s.base, k, { peerId: keypair().peerId });
    assert.strictEqual(res.status, 401);
    assert.deepStrictEqual(await (await fetch(`${s.base}/peers`)).json(), []);
  } finally { s.close(); }
});

test('signature must cover the addresses and the challenge is single-use', async () => {
  const s = await serve();
  try {
    const k = keypair();
    const bad = await register(s.base, k, { tamper: (sig) => Buffer.alloc(64, 1).toString('base64') });
    assert.strictEqual(bad.status, 401);
    // Replay: reuse a consumed nonce.
    const { nonce } = await (await fetch(`${s.base}/auth/challenge?publicKey=${encodeURIComponent(k.b64)}`)).json();
    const addrs = ['/ip4/5.6.7.8/tcp/1'];
    const body = JSON.stringify({ publicKey: k.b64, peerId: k.peerId, addrs, nonce,
      signature: crypto.sign(null, Buffer.from(registerMessage(nonce, k.peerId, addrs)), k.privateKey).toString('base64') });
    const post = () => fetch(`${s.base}/nodes/register`, { method: 'POST', headers: { 'content-type': 'application/json' }, body });
    assert.strictEqual((await post()).status, 200);
    assert.strictEqual((await post()).status, 401);
  } finally { s.close(); }
});

test('n is capped at 32', async () => {
  const s = await serve({ rateBurst: 1000 });
  try {
    for (let i = 0; i < 40; i++) assert.strictEqual((await register(s.base, keypair())).status, 200);
    assert.strictEqual((await (await fetch(`${s.base}/peers?n=1000`)).json()).length, 32);
    assert.strictEqual((await (await fetch(`${s.base}/peers`)).json()).length, 8);
  } finally { s.close(); }
});

test('oversize body is 413, bad input is 400', async () => {
  const s = await serve();
  try {
    const big = await fetch(`${s.base}/nodes/register`, { method: 'POST', headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ pad: 'x'.repeat(10_000) }) });
    assert.strictEqual(big.status, 413);
    assert.strictEqual((await fetch(`${s.base}/auth/challenge?publicKey=short`)).status, 400);
    const k = keypair();
    assert.strictEqual((await register(s.base, k, { addrs: Array(9).fill('/ip4/1.1.1.1/tcp/1') })).status, 400);
  } finally { s.close(); }
});

test('per-IP rate limit', async () => {
  const s = await serve({ rateBurst: 5, ratePerSecond: 1 });
  try {
    const codes = [];
    for (let i = 0; i < 7; i++) codes.push((await fetch(`${s.base}/peers`)).status);
    assert.deepStrictEqual(codes, [200, 200, 200, 200, 200, 429, 429]);
    s.clock.t += 2000;
    assert.strictEqual((await fetch(`${s.base}/peers`)).status, 200);
    assert.strictEqual((await fetch(`${s.base}/health`)).status, 200); // health is never limited
  } finally { s.close(); }
});
