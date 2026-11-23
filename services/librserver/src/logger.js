const pino = require('pino');

const logger = pino({
  level: process.env.LOG_LEVEL || 'info',
  // In production Render streams JSON; in local dev add LOG_PRETTY=1 and pipe
  // through `node src/index.js | npx pino-pretty` for human-readable output.
  formatters: {
    level(label) {
      return { level: label };
    },
  },
  timestamp: pino.stdTimeFunctions.isoTime,
  base: { service: 'librserver' },
});

module.exports = logger;
