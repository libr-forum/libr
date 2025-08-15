
const isDev = process.env.NODE_ENV === 'development';

export const logger = {
  info: (...args: unknown[]) => {
    if (isDev) console.info('%c[INFO]', 'color: #4caf50; font-weight: bold;', ...args);
  },
  warn: (...args: unknown[]) => {
    if (isDev) console.warn('%c[WARN]', 'color: #ff9800; font-weight: bold;', ...args);
  },
  error: (...args: unknown[]) => {
    console.error('%c[ERROR]', 'color: #f44336; font-weight: bold;', ...args);
  },
  debug: (...args: unknown[]) => {
    if (isDev) console.debug('%c[DEBUG]', 'color: #2196f3; font-weight: bold;', ...args);
  },    
};
