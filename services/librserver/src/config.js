require('dotenv').config();

module.exports = {
  port: parseInt(process.env.PORT || '3000', 10),
  mongoUri: process.env.MONGO_URI || 'mongodb+srv://peer:peerhehe@cluster0.vswojqe.mongodb.net/',
  // Seconds before a mod/relay entry is considered stale if not refreshed
  entryTtlMs: parseInt(process.env.ENTRY_TTL_SECONDS || '180', 10) * 1000,
  // How often the expiry job runs
  expiryJobIntervalMs: parseInt(process.env.EXPIRY_JOB_INTERVAL_SECONDS || '90', 10) * 1000,
  // Nonce TTL (ms) — time a challenge is valid before it must be used
  nonceTtlMs: 60_000,
};
