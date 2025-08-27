import { defineConfig, loadEnv } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");

  const serverURL = env.SERVER_URL || "http://localhost:7777";
  const discordLoginURL =
    env.DISCORD_LOGIN_URL ||
    `https://discord.com/oauth2/authorize?client_id=${env.DISCORD_CLIENT_ID}&response_type=token&redirect_uri=http://localhost:5050/login&scope=identify`;
  const kakaoLoginAPIKEY = env.KAKAO_REST_API_KEY || "your-kakao-api-key";
  const kakaoLoginRedirectURL =
    env.KAKAO_REDIRECT_URI || "http://localhost:5050/kakaoLogin";

  return {
    plugins: [svelte()],
    server: {
      port: 8086,
      host: "0.0.0.0",
      allowedHosts: [
        "admin.i-foto.co.kr",
        "localhost",
        "127.0.0.1"
      ],
      proxy: {
        "/server": {
          target: serverURL,
          changeOrigin: true,
          secure: false, // Allow HTTP for development
          rewrite: (p) => p.replace(/^\/server/, ""),
        },
      },
    },
    define: {
      "process.env.SERVER_URL": JSON.stringify(serverURL),
      "process.env.LOGIN_URL": JSON.stringify(discordLoginURL),
      "process.env.KAKAO_API_KEY": JSON.stringify(kakaoLoginAPIKEY),
      "process.env.KAKAO_REDIRECT_URL": JSON.stringify(kakaoLoginRedirectURL),
    },
  };
});
