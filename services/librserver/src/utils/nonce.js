const crypto = require('crypto');
const config = require('../config');

// Map<publicKey, { nonce: string, expiresAt: number, timer: NodeJS.Timeout }>
const store = new Map();

/**
 * Create and store a one-time nonce for the given public key.
 * Any existing pending nonce for that key is cancelled first.
 */
function issueNonce(publicKey) {
  // Cancel previous nonce if any
  const existing = store.get(publicKey);
  if (existing) clearTimeout(existing.timer);

  const nonce = crypto.randomBytes(32).toString('hex');
  const expiresAt = Date.now() + config.nonceTtlMs;

  const timer = setTimeout(() => {
    store.delete(publicKey);
  }, config.nonceTtlMs);

  // Allow the timer not to block process exit
  if (timer.unref) timer.unref();

  store.set(publicKey, { nonce, expiresAt, timer });
  return nonce;
}

/**
 * Consume and return the pending nonce for the given public key.
 * Returns null if no valid nonce exists or if it has expired.
 * Deletes the nonce on success (one-time use).
 */
function consumeNonce(publicKey) {
  const entry = store.get(publicKey);
  if (!entry) return null;
  if (Date.now() > entry.expiresAt) {
    clearTimeout(entry.timer);
    store.delete(publicKey);
    return null;
  }
  clearTimeout(entry.timer);
  store.delete(publicKey);
  return entry.nonce;
}

module.exports = { issueNonce, consumeNonce };
