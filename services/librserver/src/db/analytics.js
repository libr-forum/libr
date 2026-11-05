const { getClient } = require('./mongo');
const logger = require('../logger');

let analyticsDb;

// ── Initialization ────────────────────────────────────────────────────────────

/**
 * Connects to the 'Analytics' database using the shared MongoClient.
 * Must be called after mongo.connect().
 */
async function connectAnalytics() {
    const client = getClient();
    if (!client) {
        throw new Error('MongoClient not available — call mongo.connect() first');
    }
    analyticsDb = client.db('Analytics');

    logger.info('Analytics database connected');
}

// ── Helpers (fire-and-forget — never block the request) ───────────────────────

function safeInsert(collection, doc) {
    analyticsDb.collection(collection).insertOne(doc).catch((err) => {
        logger.warn({ err, collection }, 'Analytics insert failed');
    });
}

// ── User Tracking ─────────────────────────────────────────────────────────────

/**
 * Upserts a user entry by publicKey. Tracks first/last seen and total registrations.
 * Called on register and deregister (not on refreshes).
 *
 * @param {string} publicKey
 * @param {string} entityType  — 'mod' | 'relay' | 'node'
 */
function trackUser(publicKey, entityType) {
    analyticsDb.collection('users').updateOne(
        { publicKey },
        {
            $set: { entityType, lastSeen: new Date() },
            $setOnInsert: { firstSeen: new Date() },
            $inc: { totalRegistrations: 1 },
        },
        { upsert: true },
    ).catch((err) => {
        logger.warn({ err, publicKey }, 'Analytics trackUser failed');
    });
}

/**
 * Updates lastSeen on a user without incrementing registration count.
 * Called on refresh/heartbeat.
 *
 * @param {string} publicKey
 */
function touchUser(publicKey) {
    analyticsDb.collection('users').updateOne(
        { publicKey },
        { $set: { lastSeen: new Date() } },
    ).catch((err) => {
        logger.warn({ err, publicKey }, 'Analytics touchUser failed');
    });
}

// ── Registration Events ───────────────────────────────────────────────────────

/**
 * Logs a registration event (register or deregister).
 *
 * @param {{ publicKey: string, entityType: string, action: string, meta?: object }} data
 */
function logRegistrationEvent({ publicKey, entityType, action, meta }) {
    safeInsert('registration_events', {
        publicKey,
        entityType,
        action,
        meta: meta || {},
        timestamp: new Date(),
    });
}

// ── Daily Snapshots ───────────────────────────────────────────────────────────

/**
 * Takes a snapshot of active users by counting current entries
 * in the operational collections behind the Addrs database.
 * Called periodically by the dailySnapshot job.
 *
 * @param {{ activeMods: number, activeRelays: number, activeNodes: number }} counts
 */
async function takeDailySnapshot({ activeMods, activeRelays, activeNodes }) {
    const today = new Date().toISOString().slice(0, 10); // YYYY-MM-DD
    try {
        const totalUniqueUsers = await analyticsDb.collection('users').countDocuments();
        const startOfDay = new Date(today + 'T00:00:00.000Z');
        const newUsersToday = await analyticsDb.collection('users').countDocuments({
            firstSeen: { $gte: startOfDay },
        });

        // Insert a new snapshot point (Time Series)
        await analyticsDb.collection('daily_snapshots').insertOne({
            date: today,
            activeMods,
            activeRelays,
            activeNodes,
            totalUniqueUsers,
            newUsersToday,
            timestamp: new Date(),
        });
    } catch (err) {
        logger.warn({ err }, 'Analytics takeDailySnapshot failed');
    }
}

// ── Read Helpers (used by analytics routes) ───────────────────────────────────

async function getStats() {
    const [totalUsers, modsCount, relaysCount, nodesCount] = await Promise.all([
        analyticsDb.collection('users').countDocuments(),
        analyticsDb.collection('users').countDocuments({ entityType: 'mod' }),
        analyticsDb.collection('users').countDocuments({ entityType: 'relay' }),
        analyticsDb.collection('users').countDocuments({ entityType: 'node' }),
    ]);

    const today = new Date().toISOString().slice(0, 10);
    const startOfDay = new Date(today + 'T00:00:00.000Z');
    const newUsersToday = await analyticsDb.collection('users').countDocuments({
        firstSeen: { $gte: startOfDay },
    });

    return { totalUsers, mods: modsCount, relays: relaysCount, nodes: nodesCount, newUsersToday };
}

async function getUsers({ entityType, limit = 50, skip = 0 } = {}) {
    const filter = entityType ? { entityType } : {};
    return analyticsDb.collection('users')
        .find(filter, { projection: { _id: 0 } })
        .sort({ lastSeen: -1 })
        .skip(skip)
        .limit(limit)
        .toArray();
}

async function getRegistrationEvents({ limit = 50 } = {}) {
    return analyticsDb.collection('registration_events')
        .find({}, { projection: { _id: 0 } })
        .sort({ timestamp: -1 })
        .limit(limit)
        .toArray();
}

async function getDailySnapshots({ days = 30 } = {}) {
    return analyticsDb.collection('daily_snapshots')
        .find({}, { projection: { _id: 0 } })
        .sort({ date: -1 })
        .limit(days)
        .toArray();
}

module.exports = {
    connectAnalytics,
    trackUser,
    touchUser,
    logRegistrationEvent,
    takeDailySnapshot,
    getStats,
    getUsers,
    getRegistrationEvents,
    getDailySnapshots,
};
