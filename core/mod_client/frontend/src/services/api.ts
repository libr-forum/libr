import { SendInput, SendImageInput, FetchAll, GenerateAvatar, GenerateAlias, GetModerationLogs, GetModConfig, SaveModConfig, ModAuthentication, SaveGoogleApiKey, ManualModerate, FetchMessageReports, ModerateBySign, IsModerationCronRunning, GetPendingModerationStats, ResendToDb, GetModKeys, SaveModKey, TestLLMConnection } from "../../wailsjs/go/main/App";
import axios from 'axios';
import { Community, Message, User, ModLogEntry, useAppStore, ReportedMessage } from '../store/useAppStore';
import { types } from '../../wailsjs/go/models'
import { emojify } from 'node-emoji';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:3000/api';

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
});

// Mock data for demo purposes
const mockCommunities: Community[] = [
  {
    id: '0',
    name: 'libr',
    topic: '',
    memberCount: 1247,
    isJoined: true,
    requiresApproval: false,
  },
  {
    id: '1',
    name: 'cryptography',
    topic: 'Discussion about cryptographic protocols and implementations',
    memberCount: 1247,
    isJoined: false,
    requiresApproval: false,
  },
  {
    id: '2',
    name: 'privacy-tech',
    topic: 'Privacy technologies and digital rights',
    memberCount: 892,
    isJoined: false,
    requiresApproval: false,
  },
  {
    id: '3',
    name: 'decentralized-web',
    topic: 'Building the decentralized internet',
    memberCount: 567,
    isJoined: false,
    requiresApproval: true,
  },
  {
    id: '4',
    name: 'blockchain-dev',
    topic: 'Blockchain development and smart contracts',
    memberCount: 2341,
    isJoined: false,
    requiresApproval: false,
  },
  {
    id: '5',
    name: 'web3-gaming',
    topic: 'Web3 gaming and NFT discussions',
    memberCount: 876,
    isJoined: false,
    requiresApproval: true,
  },
  {
    id: '6',
    name: 'defi-protocols',
    topic: 'Decentralized Finance protocols and analysis',
    memberCount: 1654,
    isJoined: false,
    requiresApproval: false,
  },
  {
    id: '7',
    name: 'zero-knowledge',
    topic: 'Zero-knowledge proofs and privacy solutions',
    memberCount: 432,
    isJoined: false,
    requiresApproval: true,
  },
  {
    id: '8',
    name: 'dao-governance',
    topic: 'DAO governance and tokenomics',
    memberCount: 987,
    isJoined: false,
    requiresApproval: false,
  },
  {
    id: '9',
    name: 'layer2-scaling',
    topic: 'Layer 2 scaling solutions and optimizations',
    memberCount: 723,
    isJoined: false,
    requiresApproval: true,
  },
  {
    id: '10',
    name: 'censorship-resistance',
    topic: 'Building censorship-resistant systems',
    memberCount: 1123,
    isJoined: false,
    requiresApproval: false,
  },
  {
    id: '11',
    name: 'peer-to-peer',
    topic: 'P2P networking and distributed systems',
    memberCount: 645,
    isJoined: false,
    requiresApproval: true,
  },
  {
    id: '12',
    name: 'digital-identity',
    topic: 'Self-sovereign identity and digital credentials',
    memberCount: 498,
    isJoined: false,
    requiresApproval: false,
  },
];

export const apiService = {
  // Auth
  async authenticate(publicKey: string): Promise<User> {
    const avatar = await GenerateAvatar(publicKey)
    const alias = await GenerateAlias(publicKey)
    const isMod = await ModAuthentication(publicKey)
    const role: 'member' | 'moderator' | 'admin' = isMod ? 'moderator' : 'member';
    return {
      publicKey,
      alias,
      role,
      avatarSvg: avatar,
    };
  },

  // Communities
  async getCommunities(): Promise<Community[]> {
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 500));
    return mockCommunities;
  },

  async joinCommunity(communityId: string): Promise<void> {
    await new Promise(resolve => setTimeout(resolve, 300));
    console.log(`Joined community: ${communityId}`);
  },

  // Messages
  async getMessages(_communityId: string): Promise<Message[]> {
    try {
      const fetched = await FetchAll();
      // No caching: always return fresh mapped messages
      return await Promise.all(
        fetched.map(async (message: any) => {
          const alias = await GenerateAlias(message.public_key);
          const svg = await GenerateAvatar(message.public_key);
          return {
            content: message.msg.content,
            authorAlias: alias,
            authorPublicKey: message.public_key,
            timestamp: BigInt(message.msg.ts),
            communityId: "1",
            status: "approved",
            moderationNote: message.mod_certs,
            avatarSvg: svg,
            sign: message.sign,
          } as Message;
        })
      );
    } catch (err) {
      console.error("Failed to fetch messages:", err);
      return [];
    }
  },

  async getMessageReports(_communityId: string): Promise<ReportedMessage[]> {
    try {
      const reports = await FetchMessageReports();
      return reports.map((msg: any) => ({
        sign: msg.sign,
        content: msg.msg.content,
        note: msg.reason || '',
        type: msg.type,
        communityId: _communityId,
      }));
    } catch (err) {
      console.error("Failed to fetch messages:", err);
      return [];
    }
  },

  async manualModerate(cert: types.MsgCert, moderated: number): Promise<void> {
    try {
      await ManualModerate(cert, moderated);
    } catch (err) {
      console.error("manualModerate error:", err);
      throw err;
    }
  },

  async moderateBySign(sign: string, moderated: number): Promise<void> {
    await ModerateBySign(sign, moderated);
  },

  async isModerationCronRunning(): Promise<boolean> {
    return await IsModerationCronRunning();
  },

  async getPendingModerationStats(): Promise<{
    items: { msg_sign: string; ts: number; content: string; reason: string; approved: number; rejected: number; awaiting: number; is_image: boolean }[];
    cron_active: boolean;
  }> {
    return await GetPendingModerationStats();
  },

  async sendMessage(communityId: string, content: string): Promise<Message> {
    const isImage = /<img\s+[^>]*src="data:image/i.test(content);
    const result = isImage ? await SendImageInput(content) : await SendInput(content);

    const approved = result.status === 'sent';
    // For image messages: ANY non-rejected outcome should be treated as pending.
    // This covers: 'pending_manual', 'timeout', 'No online moderators available', 'No moderators available'
    // The pending_mods cron will track it and complete it once a mod comes online.
    const storageFailed = result.status === 'storage_failed';
    const isImagePending = isImage && result.status !== 'error' && result.status !== 'offline';
    const timeout = result.status === 'timeout';
    const modcerts: types.ModCert[] = result.mod_certs ?? [];

    const user = useAppStore.getState().user;

    // No caching: just return the new message object
    const newMessage: Message = {
      content,
      authorPublicKey: user.publicKey,
      authorAlias: user.alias,
      timestamp: BigInt(result.ts > 0 ? result.ts : Math.floor(Date.now() / 1000)),
      communityId,
      status: approved ? 'approved' : (timeout || isImagePending || storageFailed) ? 'pending' : 'rejected',
      avatarSvg: user.avatarSvg,
      moderationNote: modcerts,
      sign: result.sign ?? '',
    };

    return newMessage;
  },

  // Moderation
  async getModerationLogs(): Promise<ModLogEntry[]> {
    const logs = await GetModerationLogs();
    return logs;
  },

  async moderateMessage(action: 'approve' | 'reject', note?: string): Promise<void> {
    await new Promise(resolve => setTimeout(resolve, 300));
    console.log(`Message ${action}ed with note: ${note}`);
  },

  async GetModConfig(): Promise<{
    forbidden: string[];
    thresholds: string;
    image_auto_mod?: boolean;
    llm_enabled?: boolean;
    llm_provider?: string;
    llm_model?: string;
  }> {
    return await GetModConfig();
  },

  async SaveModConfig(data: {
    forbidden: string[];
    thresholds: string;
    image_auto_mod?: boolean;
    llm_enabled?: boolean;
    llm_provider?: string;
    llm_model?: string;
  }): Promise<void> {
    await SaveModConfig(data as any);
  },

  async SaveGoogleApiKey(key: string) {
    await SaveGoogleApiKey(key);
  },

  async GetModKeys(): Promise<Record<string, string>> {
    return await GetModKeys();
  },

  async SaveModKey(keyName: string, keyValue: string): Promise<void> {
    await SaveModKey(keyName, keyValue);
  },

  async TestLLMConnection(provider: string, model: string, apiKey: string): Promise<string> {
    return await TestLLMConnection(provider, model, apiKey);
  },

  async resendToDb(message: Message): Promise<Message> {
    const msg: types.Msg = {
      content: message.content,
      ts: Number(message.timestamp),
    };
    const msgcert = new types.MsgCert({
      public_key: message.authorPublicKey,
      msg: msg,
      mod_certs: message.moderationNote,
      sign: message.sign,
    });

    const result = await ResendToDb(msgcert);
    const approved = result.status === 'sent';
    const storageFailed = result.status === 'storage_failed';

    const user = useAppStore.getState().user;

    return {
      ...message,
      status: approved ? 'approved' : storageFailed ? 'pending' : 'rejected',
    };
  }
};

export function parseFormatting(text: string): string {
  // Escape HTML to prevent injection
  const escapeHTML = (str: string) =>
    str.replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;');

  // Apply emoji replacements first
  let formatted = emojify(text);

  // Code blocks (```...```)
  formatted = formatted.replace(/```([\s\S]*?)```/g, (_match, code) => {
    return `<pre class="bg-muted rounded p-2 overflow-x-auto my-2 text-xs"><code>${escapeHTML(code)}</code></pre>`;
  });

  // Inline code (`...`)
  formatted = formatted.replace(/`([^`\n]+?)`/g, (_match, code) => {
    return `<code class="bg-muted px-1 rounded text-xs">${escapeHTML(code)}</code>`;
  });

  // Bold (**bold**)
  formatted = formatted.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');

  // Italic (*italic*)
  formatted = formatted.replace(/\*(.+?)\*/g, '<em>$1</em>');

  // Underline (_underline_)
  formatted = formatted.replace(/_(.+?)_/g, '<u>$1</u>');

  // Strikethrough (~strike~)
  formatted = formatted.replace(/~(.+?)~/g, '<s>$1</s>');

  // Newlines to <br/>
  formatted = formatted.replace(/\n/g, '<br/>');

  return formatted;
}