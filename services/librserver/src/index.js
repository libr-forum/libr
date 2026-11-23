const logger = require('./logger');
const config = require('./config');
const { createApp } = require('./app');

createApp({ ttlMs: config.ttlMs, trustProxy: config.trustProxy }).listen(config.port, '0.0.0.0', () => {
  logger.info({ port: config.port }, 'librserver (seed) started');
});
