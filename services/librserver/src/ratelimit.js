// Per-IP token bucket. Behind a proxy (Render) set TRUST_PROXY=1 so req.ip is the client.
function rateLimit({ perSecond, burst, maxTracked = 50_000, now = Date.now }) {
  const buckets = new Map();
  return (req, res, next) => {
    const t = now();
    let b = buckets.get(req.ip);
    if (!b) {
      // shortcut: forget everyone when full rather than track LRU; fine at seed-server scale
      if (buckets.size >= maxTracked) buckets.clear();
      b = { tokens: burst, at: t };
      buckets.set(req.ip, b);
    }
    b.tokens = Math.min(burst, b.tokens + ((t - b.at) / 1000) * perSecond);
    b.at = t;
    if (b.tokens < 1) return res.status(429).json({ error: 'rate limited' });
    b.tokens -= 1;
    next();
  };
}

module.exports = { rateLimit };
