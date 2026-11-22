require('dotenv').config();

module.exports = {
  port: parseInt(process.env.PORT || '3000', 10),
  // Seconds a registration lives; nodes re-register at half this.
  ttlMs: parseInt(process.env.ENTRY_TTL_SECONDS || '600', 10) * 1000,
  trustProxy: process.env.TRUST_PROXY === '1',
};
