"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { PropsWithChildren } from "react";
import { ArrowUpRight, Layers3, Sparkles } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

const navItems = [
  { href: "/", label: "Overview" },
  { href: "/contents", label: "Contents" },
  { href: "/channels", label: "Channels" },
  { href: "/publish/new", label: "Publish" },
  { href: "/history", label: "History" },
];

export function Shell({ children }: PropsWithChildren) {
  const pathname = usePathname();

  return (
    <div className="relative min-h-screen overflow-hidden">
      <div className="pointer-events-none absolute inset-x-0 top-0 h-64 bg-[radial-gradient(circle_at_center,rgba(57,217,138,0.14),transparent_60%)]" />
      <header className="mx-auto flex max-w-7xl flex-col gap-5 px-5 py-6 sm:px-6 lg:px-8">
        <div className="rounded-[32px] border border-border bg-card/80 px-5 py-5 shadow-[0_24px_90px_rgba(0,0,0,0.28)] backdrop-blur-xl">
          <div className="flex flex-col gap-5 lg:flex-row lg:items-center lg:justify-between">
            <div className="space-y-3">
              <div className="flex items-center gap-3">
                <span className="inline-flex size-10 items-center justify-center rounded-2xl border border-primary/35 bg-primary/10 text-primary">
                  <Layers3 className="size-5" />
                </span>
                <div>
                  <p className="font-mono text-[11px] uppercase tracking-[0.34em] text-primary/85">
                    post-sync control plane
                  </p>
                  <h1 className="text-xl font-semibold tracking-[0.02em] text-foreground sm:text-2xl">
                    Markdown distribution console
                  </h1>
                </div>
              </div>
              <p className="max-w-3xl text-sm leading-7 text-foreground/70 sm:text-[15px]">
                Upload one Markdown source, route it across Telegram and Feishu targets,
                track delivery state, and keep duplicate-safe history in one operational surface.
              </p>
            </div>
            <div className="flex flex-wrap items-center gap-3">
              <Badge className="border-primary/30 bg-primary/10 text-primary">Operational UI</Badge>
              <div className="inline-flex items-center gap-2 rounded-full border border-border bg-background/40 px-3 py-2 text-xs text-foreground/60">
                <Sparkles className="size-3.5 text-accent" />
                Next.js dashboard
              </div>
            </div>
          </div>
        </div>
        <nav className="flex flex-wrap gap-2">
          {navItems.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "inline-flex min-h-11 items-center gap-2 rounded-full border px-4 py-2 text-sm font-medium transition duration-200",
                pathname === item.href
                  ? "border-primary/40 bg-primary/10 text-primary shadow-[0_0_0_1px_rgba(57,217,138,0.1)]"
                  : "border-border bg-card/65 text-foreground/72 hover:border-primary/30 hover:bg-card hover:text-foreground",
              )}
            >
              <span>{item.label}</span>
              {pathname === item.href ? <ArrowUpRight className="size-4" /> : null}
            </Link>
          ))}
        </nav>
      </header>
      <div className="mx-auto flex max-w-7xl flex-col gap-6 px-5 pb-12 sm:px-6 lg:px-8">
        {children}
      </div>
    </div>
  );
}
