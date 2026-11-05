const { Router } = require('express');
const db = require('../db/mongo');
const config = require('../config');
const logger = require('../logger');

const router = Router();

/**
 * GET /mods
 * Returns all mod entries whose lastSeen is within the configured TTL.
 * Response: [{ peerId, publicKey, lastSeen }]
 */
router.get('/mods', async (_req, res) => {
  try {
    const mods = await db.getMods(config.entryTtlMs);
    res.json(mods);
  } catch (err) {
    logger.error({ err }, 'GET /mods error');
    res.status(500).json({ error: 'Internal server error' });
  }
});

/**
 * GET /mods/check?publicKey=<base64>
 * Checks whether a public key is in the mod allowlist (mods collection).
 * Used by mod_client at startup to decide if it is a moderator.
 * No authentication required — the allowlist is not sensitive.
 * Response: { allowed: true|false }
 */
router.get('/mods/check', async (req, res) => {
  const { publicKey } = req.query;
  if (!publicKey) {
    return res.status(400).json({ error: 'publicKey query parameter is required' });
  }
  try {
    const allowed = await db.isModAllowed(publicKey);
    res.json({ allowed });
  } catch (err) {
    logger.error({ err }, 'GET /mods/check error');
    res.status(500).json({ error: 'Internal server error' });
  }
});

/**
 * GET /relays
 * Returns all relay entries whose lastSeen is within the configured TTL.
 * Response: [{ address, publicKey, lastSeen }]
 */
router.get('/relays', async (_req, res) => {
  try {
    const relays = await db.getRelays();
    res.json(relays);
  } catch (err) {
    logger.error({ err }, 'GET /relays error');
    res.status(500).json({ error: 'Internal server error' });
  }
});

/**
 * GET /nodes
 * Returns all DB node entries whose lastSeen is within the configured TTL.
 * Response: [{ nodeId, peerId, publicKey, lastSeen }]
 */
router.get('/nodes', async (_req, res) => {
  try {
    const nodes = await db.getNodes(config.entryTtlMs);
    res.json(nodes);
  } catch (err) {
    logger.error({ err }, 'GET /nodes error');
    res.status(500).json({ error: 'Internal server error' });
  }
});

module.exports = router;
