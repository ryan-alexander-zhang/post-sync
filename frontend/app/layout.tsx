import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "post-sync",
  description: "Markdown content distributor MVP",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
