import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('../../wailsjs/go/main/App', () => ({
  FetchAll: vi.fn(),
  ListCommunities: vi.fn(),
  GetModQueue: vi.fn(),
  GetPending: vi.fn(),
  SendInput: vi.fn(),
  ModDecide: vi.fn(),
  GenerateAlias: vi.fn(async (k: string) => `alias-${k}`),
  GenerateAvatar: vi.fn(async () => '<svg/>'),
}));

import * as App from '../../wailsjs/go/main/App';
import { apiService } from './api';
import { useAppStore } from '../store/useAppStore';

const general = { name: 'general', is_mod: false, mods: 4, threshold: 3, mod_keys: ['m1', 'm2', 'm3', 'm4'] };

describe('apiService', () => {
  beforeEach(() => vi.clearAllMocks());

  it('lists communities from the node, with the moderator set', async () => {
    vi.mocked(App.ListCommunities).mockResolvedValue([general, { ...general, name: 'dev', is_mod: true }] as never);
    const cs = await apiService.getCommunities();
    expect(cs.map(c => c.id)).toEqual(['general', 'dev']);
    expect(cs[0]).toMatchObject({ name: 'general', modCount: 4, modKeys: ['m1', 'm2', 'm3', 'm4'], isMod: false, isJoined: true });
    expect(cs[1].isMod).toBe(true);
  });

  it('gives the moderator role per community, only where the key is in the moderator set', async () => {
    vi.mocked(App.ListCommunities).mockResolvedValueOnce([general, { ...general, name: 'dev', is_mod: true }] as never);
    const mod = await apiService.authenticate('pk');
    expect(mod.role).toBe('moderator');
    expect(mod.communityRoles).toEqual({ general: 'member', dev: 'moderator' });
    vi.mocked(App.ListCommunities).mockResolvedValueOnce([general] as never);
    expect((await apiService.authenticate('pk')).role).toBe('member');
  });

  it('maps posts of the requested community to messages', async () => {
    vi.mocked(App.FetchAll).mockResolvedValue([
      { hash: 'h1', msg_id: 'id1', community: 'general', author: 'pk1', ts: 1700000000, content: 'hello', votes: 3, mine: false },
    ] as never);
    const msgs = await apiService.getMessages('general');
    expect(App.FetchAll).toHaveBeenCalledWith('general');
    expect(msgs[0]).toMatchObject({ content: 'hello', authorPublicKey: 'pk1', authorAlias: 'alias-pk1', sign: 'h1', votes: 3, timestamp: 1700000000n, status: 'approved' });
  });

  it('maps send outcomes to message status', async () => {
    useAppStore.setState({ user: { publicKey: 'me', alias: 'me', role: 'member' } });
    for (const [status, want] of [['posted', 'approved'], ['pending', 'pending'], ['rejected', 'rejected']]) {
      vi.mocked(App.SendInput).mockResolvedValueOnce({ status, hash: 'h', ts: 1 } as never);
      expect((await apiService.sendMessage('general', 'x')).status).toBe(want);
    }
    expect(App.SendInput).toHaveBeenCalledWith('general', 'x');
  });

  it('shows only undecided queue items, images as manual_mod, and decides by key', async () => {
    vi.mocked(App.GetModQueue).mockResolvedValue([
      { key: 'post/1', kind: 'post', community: 'general', content: '<img>', reason: '', received: 1, decided: false, approved: false },
      { key: 'report/2', kind: 'report', community: 'dev', content: 'x', reason: 'spam', received: 2, decided: false, approved: false },
      { key: 'post/3', kind: 'post', community: 'general', content: 'y', reason: '', received: 3, decided: true, approved: true },
    ] as never);
    const q = await apiService.getMessageReports();
    expect(q.map(r => [r.sign, r.type])).toEqual([['post/1', 'manual_mod'], ['report/2', 'report']]);
    expect((await apiService.getMessageReports('dev')).map(r => r.sign)).toEqual(['report/2']);
    await apiService.moderateBySign('post/1', 1);
    expect(App.ModDecide).toHaveBeenCalledWith('post/1', true);
  });

  it('maps pending items with vote counts', async () => {
    vi.mocked(App.GetPending).mockResolvedValue([
      { id: 'post-ab', kind: 'post', community: 'general', content: '<img src="data:x">', created: 5, approved: 1, rejected: 0, waiting: 3, status: 'pending' },
    ] as never);
    expect((await apiService.getPending())[0]).toMatchObject({ id: 'post-ab', reason: 'Image attached', approved: 1, awaitingMods: 3, totalMods: 4 });
  });
});
