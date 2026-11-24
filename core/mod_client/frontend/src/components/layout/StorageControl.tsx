import React from 'react';
import { HardDrive } from 'lucide-react';
import { toast } from 'sonner';
import { GetStorageStatus, SetStorageCommunities, SetStorageContribution } from 'wailsjs/go/main/App';

type Storage = { enabled: boolean; running: boolean; communities?: { name: string; stored: boolean }[] };

// Contribute Storage: run a DB node inside the app, for the communities the user picks.
export const StorageControl: React.FC = () => {
  const [storage, setStorage] = React.useState<Storage>({ enabled: false, running: false });
  const [busy, setBusy] = React.useState(false);

  React.useEffect(() => {
    GetStorageStatus().then(setStorage).catch(() => {});
  }, []);

  const run = async (action: () => Promise<void>, ok?: string) => {
    setBusy(true);
    try {
      await action();
      if (ok) toast.success(ok);
    } catch (error) {
      toast.error('Could not change storage', { description: String(error) });
    } finally {
      setStorage(await GetStorageStatus().catch(() => storage));
      setBusy(false);
    }
  };

  const toggle = () =>
    run(() => SetStorageContribution(!storage.enabled), storage.enabled ? 'Stopped contributing storage' : 'Contributing storage to the network');

  const toggleCommunity = (name: string) => {
    const next = (storage.communities ?? []).map((c) => (c.name === name ? { ...c, stored: !c.stored } : c));
    setStorage({ ...storage, communities: next }); // show the choice at once; the node restarts behind it
    return run(() => SetStorageCommunities(next.filter((c) => c.stored).map((c) => c.name)));
  };

  const stored = (storage.communities ?? []).filter((c) => c.stored).length;
  return (
    <>
      <button
        onClick={toggle}
        disabled={busy}
        className={`mt-3 flex w-full items-start gap-3 rounded-2xl border px-4 py-3 text-left transition-colors ${
          storage.enabled
            ? 'border-libr-accent1/60 bg-libr-accent1/20 text-libr-accent1'
            : 'border-border/60 bg-muted/20 text-foreground hover:bg-muted/40'
        } ${busy ? 'opacity-70 cursor-not-allowed' : ''}`}
      >
        <HardDrive className={`mt-0.5 h-4 w-4 ${storage.enabled ? 'text-libr-accent1' : 'text-muted-foreground'}`} />
        <div className="flex-1">
          <div className="font-medium">
            {busy ? 'Switching…' : storage.enabled ? (storage.running ? 'Contributing Storage' : 'Storage On (starts when online)') : 'Contribute Storage'}
          </div>
          <div className={`mt-1 text-xs ${storage.enabled ? 'text-libr-accent1/80' : 'text-muted-foreground'}`}>
            {storage.enabled
              ? 'This computer stores posts for the network while the app runs.'
              : 'Help keep posts available: run a storage node while the app is open. Uses its own key, not your identity.'}
          </div>
        </div>
      </button>
      {storage.enabled && (storage.communities?.length ?? 0) > 0 && (
        <div className="mt-2 rounded-2xl border border-border/60 bg-muted/10 px-4 py-3 text-sm">
          <div className="mb-2 text-xs text-muted-foreground">Store posts of:</div>
          {storage.communities!.map((c) => (
            <label key={c.name} className="flex cursor-pointer items-center gap-2 py-1">
              <input type="checkbox" checked={c.stored} disabled={busy || (c.stored && stored === 1)} onChange={() => toggleCommunity(c.name)} />
              <span>{c.name}</span>
            </label>
          ))}
        </div>
      )}
    </>
  );
};
