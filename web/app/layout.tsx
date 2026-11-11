import type { Metadata } from "next";
import Link from "next/link";
import type { ReactNode } from "react";
import "./globals.css";

export const metadata: Metadata = {
  title: "SBOM Training Blog",
  description: "SBOM による依存管理・脆弱性管理のサンプルアプリケーション",
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="ja">
      <body>
        <header className="site-header">
          <Link href="/" className="site-title">
            SBOM Training Blog
          </Link>
          <Link href="/posts/new" className="button">
            記事を書く
          </Link>
        </header>
        <main>{children}</main>
      </body>
    </html>
  );
}
