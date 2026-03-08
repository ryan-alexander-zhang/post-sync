import { HTMLAttributes } from "react";

import { cn } from "@/lib/utils";

export function Card({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn(
        "rounded-[28px] border border-border bg-card p-6 shadow-[0_24px_80px_rgba(0,0,0,0.2)] backdrop-blur-xl",
        className,
      )}
      {...props}
    />
  );
}
