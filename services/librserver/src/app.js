// librserver v2: a bootstrap seed list, nothing more. Nodes register their own
// addresses with a signed, key-bound request; clients fetch a random sample to join
// the DHT. Moderators, relays and communities are not decided here.
const express = require('express');
const pinoHttp = require('pino-http');
const logger = require('./logger');
const { Registry } = require('./registry');
const { Nonces } = require('./utils/nonce');
const { rateLimit } = require('./ratelimit');
const { verifyEd25519 } = require('./utils/crypto');
const { peerIdFromEd25519 } = require('./peerid');

const MAX_ADDRS = 8;
const MAX_ADDR_LEN = 256;
const MAX_PEERS = 32;

// What a node signs to register: binds the challenge, its peer ID and its addresses.
function registerMessage(nonce, peerId, addrs) {
  return ['libr-seed/v2', nonce, peerId, ...addrs].join('\n');
}

function createApp(opts = {}) {
  const now = opts.now || Date.now;
  const registry = new Registry({ ttlMs: opts.ttlMs ?? 10 * 60_000, maxEntries: opts.maxEntries ?? 10_000, now });
  const nonces = new Nonces({ ttlMs: opts.nonceTtlMs ?? 60_000, now });

  const app = express();
  if (opts.trustProxy) app.set('trust proxy', 1);
  if (opts.log !== false) app.use(pinoHttp({ logger }));
  app.get('/health', (_req, res) => res.json({ ok: true }));
  app.use(rateLimit({ perSecond: opts.ratePerSecond ?? 1, burst: opts.rateBurst ?? 30, now }));
  app.use(express.json({ limit: '4kb' }));

  app.get('/auth/challenge', (req, res) => {
    const pk = Buffer.from(String(req.query.publicKey || ''), 'base64');
    if (pk.length !== 32) return res.status(400).json({ error: 'publicKey must be a base64 Ed25519 key' });
    const nonce = nonces.issue(req.query.publicKey);
    if (!nonce) return res.status(503).json({ error: 'too many pending challenges' });
    res.json({ nonce });
  });

  // Body: { publicKey, peerId, addrs[], nonce, signature }. peerId must derive from
  // publicKey, so nobody can register addresses under someone else's identity.
  app.post('/nodes/register', (req, res) => {
    const { publicKey, peerId, addrs, nonce, signature } = req.body || {};
    if (typeof publicKey !== 'string' || typeof peerId !== 'string' || typeof nonce !== 'string' ||
        typeof signature !== 'string' || !Array.isArray(addrs) || addrs.length === 0 || addrs.length > MAX_ADDRS ||
        !addrs.every((a) => typeof a === 'string' && a.startsWith('/') && a.length <= MAX_ADDR_LEN)) {
      return res.status(400).json({ error: `need publicKey, peerId, nonce, signature and 1..${MAX_ADDRS} multiaddrs` });
    }
    const raw = Buffer.from(publicKey, 'base64');
    if (raw.length !== 32 || peerIdFromEd25519(raw) !== peerId) {
      return res.status(401).json({ error: 'peerId does not match publicKey' });
    }
    const issued = nonces.consume(publicKey);
    if (!issued || issued !== nonce || !verifyEd25519(publicKey, registerMessage(nonce, peerId, addrs), signature)) {
      return res.status(401).json({ error: 'bad or expired challenge signature' });
    }
    if (!registry.put(peerId, addrs)) return res.status(503).json({ error: 'registry full' });
    res.json({ ok: true, ttlSeconds: Math.floor(registry.ttlMs / 1000) });
  });

  app.get('/peers', (req, res) => {
    const n = Math.min(Math.max(parseInt(req.query.n, 10) || 8, 1), MAX_PEERS);
    res.json(registry.sample(n));
  });

  return app;
}

module.exports = { createApp, registerMessage };
