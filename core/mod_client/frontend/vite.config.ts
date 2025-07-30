import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";
import path from "path";
import { componentTagger } from "lovable-tagger";

export default defineConfig(({ mode }) => ({
  server: {
    host: "::",
    port: 8080,
  },
  plugins: [
    react(),
    mode === "development" && componentTagger(),
  ].filter(Boolean),
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
      "wailsjs": path.resolve(__dirname, "./wailsjs"),
    },
    dedupe: ["react", "react-dom"], // ✅ CRITICAL to prevent hook errors
  },
  optimizeDeps: {
    include: ["react", "react-dom"], // ✅ Optional but helps
  },
}));
