import { createRoot } from 'react-dom/client'
import App from './App.tsx'
import './index.css'
import {
  SendInput,
  FetchAll,
  FetchTimestamp,
  Connect,
  GetRelayStatus
} from "../wailsjs/go/main/App";


createRoot(document.getElementById("root")!).render(<App />);
