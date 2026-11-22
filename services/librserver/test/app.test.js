const { test, before, after } = require('node:test');
const assert = require('node:assert');
const app = require('../src/app');

let server;
let base;
before(async () => {
  server = app.listen(0);
  await new Promise((r) => server.once('listening', r));
  base = `http://127.0.0.1:${server.address().port}`;
});
after(() => server.close());

test('GET /health', async () => {
  const res = await fetch(`${base}/health`);
  assert.strictEqual(res.status, 200);
  assert.deepStrictEqual(await res.json(), { ok: true });
});

test('analytics refused without token', async () => {
  delete process.env.ANALYTICS_TOKEN;
  assert.strictEqual((await fetch(`${base}/analytics/stats`)).status, 401);
  process.env.ANALYTICS_TOKEN = 'secret';
  assert.strictEqual((await fetch(`${base}/analytics/registrations`)).status, 401);
  const wrong = await fetch(`${base}/analytics/stats`, { headers: { authorization: 'Bearer nope' } });
  assert.strictEqual(wrong.status, 401);
});
