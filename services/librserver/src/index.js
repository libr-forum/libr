const express = require('express');
const pinoHttp = require('pino-http');
const logger = require('./logger');
const config = require('./config');
const { connect } = require('./db/mongo');
const { connectAnalytics } = require('./db/analytics');
const readRoutes = require('./routes/read');
const writeRoutes = require('./routes/write');
const analyticsRoutes = require('./routes/analytics');
const { startExpiryJob } = require('./jobs/expiry');
const { startDailySnapshotJob } = require('./jobs/dailySnapshot');

const app = express();

// Structured HTTP request logging (every request gets a log line with method,
// url, statusCode, responseTime)
app.use(pinoHttp({ logger }));
app.use(express.json());

// ── Health check (used by Render) ─────────────────────────────────────────────
app.get('/health', (_req, res) => res.json({ ok: true }));

// ── Routes ────────────────────────────────────────────────────────────────────
// Public read-only endpoints — no authentication required
app.use('/', readRoutes);

// Authenticated write endpoints — Ed25519 challenge-response required
app.use('/', writeRoutes);

// Analytics endpoints — read-only
app.use('/', analyticsRoutes);

// ── Startup ───────────────────────────────────────────────────────────────────
async function main() {
  await connect();
  await connectAnalytics();
  startExpiryJob();
  startDailySnapshotJob();
  app.listen(config.port, '0.0.0.0', () => {
    logger.info({ port: config.port }, 'librserver started');
  });
}

main().catch((err) => {
  logger.fatal({ err }, 'Failed to start server');
  process.exit(1);
});

