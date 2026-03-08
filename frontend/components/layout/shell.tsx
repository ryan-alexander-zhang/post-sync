import Link from "next/link";
import { PropsWithChildren } from "react";

import { Badge } from "@/components/ui/badge";

const navItems = [
  { href: "/", label: "Overview" },
  { href: "/contents", label: "Contents" },
  { href: "/channels", label: "Channels" },
  { href: "/publish/new", label: "Publish" },
  { href: "/history", label: "History" },
];

export function Shell({ children }: PropsWithChildren) {
  return (
    <div className="min-h-screen">
      <header className="mx-auto flex max-w-6xl flex-col gap-4 px-6 py-8">
        <div className="flex flex-wrap items-center justify-between gap-3 rounded-[28px] border border-border bg-card/90 px-6 py-4 shadow-[0_20px_80px_rgba(18,91,80,0.08)]">
          <div className="space-y-1">
            <p className="text-xs uppercase tracking-[0.3em] text-primary">
              post-sync
            </p>
            <h1 className="text-xl font-semibold">Markdown distribution console</h1>
          </div>
          <Badge>MVP</Badge>
        </div>
        <nav className="flex flex-wrap gap-2">
          {navItems.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              className="rounded-full border border-border bg-white/70 px-4 py-2 text-sm font-medium text-foreground/75 transition hover:border-primary hover:text-primary"
            >
              {item.label}
            </Link>
          ))}
        </nav>
      </header>
      <div className="mx-auto flex max-w-6xl flex-col gap-6 px-6 pb-12">
        {children}
      </div>
    </div>
  );
}
