import { NativeModules, NativeEventEmitter, Platform } from 'react-native';

/**
 * Corresponds to types.Mod in Go.
 */
export interface Mod {
  peer_id: string;
  public_key: string;
}

/**
 * Corresponds to types.ModCert in Go.
 */
export interface ModCert {
  sign: string;
  public_key: string;
  status: string;
}

/**
 * Corresponds to types.Msg in Go.
 */
export interface Msg {
  content: string;
  ts: number;
}

/**
 * Corresponds to types.RetMsgCert in Go.
 */
export interface RetMsgCert {
  public_key: string;
  msg: Msg;
  mod_certs: ModCert[];
  sign: string;
  deleted: string;
}

/**
 * Corresponds to types.MsgCert in Go.
 */
export interface MsgCert {
  public_key: string;
  msg: Msg;
  mod_certs: ModCert[];
  sign: string;
  reason?: string;
  type?: string;
}

/**
 * Corresponds to types.SendResult in Go.
 */
export interface SendResult {
  status: string;
  mod_certs: ModCert[];
  sign: string;
  ts: number;
}

/**
 * LibrCore interface for interacting with the Go bridge.
 */
interface LibrCoreInterface {
  // ── Node lifecycle ───────────────────────────────────────────────────────
  /** Initialize libp2p node with JSON array of relay multiaddresses. */
  initNode(relayAddrsJson: string): Promise<string>;
  /** Get this node's libp2p Peer ID. */
  getPeerID(): Promise<string>;
  /** Stop the libp2p node. */
  stopNode(): Promise<void>;
  /** Send a raw message to a peer (low-level). */
  sendMessage(targetPeerID: string, message: string): Promise<string>;

  // ── App / key management ─────────────────────────────────────────────────
  /** Init keys + discovery client. Returns base64 public key. */
  initApp(serverURL: string): Promise<string>;
  /** Returns the current Ed25519 public key (base64). */
  getPublicKey(): Promise<string>;
  /** Regenerate key pair; returns new public key (base64). */
  regenKeys(): Promise<string>;
  /** Enables a temporary incognito identity and returns its public key. */
  enableIncognito(): Promise<string>;
  /** Restores the primary identity and returns its public key. */
  disableIncognito(): Promise<string>;
  /** Returns whether the app is currently using an incognito identity. */
  isIncognitoEnabled(): Promise<boolean>;

  // ── Discovery ────────────────────────────────────────────────────────────
  /** Returns JSON array of relay multiaddress strings. */
  getRelayAddresses(): Promise<string>;
  /** Returns JSON array of Mod objects. */
  getOnlineMods(): Promise<string>;
  /** Returns true if this node's key is in the mod allowlist. */
  amIMod(): Promise<boolean>;

  // ── Messaging ────────────────────────────────────────────────────────────
  /** Send a plain-text message through moderation → Kademlia. Returns JSON SendResult. */
  sendTextMessage(content: string): Promise<string>;
  /** Fetch messages from the last hour. Returns JSON array of RetMsgCert. */
  fetchMessages(): Promise<string>;
  /** Report a message. msgCertJSON is a JSON-encoded MsgCert. Returns "ok" or "error:...". */
  reportMessage(msgCertJSON: string, reason: string): Promise<string>;
  /** Fetch recent reports. Returns JSON array of ReportCert. */
  fetchReports(): Promise<string>;
  /** Fetch pending reports. Returns JSON array of ReportCert. */
  getPendingReports(): Promise<string>;
  /** Moderate a message. action is "approve" or "reject". Returns "ok" or "error:...". */
  moderateMessage(msgCertJSON: string, action: 'approve' | 'reject'): Promise<string>;
  /** Delete a message. msgCertJSON is a JSON-encoded MsgCert. Returns "ok" or "error:...". */
  deleteMessage(msgCertJSON: string): Promise<string>;
  /** Dismiss / delete a completed moderation or report by signature. */
  dismissReport(msgSign: string): Promise<string>;
  /** Re-send an approved message certificate to the DHT. */
  resendToDb(msgCertJSON: string): Promise<string>;

  // ── Cron & Tasks ────────────────────────────────────────────────────────
  /** Start the background moderation cron job. */
  startCron(): Promise<string>;
  /** Stop the background moderation cron job. */
  stopCron(): Promise<string>;
  /** Perform a one-shot background cron tick. */
  tickCron(): Promise<string>;

  // ── Identity ─────────────────────────────────────────────────────────────
  /** Returns a human-readable alias for a base64 public key. */
  generateAlias(key: string): Promise<string>;
  /** Returns a base64-encoded SVG avatar for a base64 public key. */
  generateAvatar(key: string): Promise<string>;
  /** Show an OS status bar notification. */
  showSystemNotification(title: string, message: string): Promise<string>;
}

const { LibrCore } = NativeModules;

const MockLibrCore: LibrCoreInterface = {
  initNode: async () => { console.warn('[LibrCore mock] initNode'); return 'success'; },
  getPeerID: async () => 'mock_peer_id_12345',
  stopNode: async () => { },
  sendMessage: async (t, m) => { console.log(`[mock] send to ${t}: ${m}`); return 'ACK'; },
  initApp: async () => { console.warn('[LibrCore mock] initApp'); return 'mock_pubkey_base64'; },
  getPublicKey: async () => 'mock_pubkey_base64',
  regenKeys: async () => 'new_mock_pubkey_base64',
  enableIncognito: async () => 'incognito_mock_pubkey_base64',
  disableIncognito: async () => 'mock_pubkey_base64',
  isIncognitoEnabled: async () => false,
  getRelayAddresses: async () => '[]',
  getOnlineMods: async () => '[]',
  amIMod: async () => false,
  sendTextMessage: async (c) => {
    console.log(`[mock] sendTextMessage: ${c}`);
    return JSON.stringify({ status: 'sent', mod_certs: [], sign: 'mock_sign', ts: Date.now() / 1000 });
  },
  fetchMessages: async () => '[]',
  reportMessage: async (json, reason) => {
    console.log('[LibrCore Mock] reportMessage', json, reason);
    return 'ok';
  },
  getPendingReports: async () => '{}',
  fetchReports: async () => '[]',
  moderateMessage: async (json, action) => {
    console.log(`[LibrCore Mock] moderateMessage: ${action}`, json);
    return 'ok';
  },
  deleteMessage: async (json) => {
    console.log('[LibrCore Mock] deleteMessage', json);
    return 'ok';
  },
  dismissReport: async (sign) => {
    console.log('[LibrCore Mock] dismissReport', sign);
    return 'ok';
  },
  resendToDb: async (json) => {
    console.log('[LibrCore Mock] resendToDb', json);
    return 'ok';
  },
  startCron: async () => 'ok',
  stopCron: async () => 'ok',
  tickCron: async () => 'ok',
  generateAlias: async (k) => k.slice(0, 10),
  generateAvatar: async (k) => '',
  showSystemNotification: async (title, message) => {
    console.log('[LibrCore Mock] showSystemNotification:', title, message);
    return 'ok';
  },
};

export const LibrCoreEvents = LibrCore ? new NativeEventEmitter(LibrCore) : null;
export default (LibrCore || MockLibrCore) as LibrCoreInterface;
