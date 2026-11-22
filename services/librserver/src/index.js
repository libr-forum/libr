const logger = require('./logger');
const config = require('./config');
const app = require('./app');
const { connect } = require('./db/mongo');
const { connectAnalytics } = require('./db/analytics');
const { startExpiryJob } = require('./jobs/expiry');
const { startDailySnapshotJob } = require('./jobs/dailySnapshot');

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

