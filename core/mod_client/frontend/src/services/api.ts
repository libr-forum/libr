import { SendInput, FetchAll, GenerateAvatar, GenerateAlias, GetModerationLogs, GetModConfig, SaveModConfig, SaveGoogleApiKey, GetModQueue, ModDecide, GetPending, ListCommunities, GetModKeys, SaveModKey, TestLLMConnection } from "../../wailsjs/go/main/App";
import { Community, Message, User, ModLogEntry, useAppStore, ReportedMessage, PendingModeration } from '../store/useAppStore';
import { app } from '../../wailsjs/go/models';
import { emojify } from 'node-emoji';

export function toCommunity(c: app.Community): Community {
  return {
    id: c.name,
    name: c.name,
    topic: `${c.threshold} of ${c.mods} moderators approve each post`,
    modCount: c.mods,
    modKeys: c.mod_keys ?? [],
    isMod: c.is_mod,
    isJoined: true, // every community is readable; there is no membership list
    requiresApproval: false,
  };
}

export function toStatus(s: string): Message['status'] {
  return s === 'posted' ? 'approved' : s === 'pending' ? 'pending' : 'rejected';
}

export function toPending(p: app.PendingInfo): PendingModeration {
  return {
    id: p.id,
    ts: p.created,
    content: p.content,
    reason: p.kind === 'report' ? p.reason : (/<img\s/i.test(p.content) ? 'Image attached' : ''),
    approved: p.approved,
    rejected: p.rejected,
    totalMods: p.approved + p.rejected + p.waiting,
    ackCount: p.approved + p.rejected,
    awaitingMods: p.waiting,
  };
}

export const apiService = {
  // The role is per community: moderator where the user's key is in the moderator set.
  async authenticate(publicKey: string): Promise<User> {
    const [avatar, alias, communities] = await Promise.all([
      GenerateAvatar(publicKey), GenerateAlias(publicKey), ListCommunities(),
    ]);
    const communityRoles: Record<string, 'member' | 'moderator'> = {};
    for (const c of communities ?? []) communityRoles[c.name] = c.is_mod ? 'moderator' : 'member';
    const role = Object.values(communityRoles).includes('moderator') ? 'moderator' : 'member';
    return { publicKey, alias, role, communityRoles, avatarSvg: avatar };
  },

  async getCommunities(): Promise<Community[]> {
    return ((await ListCommunities()) ?? []).map(toCommunity);
  },

  async getMessages(communityId: string): Promise<Message[]> {
    try {
      const posts = (await FetchAll(communityId)) ?? [];
      return await Promise.all(
        posts.map(async (p) => ({
          content: p.content,
          authorAlias: await GenerateAlias(p.author),
          authorPublicKey: p.author,
          timestamp: BigInt(p.ts),
          communityId: p.community,
          status: 'approved',
          votes: p.votes,
          avatarSvg: await GenerateAvatar(p.author),
          sign: p.hash,
        } as Message))
      );
    } catch (err) {
      console.error("Failed to fetch messages:", err);
      return [];
    }
  },

  // Human review queue of this moderator: image posts and reports.
  async getMessageReports(communityId?: string): Promise<ReportedMessage[]> {
    const items = (await GetModQueue()) ?? [];
    return items
      .filter((it) => !it.decided && (!communityId || it.community === communityId))
      .map((it) => ({
        sign: it.key,
        content: it.content,
        note: it.reason || '',
        type: it.kind === 'post' ? 'manual_mod' : 'report',
        communityId: it.community,
      }));
  },

  // value 1: publish the post / remove the reported post; 0: the opposite.
  async moderateBySign(key: string, value: number): Promise<void> {
    await ModDecide(key, value === 1);
  },

  async getPending(): Promise<PendingModeration[]> {
    return ((await GetPending()) ?? []).map(toPending);
  },

  async sendMessage(communityId: string, content: string): Promise<Message> {
    const result = await SendInput(communityId, content);
    const user = useAppStore.getState().user!;
    return {
      content,
      authorPublicKey: user.publicKey,
      authorAlias: user.alias,
      timestamp: BigInt(result.ts > 0 ? result.ts : Math.floor(Date.now() / 1000)),
      communityId,
      status: toStatus(result.status),
      avatarSvg: user.avatarSvg,
      sign: result.hash || result.id || '',
    };
  },

  async getModerationLogs(): Promise<ModLogEntry[]> {
    return await GetModerationLogs();
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