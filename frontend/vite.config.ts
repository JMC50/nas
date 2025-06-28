import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

import { serverURL, loginURL, kakaoLoginAPIKEY, kakaoLoginRediectURL } from "./config.local.json"

// https://vite.dev/config/
export default defineConfig({
  plugins: [svelte()],
  server: {
    port: 5050,
    proxy:{
      "/server": {
        target: serverURL,
        changeOrigin: true,
        secure: true,
        rewrite: (p) => p.replace(/^\/server/, ''),
      },
    }
  },
  define: {
    'process.env.SERVER_URL': JSON.stringify(serverURL),
    'process.env.LOGIN_URL': JSON.stringify(loginURL),
    'process.env.KAKAO_API_KEY': JSON.stringify(kakaoLoginAPIKEY),
    'process.env.KAKAO_REDIRECT_URL': JSON.stringify(kakaoLoginRediectURL)
  }
})
