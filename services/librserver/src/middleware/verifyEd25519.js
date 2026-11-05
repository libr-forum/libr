const { consumeNonce } = require('../utils/nonce');
const { verifyEd25519 } = require('../utils/crypto');
const logger = require('../logger');

/**
 * Express middleware that enforces Ed25519 challenge-response authentication.
 *
 * Expected request body fields:
 *   publicKey  {string} — raw 32-byte Ed25519 public key, base64-encoded
 *   nonce      {string} — the nonce returned by GET /auth/challenge
 *   signature  {string} — Ed25519 signature of the nonce, base64-encoded
 *
 * On success the middleware sets `req.verifiedPublicKey` and calls `next()`.
 * On failure it responds 401.
 */
function verifyEd25519Middleware(req, res, next) {
  const { publicKey, nonce, signature } = req.body || {};

  if (!publicKey || !nonce || !signature) {
    return res.status(400).json({ error: 'publicKey, nonce and signature are required' });
  }

  // Retrieve and consume the stored nonce (one-time use)
  const storedNonce = consumeNonce(publicKey);
  if (!storedNonce) {
    logger.warn({ publicKey }, 'Auth failed: no valid challenge found');
    return res.status(401).json({ error: 'No valid challenge found — request a new challenge first' });
  }

  // The client must present the exact nonce we issued
  if (storedNonce !== nonce) {
    logger.warn({ publicKey }, 'Auth failed: nonce mismatch');
    return res.status(401).json({ error: 'Nonce mismatch' });
  }

  // Verify the Ed25519 signature over the nonce bytes
  if (!verifyEd25519(publicKey, nonce, signature)) {
    logger.warn({ publicKey }, 'Auth failed: invalid Ed25519 signature');
    return res.status(401).json({ error: 'Invalid signature' });
  }

  req.verifiedPublicKey = publicKey;
  next();
}

module.exports = verifyEd25519Middleware;
