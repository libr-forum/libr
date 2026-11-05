const { Router } = require('express');
const analytics = require('../db/analytics');
const logger = require('../logger');

const router = Router();

/**
 * GET /analytics/stats
 * Returns aggregate user counts: total, by type, and new users today.
 */
router.get('/analytics/stats', async (_req, res) => {
    try {
        const stats = await analytics.getStats();
        res.json(stats);
    } catch (err) {
        logger.error({ err }, 'GET /analytics/stats error');
        res.status(500).json({ error: 'Internal server error' });
    }
});

/**
 * GET /analytics/users?type=mod|relay|node&limit=50&skip=0
 * Returns user directory sorted by lastSeen descending.
 */
router.get('/analytics/users', async (req, res) => {
    try {
        const { type, limit, skip } = req.query;
        const users = await analytics.getUsers({
            entityType: type || undefined,
            limit: Math.min(parseInt(limit) || 50, 200),
            skip: parseInt(skip) || 0,
        });
        res.json(users);
    } catch (err) {
        logger.error({ err }, 'GET /analytics/users error');
        res.status(500).json({ error: 'Internal server error' });
    }
});

/**
 * GET /analytics/registrations?limit=50
 * Returns recent registration events (register/deregister only, no refreshes).
 */
router.get('/analytics/registrations', async (req, res) => {
    try {
        const { limit } = req.query;
        const events = await analytics.getRegistrationEvents({
            limit: Math.min(parseInt(limit) || 50, 200),
        });
        res.json(events);
    } catch (err) {
        logger.error({ err }, 'GET /analytics/registrations error');
        res.status(500).json({ error: 'Internal server error' });
    }
});

/**
 * GET /analytics/daily?days=30
 * Returns daily snapshot history.
 */
router.get('/analytics/daily', async (req, res) => {
    try {
        const { days } = req.query;
        const snapshots = await analytics.getDailySnapshots({
            days: Math.min(parseInt(days) || 30, 365),
        });
        res.json(snapshots);
    } catch (err) {
        logger.error({ err }, 'GET /analytics/daily error');
        res.status(500).json({ error: 'Internal server error' });
    }
});

module.exports = router;
