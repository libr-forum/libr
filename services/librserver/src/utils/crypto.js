/**
 * Ed25519 verification using Node.js built-in crypto module.
 *
 * The Go codebase stores raw 32-byte Ed25519 public keys encoded as base64.
 * Node's crypto API requires SPKI (SubjectPublicKeyInfo) DER format, so we
 * prepend the standard Ed25519 SPKI OID header before importing the key.
 */
const crypto = require('crypto');

// ASN.1 DER header for an Ed25519 SubjectPublicKeyInfo (RFC 8410)
// Hex: 30 2a 30 05 06 03 2b 65 70 03 21 00  (12 bytes)
const ED25519_SPKI_PREFIX = Buffer.from('302a300506032b6570032100', 'hex');

/**
 * Import a raw 32-byte Ed25519 public key (base64-encoded) as a Node KeyObject.
 * Returns null if the input is invalid.
 */
function importPublicKey(publicKeyBase64) {
  try {
    const raw = Buffer.from(publicKeyBase64, 'base64');
    if (raw.length !== 32) return null;
    const spkiDer = Buffer.concat([ED25519_SPKI_PREFIX, raw]);
    return crypto.createPublicKey({ key: spkiDer, format: 'der', type: 'spki' });
  } catch {
    return null;
  }
}

/**
 * Verify an Ed25519 signature.
 *
 * @param {string} publicKeyBase64  - raw 32-byte public key, base64-encoded
 * @param {string} message          - the plaintext message that was signed
 * @param {string} signatureBase64  - the signature, base64-encoded
 * @returns {boolean}
 */
function verifyEd25519(publicKeyBase64, message, signatureBase64) {
  try {
    const keyObject = importPublicKey(publicKeyBase64);
    if (!keyObject) return false;
    const sig = Buffer.from(signatureBase64, 'base64');
    const data = Buffer.from(message);
    // null algorithm — Ed25519 is self-contained (no separate hash step)
    return crypto.verify(null, data, keyObject, sig);
  } catch {
    return false;
  }
}

module.exports = { verifyEd25519 };
