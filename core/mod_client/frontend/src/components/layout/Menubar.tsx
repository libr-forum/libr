import React from 'react';
import { BrowserOpenURL } from '../../../wailsjs/runtime';
import { PencilLine,Globe, Database } from 'lucide-react';
import { logger } from '../../logger/logger';
import {
  GetOnlineMods,
  GenerateAlias,
  GenerateAvatar
} from "../../../wailsjs/go/main/App";

type ModDisplay = {
  key: string;
  alias: string;
  avatarSvg: string;
};

const ComingSoonDialog: React.FC<{ open: boolean; onClose: () => void }> = ({ open, onClose }) => {
  if (!open) return null;
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      <div className="bg-card border border-border/50 rounded-2xl shadow-xl text-foreground p-6 w-[90%] max-w-md flex flex-col ">
        <span className="text-lg font-semibold mb-4 text-libr-secondary">Feature Coming Soon</span>
        <p className="text-muted-foreground mb-6 text-left">
          This feature is not available yet. Stay tuned for updates!
        </p>
        <div className="flex justify-end space-x-2">
          <button
            onClick={onClose}
            className="libr-button bg-muted hover:bg-muted/70 text-foreground px-6 py-2"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
};


export const Menubar: React.FC = () => {
  const [mods, setMods] = React.useState<ModDisplay[]>([]);
  const [dialogOpen, setDialogOpen] = React.useState(false);
  React.useEffect(() => {
    logger.debug('[Menubar] Component mounted.');
    async function fetchMods() {
      try {
        const keys = await GetOnlineMods();
        logger.debug('[Menubar] Received keys:', keys);
        const resolved = await Promise.all(
          keys.map(async (key) => {
            const alias = await GenerateAlias(key);
            const avatarSvg = await GenerateAvatar(key);
            logger.debug(`[Menubar] Processed mod: ${alias}`);
            return { key, alias, avatarSvg };
          })
        );
        setMods(resolved);
        logger.info('[Menubar] Mods loaded successfully.');
      } catch (err) {
        logger.error('[Menubar] Failed to load online mods:', err);
      }
    }

    fetchMods();

    // Only return cleanup function, not async code
    return () => logger.debug('[Menubar] Component unmounted.');
  }, []);

  return (
    <div className="w-full p-2 bg-card shadow-md items-center rounded-3xl h-full flex flex-col z-50">
      <ComingSoonDialog open={dialogOpen} onClose={() => {
        logger.info('[ComingSoonDialog] Closed.');
        setDialogOpen(false)}} />
      {/* Scrollable content */}
      <div className="flex-1 overflow-y-auto flex flex-col w-full items-center">
        <div className="text-left w-full mt-4 mb-4 pl-2 flex items-center">
          <h3 className="text-sm font-semibold text-muted-foreground">
            Moderators
          </h3>
        </div>

        <div className="flex flex-col gap-3 w-full pl-2 pb-4">
          {mods.map(({ key, alias, avatarSvg }) => (
            <div key={key} className="flex items-center justify-start space-x-3 py-1">
              {avatarSvg && avatarSvg !== "unknown" ? (
                <img
                  src={`data:image/svg+xml;base64,${avatarSvg}`}
                  alt="avatar"
                  className="w-10 h-10 rounded-xl"
                />
              ) : (
                <div className="w-10 h-10 bg-libr-accent1 rounded-full flex items-center justify-center">
                  <span className="text-white text-sm font-medium">
                    {alias.charAt(0).toUpperCase()}
                  </span>
                </div>
              )}
              <span className="text-sm font-medium text-foreground">{alias}</span>
            </div>
          ))}
        </div>
      </div>
      <div className="rounded-3xl m-2 bg-card w-[92%]">
        <button
          onClick={() => {
            logger.info('[Menubar] Feedback button clicked.');
            BrowserOpenURL("https://forms.gle/Uchqc6Z49aoJwjvZ9");
          
          }}
          className='flex justify-start hover:bg-muted/50 libr-button w-[100%] items-center space-x-2'
        >
          <PencilLine className="aspect-square h-[40%]" />
          <span className="mt-0.5">Feedback</span>
        </button>
      </div>
      <div className="rounded-3xl m-2 bg-card w-[92%]">
        <button
          onClick={() => {
            logger.info('[Menubar] Website link clicked.');
            BrowserOpenURL("https://libr-ashen.vercel.app/")}}
          className="flex justify-start hover:bg-muted/50 libr-button w-[100%] items-center space-x-2"
        >
          <Globe className="aspect-square h-[40%]" />
          <span className="mt-0.5">Visit Website</span>
        </button>
      </div>
      <div className="rounded-3xl m-2 bg-card w-[92%]">
        <button
          onClick={() => {
            logger.info('[Menubar] Open host database dialog.');
            setDialogOpen(true)}}
          className="flex justify-start libr-button hover:bg-muted/50 w-[100%] items-center space-x-2"
        >
          <Database className="aspect-square h-[40%]" />
          <span className="mt-0.5">Host a database</span>
        </button>
      </div>
    </div>
  );
};
