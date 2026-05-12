import { sveltekit } from "@sveltejs/kit/vite";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig, loadEnv } from "vite";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  const serverURL = env.SERVER_URL || "http://localhost:7777";
  const discordLoginURL =
    env.DISCORD_LOGIN_URL ||
    `https://discord.com/oauth2/authorize?client_id=${env.DISCORD_CLIENT_ID}&response_type=token&redirect_uri=http://localhost:5050/login&scope=identify`;

  return {
    plugins: [tailwindcss(), sveltekit()],
    server: {
      port: 8086,
      host: "0.0.0.0",
      allowedHosts: ["admin.i-foto.co.kr", "localhost", "127.0.0.1"],
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
      "process.env.LOGIN_URL": JSON.stringify(discordLoginURL),
      "process.env.GOOGLE_CLIENT_ID": JSON.stringify(env.GOOGLE_CLIENT_ID ?? ""),
      "process.env.GOOGLE_REDIRECT_URI": JSON.stringify(env.GOOGLE_REDIRECT_URI ?? ""),
    },
  };
});
