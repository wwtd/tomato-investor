import type { CapacitorConfig } from "@capacitor/cli";

const config: CapacitorConfig = {
  appId: "com.wwtd.tomatoinvestor",
  appName: "番茄投资人",
  webDir: "dist",
  // APK 内页面通过本地 dev server 加载；API 地址改为运行时配置，
  // 见 src/api.ts 的 getBase()。此处不设 server.url，避免写死。
};

export default config;
