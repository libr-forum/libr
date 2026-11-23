import { NativeModules, NativeEventEmitter } from 'react-native';

// ── v2 types (JSON from the Go bridge, core/libr/app) ─────────────────────────

/** app.Community */
export interface Community {
  name: string;
  is_mod: boolean;
  mods: number;
  threshold: number;
  mod_keys: string[];
}

/** app.Post: one verified post. */
export interface Post {
  hash: string;
  msg_id: string;
  community: string;
  author: string; // base64 Ed25519 key
  ts: number;
  content: string;
  votes: number;
  mine: boolean;
}

/** app.Outcome of post and report. */
export interface Outcome {
  status: 'posted' | 'removed' | 'pending' | 'rejected';
  hash?: string;
  ts?: number;
  id?: string; // pending item, while status is pending
}

/** moderator.Item: an entry of the moderator's human review queue. */
export interface QueueItem {
  key: string;
  kind: 'post' | 'report';
  cert: { msg: { community: string; author: string; ts: number; content: string } };
  reason?: string;
  received: number;
  vote?: { approve: boolean } | null;
}

/** Per-item status of the user's pending posts and reports. */
export interface PendingStatus {
  total: number;
  approved: number;
  rejected: number;
  final_status?: 'approved' | 'rejected';
  type?: 'report' | 'manual_mod';
}

// ── Shapes the screens render ─────────────────────────────────────────────────

export interface Msg {
  content: string;
  ts: number;
}

/** A feed entry as the screens render it; sign is the post hash. */
export interface RetMsgCert {
  public_key: string;
  msg: Msg;
  mod_certs: unknown[];
  sign: string;
  deleted: string;
  votes?: number;
}

export interface SendResult {
  status: string; // sent | pending_manual | rejected | error:...
  sign: string;
  ts: number;
}

/** go/bridge returns the post's content as text; images are inline data URLs. */
export function toRetMsgCert(p: Post): RetMsgCert {
  return { public_key: p.author, msg: { content: p.content, ts: p.ts }, mod_certs: [], sign: p.hash, deleted: '0', votes: p.votes };
}

/** Bridge strings are JSON on success and "error:..." on failure. */
export function parse<T>(raw: string): T {
  if (!raw || raw.startsWith('error:')) throw new Error(raw ? raw.slice(6) : 'empty reply from the LIBR core');
  return JSON.parse(raw) as T;
}

export function toSendResult(raw: string): SendResult {
  if (!raw || raw.startsWith('error:')) return { status: raw || 'error:empty reply', sign: '', ts: 0 };
  const o: Outcome = JSON.parse(raw);
  const status = o.status === 'posted' ? 'sent' : o.status === 'pending' ? 'pending_manual' : o.status;
  return { status, sign: o.hash || o.id || '', ts: o.ts ?? Math.floor(Date.now() / 1000) };
}

// Expo inlines EXPO_PUBLIC_* only where written literally as process.env.NAME.
const buildEnv = {
  EXPO_PUBLIC_LIBR_RELAYS: process.env.EXPO_PUBLIC_LIBR_RELAYS,
  EXPO_PUBLIC_LIBR_BOOTSTRAP: process.env.EXPO_PUBLIC_LIBR_BOOTSTRAP,
  EXPO_PUBLIC_LIBR_SEED_URLS: process.env.EXPO_PUBLIC_LIBR_SEED_URLS,
  EXPO_PUBLIC_LIBR_COMMUNITY: process.env.EXPO_PUBLIC_LIBR_COMMUNITY,
};

/** Network config for startNode, from EXPO_PUBLIC_LIBR_* at build time. */
export function nodeConfig(env: Record<string, string | undefined> = buildEnv): string {
  const list = (v?: string) => (v ?? '').split(',').map(s => s.trim()).filter(Boolean);
  const community = env.EXPO_PUBLIC_LIBR_COMMUNITY; // a genesis file (JSON) for dev networks
  return JSON.stringify({
    relays: list(env.EXPO_PUBLIC_LIBR_RELAYS),
    bootstrap: list(env.EXPO_PUBLIC_LIBR_BOOTSTRAP),
    seed_urls: list(env.EXPO_PUBLIC_LIBR_SEED_URLS),
    communities: community ? [JSON.parse(community)] : [],
  });
}

// ── Native module ─────────────────────────────────────────────────────────────

interface LibrCoreInterface {
  /** Load or create keys; returns the base64 public key. */
  initApp(): Promise<string>;
  getPublicKey(): Promise<string>;
  regenKeys(): Promise<string>;
  enableIncognito(): Promise<string>;
  disableIncognito(): Promise<string>;
  isIncognitoEnabled(): Promise<boolean>;

  /** Join the network; configJSON from nodeConfig(). Resolves "ok" or "error:...". */
  startNode(configJSON: string): Promise<string>;
  stopNode(): Promise<void>;
  getPeerID(): Promise<string>;
  /** JSON {online, peers, routing_table, reservations}. */
  networkStatus(): Promise<string>;
  /** JSON Community[]. */
  listCommunities(): Promise<string>;
  isMod(community: string): Promise<boolean>;

  /** JSON Outcome. */
  post(community: string, content: string): Promise<string>;
  /** JSON Post[] (last 24 h). */
  feed(community: string): Promise<string>;
  /** JSON Outcome. */
  report(hash: string, reason: string): Promise<string>;
  /** "ok" or "error:...". */
  deletePost(hash: string): Promise<string>;
  /** Retry pending posts and reports now. */
  tick(): Promise<string>;
  /** JSON Record<key, PendingStatus>: key is the reported post's hash, or a pending post's id. */
  pendingStatus(): Promise<string>;
  dismissReport(key: string): Promise<string>;

  /** JSON QueueItem[] (moderators only). */
  modQueue(): Promise<string>;
  modDecide(key: string, decision: 'approve' | 'reject'): Promise<string>;

  generateAlias(key: string): Promise<string>;
  generateAvatar(key: string): Promise<string>;
  showSystemNotification(title: string, message: string): Promise<string>;
}

const { LibrCore } = NativeModules;

const general: Community = { name: 'general', is_mod: false, mods: 4, threshold: 3, mod_keys: [] };

const MockLibrCore: LibrCoreInterface = {
  initApp: async () => 'mock_pubkey_base64',
  getPublicKey: async () => 'mock_pubkey_base64',
  regenKeys: async () => 'new_mock_pubkey_base64',
  enableIncognito: async () => 'incognito_mock_pubkey_base64',
  disableIncognito: async () => 'mock_pubkey_base64',
  isIncognitoEnabled: async () => false,
  startNode: async () => 'ok',
  stopNode: async () => { },
  getPeerID: async () => 'mock_peer_id',
  networkStatus: async () => JSON.stringify({ online: true, peers: 0, routing_table: 0, reservations: 0 }),
  listCommunities: async () => JSON.stringify([general]),
  isMod: async () => false,
  post: async () => JSON.stringify({ status: 'posted', hash: 'mock_hash', ts: Math.floor(Date.now() / 1000) }),
  feed: async () => '[]',
  report: async () => JSON.stringify({ status: 'pending', id: 'mock' }),
  deletePost: async () => 'ok',
  tick: async () => 'ok',
  pendingStatus: async () => '{}',
  dismissReport: async () => 'ok',
  modQueue: async () => '[]',
  modDecide: async () => 'ok',
  generateAlias: async (k) => k.slice(0, 10),
  generateAvatar: async () => '',
  showSystemNotification: async () => 'ok',
};

export const LibrCoreEvents = LibrCore ? new NativeEventEmitter(LibrCore) : null;
export default (LibrCore || MockLibrCore) as LibrCoreInterface;
