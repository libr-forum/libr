// In-memory seed registry: peerId -> { peerId, addrs, expiresAt }. Seeds are soft
// state that nodes re-register, so nothing needs to survive a restart.
class Registry {
  constructor({ ttlMs, maxEntries, now = Date.now }) {
    this.ttlMs = ttlMs;
    this.maxEntries = maxEntries;
    this.now = now;
    this.entries = new Map();
  }

  // Returns false when full and peerId is new.
  put(peerId, addrs) {
    this.sweep();
    if (!this.entries.has(peerId) && this.entries.size >= this.maxEntries) return false;
    this.entries.set(peerId, { peerId, addrs, expiresAt: this.now() + this.ttlMs });
    return true;
  }

  // Up to n live entries, randomly chosen so load spreads over seeds.
  sample(n) {
    this.sweep();
    const all = [...this.entries.values()];
    for (let i = all.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
      [all[i], all[j]] = [all[j], all[i]];
    }
    return all.slice(0, n).map(({ peerId, addrs }) => ({ peerId, addrs }));
  }

  sweep() {
    const t = this.now();
    for (const [k, e] of this.entries) if (e.expiresAt <= t) this.entries.delete(k);
  }
}

module.exports = { Registry };
