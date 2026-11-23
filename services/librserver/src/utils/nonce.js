const crypto = require('crypto');

// One pending challenge per public key, valid for ttlMs, used once. Bounded so a
// flood of fresh keys cannot grow memory without limit.
class Nonces {
  constructor({ ttlMs, max = 10_000, now = Date.now }) {
    Object.assign(this, { ttlMs, max, now, store: new Map() });
  }

  issue(publicKey) {
    const t = this.now();
    if (this.store.size >= this.max) {
      for (const [k, e] of this.store) if (e.expiresAt <= t) this.store.delete(k);
      if (this.store.size >= this.max) return null;
    }
    const nonce = crypto.randomBytes(32).toString('hex');
    this.store.set(publicKey, { nonce, expiresAt: t + this.ttlMs });
    return nonce;
  }

  consume(publicKey) {
    const e = this.store.get(publicKey);
    this.store.delete(publicKey);
    if (!e || e.expiresAt <= this.now()) return null;
    return e.nonce;
  }
}

module.exports = { Nonces };
