import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // 実行に必要なファイルだけを .next/standalone に集約する。
  // コンテナイメージ(とその SBOM)に本番で使う依存だけが入るようにするため。
  output: "standalone",
  poweredByHeader: false,
};

export default nextConfig;
