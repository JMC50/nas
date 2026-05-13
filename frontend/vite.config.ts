import { sveltekit } from "@sveltejs/kit/vite";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig, loadEnv } from "vite";

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
    },
  };
});
