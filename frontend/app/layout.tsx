import type { Metadata } from "next";

import "./globals.css";

export const metadata: Metadata = {
  title: "post-sync console",
  description: "Markdown content distribution control center",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="zh-CN">
      <body>{children}</body>
    </html>
  );
}
