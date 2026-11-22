const express = require('express');
const pinoHttp = require('pino-http');
const logger = require('./logger');
const readRoutes = require('./routes/read');
const writeRoutes = require('./routes/write');
const analyticsRoutes = require('./routes/analytics');

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

module.exports = app;
