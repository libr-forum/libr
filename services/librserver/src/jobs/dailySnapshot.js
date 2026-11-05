const db = require('../db/mongo');
const analytics = require('../db/analytics');
const config = require('../config');
const logger = require('../logger');

/**
 * Takes an hourly snapshot of active mods, relays, and nodes.
 * Counts are based on the operational Addrs database.
 * Results are stored in Analytics.daily_snapshots.
 */
function startDailySnapshotJob() {
    const INTERVAL_MS = 60 * 60 * 1000; // 1 hour

    const interval = setInterval(async () => {
        try {
            const [mods, relays, nodes] = await Promise.all([
                db.getMods(config.entryTtlMs),
                db.getRelays(),
                db.getNodes(config.entryTtlMs),
            ]);

            await analytics.takeDailySnapshot({
                activeMods: mods.length,
                activeRelays: relays.length,
                activeNodes: nodes.length,
            });

            logger.info(
                { activeMods: mods.length, activeRelays: relays.length, activeNodes: nodes.length },
                'Daily snapshot updated',
            );
        } catch (err) {
            logger.error({ err }, 'Daily snapshot job error');
        }
    }, INTERVAL_MS);

    if (interval.unref) interval.unref();

    logger.info('Daily snapshot job started (runs every hour)');

    // Take the first snapshot immediately on startup
    setImmediate(async () => {
        try {
            const [mods, relays, nodes] = await Promise.all([
                db.getMods(config.entryTtlMs),
                db.getRelays(),
                db.getNodes(config.entryTtlMs),
            ]);
            await analytics.takeDailySnapshot({
                activeMods: mods.length,
                activeRelays: relays.length,
                activeNodes: nodes.length,
            });
            logger.info('Initial daily snapshot taken');
        } catch (err) {
            logger.warn({ err }, 'Initial daily snapshot failed');
        }
    });
}

module.exports = { startDailySnapshotJob };
