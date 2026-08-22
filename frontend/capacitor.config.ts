import type { CapacitorConfig } from "@capacitor/cli";

const config: CapacitorConfig = {
  appId: "com.wwtd.tomatoinvestor",
  appName: "番茄投资人",
  webDir: "dist",
  plugins: {
    // 启用 CapacitorHttp：在原生端 patch window.fetch / XMLHttpRequest，
    // 让 API 请求走原生网络栈，绕开 WebView 对局域网 http 明文 + 跨源(CORS) 的限制。
    CapacitorHttp: {
      enabled: true,
    },
  },
  // APK 内页面通过本地 assets 加载；API 地址改为运行时配置，
  // 见 src/api.ts 的 getBase()。此处不设 server.url，避免写死。
};

export default config;
