import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('../../wailsjs/go/main/App', () => ({
  FetchAll: vi.fn(),
  GenerateAlias: vi.fn(async (k: string) => `alias-${k}`),
  GenerateAvatar: vi.fn(async () => '<svg/>'),
  ModAuthentication: vi.fn(),
}));

import * as App from '../../wailsjs/go/main/App';
import { apiService } from './api';

describe('apiService', () => {
  beforeEach(() => vi.clearAllMocks());

  it('maps FetchAll certs to messages', async () => {
    vi.mocked(App.FetchAll).mockResolvedValue([
      { public_key: 'pk1', msg: { content: 'hello', ts: 1700000000 }, mod_certs: [], sign: 's1', deleted: '0' },
    ] as never);
    const msgs = await apiService.getMessages('1');
    expect(msgs).toHaveLength(1);
    expect(msgs[0]).toMatchObject({ content: 'hello', authorPublicKey: 'pk1', authorAlias: 'alias-pk1', sign: 's1', timestamp: 1700000000n });
  });

  it('gives the moderator role only when the backend confirms it', async () => {
    vi.mocked(App.ModAuthentication).mockResolvedValueOnce(true).mockResolvedValueOnce(false);
    expect((await apiService.authenticate('pk')).role).toBe('moderator');
    expect((await apiService.authenticate('pk')).role).toBe('member');
  });
});
