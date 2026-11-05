const { Router } = require('express');
const db = require('../db/mongo');
const analytics = require('../db/analytics');
const { issueNonce } = require('../utils/nonce');
const verifyEd25519 = require('../middleware/verifyEd25519');
const logger = require('../logger');

const router = Router();

// ── Challenge ─────────────────────────────────────────────────────────────────

/**
 * GET /auth/challenge?publicKey=<base64>
 * Issues a one-time 32-byte hex nonce for the given public key.
 * The client must sign this nonce and include it in the subsequent write request.
 * The nonce expires after 60 seconds.
 */
router.get('/auth/challenge', (req, res) => {
  const { publicKey } = req.query;
  if (!publicKey) {
    return res.status(400).json({ error: 'publicKey query parameter is required' });
  }
  const nonce = issueNonce(publicKey);
  logger.debug({ publicKey }, 'Challenge issued');
  res.json({ nonce });
});

// ── Mods ──────────────────────────────────────────────────────────────────────

/**
 * POST /mods/register
 * Body: { peerId, publicKey, nonce, signature }
 * Registers (or refreshes) a mod entry. Sets lastSeen = now.
 */
router.post('/mods/register', verifyEd25519, async (req, res) => {
  const { peerId } = req.body;
  const publicKey = req.verifiedPublicKey;

  if (!peerId) {
    return res.status(400).json({ error: 'peerId is required' });
  }

  try {
    const allowed = await db.isModAllowed(publicKey);
    if (!allowed) {
      logger.warn({ publicKey }, 'Mod registration denied — not in allowlist');
      return res.status(403).json({ error: 'Public key is not in the moderator allowlist' });
    }
    await db.upsertMod(peerId, publicKey);
    analytics.trackUser(publicKey, 'mod');
    analytics.logRegistrationEvent({ publicKey, entityType: 'mod', action: 'register', meta: { peerId } });
    logger.info({ publicKey, peerId }, 'Mod registered');
    res.json({ ok: true });
  } catch (err) {
    logger.error({ err }, 'POST /mods/register error');
    res.status(500).json({ error: 'Internal server error' });
  }
});

/**
 * POST /mods/refresh
 * Body: { publicKey, nonce, signature }
 * Updates lastSeen for an existing mod entry to keep it alive.
 */
router.post('/mods/refresh', verifyEd25519, async (req, res) => {
  const publicKey = req.verifiedPublicKey;
  try {
    const allowed = await db.isModAllowed(publicKey);
    if (!allowed) {
      logger.warn({ publicKey }, 'Mod refresh denied — not in allowlist');
      return res.status(403).json({ error: 'Public key is not in the moderator allowlist' });
    }
    const result = await db.touchMod(publicKey);
    if (result.matchedCount === 0) {
      return res.status(404).json({ error: 'Mod entry not found — register first' });
    }
    analytics.touchUser(publicKey);
    logger.debug({ publicKey }, 'Mod presence refreshed');
    res.json({ ok: true });
  } catch (err) {
    logger.error({ err }, 'POST /mods/refresh error');
    res.status(500).json({ error: 'Internal server error' });
  }
});

/**
 * POST /mods/deregister
 * Body: { publicKey, nonce, signature }
 * Removes the mod entry immediately (best-effort; entries expire anyway on TTL).
 */
router.post('/mods/deregister', verifyEd25519, async (req, res) => {
  const publicKey = req.verifiedPublicKey;
  try {
    const allowed = await db.isModAllowed(publicKey);
    if (!allowed) {
      logger.warn({ publicKey }, 'Mod deregister denied — not in allowlist');
      return res.status(403).json({ error: 'Public key is not in the moderator allowlist' });
    }
    await db.removeMod(publicKey);
    analytics.trackUser(publicKey, 'mod');
    analytics.logRegistrationEvent({ publicKey, entityType: 'mod', action: 'deregister' });
    logger.info({ publicKey }, 'Mod deregistered');
    res.json({ ok: true });
  } catch (err) {
    logger.error({ err }, 'POST /mods/deregister error');
    res.status(500).json({ error: 'Internal server error' });
  }
});

// ── Relays ────────────────────────────────────────────────────────────────────

/**
 * POST /relays/register
 * Body: { address, publicKey, nonce, signature }
 * Registers (or refreshes) a relay entry. Sets lastSeen = now.
 * Open self-registration — any node with a valid keypair may register.
 */
router.post('/relays/register', verifyEd25519, async (req, res) => {
  const { address } = req.body;
  const publicKey = req.verifiedPublicKey;

  if (!address) {
    return res.status(400).json({ error: 'address is required' });
  }
  if (!address.startsWith('/')) {
    return res.status(400).json({ error: 'address must be a valid libp2p multiaddr (starts with /)' });
  }

  try {
    await db.upsertRelay(address.trim(), publicKey);
    analytics.trackUser(publicKey, 'relay');
    analytics.logRegistrationEvent({ publicKey, entityType: 'relay', action: 'register', meta: { address } });
    logger.info({ publicKey, address }, 'Relay registered');
    res.json({ ok: true });
  } catch (err) {
    logger.error({ err, address, publicKey }, 'POST /relays/register error');
    res.status(500).json({ error: 'Internal server error', detail: err.message });
  }
});


// ── Nodes ─────────────────────────────────────────────────────────────────────

/**
 * POST /nodes/register
 * Body: { nodeId, peerId, publicKey, nonce, signature }
 * Open self-registration — any DB node with a valid keypair may register.
 * Sets lastSeen = now.
 */
router.post('/nodes/register', verifyEd25519, async (req, res) => {
  const { nodeId, peerId } = req.body;
  const publicKey = req.verifiedPublicKey;

  if (!nodeId || !peerId) {
    return res.status(400).json({ error: 'nodeId and peerId are required' });
  }

  try {
    await db.upsertNode(nodeId, peerId, publicKey);
    analytics.trackUser(publicKey, 'node');
    analytics.logRegistrationEvent({ publicKey, entityType: 'node', action: 'register', meta: { nodeId, peerId } });
    logger.info({ publicKey, peerId, nodeId }, 'Node registered');
    res.json({ ok: true });
  } catch (err) {
    logger.error({ err, peerId, nodeId }, 'POST /nodes/register error');
    res.status(500).json({ error: 'Internal server error', detail: err.message });
  }
});

/**
 * POST /nodes/refresh
 * Body: { publicKey, nonce, signature }
 * Updates lastSeen for an existing node entry.
 */
router.post('/nodes/refresh', verifyEd25519, async (req, res) => {
  const publicKey = req.verifiedPublicKey;
  try {
    const result = await db.touchNode(publicKey);
    if (result.matchedCount === 0) {
      return res.status(404).json({ error: 'Node entry not found — register first' });
    }
    analytics.touchUser(publicKey);
    logger.debug({ publicKey }, 'Node presence refreshed');
    res.json({ ok: true });
  } catch (err) {
    logger.error({ err }, 'POST /nodes/refresh error');
    res.status(500).json({ error: 'Internal server error' });
  }
});

/**
 * POST /nodes/deregister
 * Body: { publicKey, nonce, signature }
 * Removes the node entry immediately (best-effort; expires naturally on TTL anyway).
 */
router.post('/nodes/deregister', verifyEd25519, async (req, res) => {
  const publicKey = req.verifiedPublicKey;
  try {
    await db.removeNode(publicKey);
    analytics.trackUser(publicKey, 'node');
    analytics.logRegistrationEvent({ publicKey, entityType: 'node', action: 'deregister' });
    logger.info({ publicKey }, 'Node deregistered');
    res.json({ ok: true });
  } catch (err) {
    logger.error({ err }, 'POST /nodes/deregister error');
    res.status(500).json({ error: 'Internal server error' });
  }
});

module.exports = router;
