const { MongoClient } = require('mongodb');
const config = require('../config');
const logger = require('../logger');

let db;
let client;

async function connect() {
  client = new MongoClient(config.mongoUri);
  await client.connect();
  db = client.db('Addrs');
  logger.info('MongoDB connected');
}

/** Returns the shared MongoClient (call after connect()). */
function getClient() {
  return client;
}

// ── Mods allowlist (admin-managed, static) ────────────────────────────────────
// Collection: 'mods'  — contains { publicKey } for every permitted moderator.
// Only admins write to this collection directly in MongoDB.

/**
 * Returns true if the given publicKey exists in the mods allowlist.
 */
async function isModAllowed(publicKey) {
  const doc = await db.collection('mods').findOne({ publicKey }, { projection: { _id: 1 } });
  return doc !== null;
}

// ── Online mods live registry (server-managed, TTL-based) ─────────────────────
// Collection: 'onlinemods' — contains { peerId, publicKey, lastSeen } for mods
// that are currently online. Written by the server on register/refresh/deregister.

function getMods(sinceMs) {
  const filter = sinceMs ? { lastSeen: { $gte: new Date(Date.now() - sinceMs) } } : {};
  return db.collection('onlinemods').find(filter, { projection: { _id: 0 } }).toArray();
}

function upsertMod(peerId, publicKey) {
  return db.collection('onlinemods').updateOne(
    { publicKey },
    { $set: { peerId, publicKey, lastSeen: new Date() } },
    { upsert: true },
  );
}

function touchMod(publicKey) {
  return db.collection('onlinemods').updateOne({ publicKey }, { $set: { lastSeen: new Date() } });
}

function removeMod(publicKey) {
  return db.collection('onlinemods').deleteOne({ publicKey });
}

function removeStaleModsBefore(cutoff) {
  return db.collection('onlinemods').deleteMany({ lastSeen: { $lt: cutoff } });
}

// ── Relays (permanent entries — no TTL expiry) ────────────────────────────────

function getRelays() {
  return db.collection('relays').find({}, { projection: { _id: 0 } }).toArray();
}

function upsertRelay(address, publicKey) {
  return db.collection('relays').updateOne(
    { address },
    { $set: { address, publicKey, lastSeen: new Date() } },
    { upsert: true },
  );
}




// ── Nodes (bootstrap DB nodes — dynamic, TTL-based, same as relays) ─────────
// Collection: 'nodes' — { nodeId, peerId, publicKey, lastSeen }
// DB nodes self-register on startup and refresh periodically.

function getNodes(sinceMs) {
  const filter = sinceMs ? { lastSeen: { $gte: new Date(Date.now() - sinceMs) } } : {};
  return db.collection('nodes').find(filter, { projection: { _id: 0 } }).toArray();
}

function upsertNode(nodeId, peerId, publicKey) {
  return db.collection('nodes').updateOne(
    { peer_id: peerId },
    { $set: { node_id: nodeId, peer_id: peerId, publicKey, lastSeen: new Date() } },
    { upsert: true },
  );
}

function touchNode(publicKey) {
  return db.collection('nodes').updateOne({ publicKey }, { $set: { lastSeen: new Date() } });
}

function removeNode(publicKey) {
  return db.collection('nodes').deleteOne({ publicKey });
}

function removeStaleNodesBefore(cutoff) {
  return db.collection('nodes').deleteMany({ lastSeen: { $lt: cutoff } });
}

module.exports = {
  connect,
  getClient,
  isModAllowed,
  getMods,
  upsertMod,
  touchMod,
  removeMod,
  removeStaleModsBefore,
  getRelays,
  upsertRelay,
  getNodes,
  upsertNode,
  touchNode,
  removeNode,
  removeStaleNodesBefore,
};
