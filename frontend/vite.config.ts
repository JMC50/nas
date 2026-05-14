import { sveltekit } from "@sveltejs/kit/vite";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig, loadEnv } from "vite";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

// Single source of truth for the application version: repo-root VERSION file.
// Vite injects `__APP_VERSION__` as a global build-time constant so UI code
// (StatusBar, About dialogs) can render the live version without duplicating it.
// scripts/release.sh keeps VERSION and frontend/package.json in lockstep.
function readVersion(): string {
  try {
    return readFileSync(resolve(__dirname, "..", "VERSION"), "utf8").trim();
  } catch {
    return "0.0.0-unknown";
  }
}

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  const serverURL = env.SERVER_URL || "http://localhost:7777";

  return {
    plugins: [tailwindcss(), sveltekit()],
    server: {
      port: 8086,
      host: "0.0.0.0",
      fs: { allow: [".."] },
      proxy: {
        "/server": {
          target: serverURL,
          changeOrigin: true,
          secure: false,
          rewrite: (p) => p.replace(/^\/server/, ""),
        },
      },
    },
    define: {
      "process.env.SERVER_URL": JSON.stringify(serverURL),
      __APP_VERSION__: JSON.stringify(readVersion()),
    },
  };
});
