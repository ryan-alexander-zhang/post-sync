import { HTMLAttributes } from "react";

import { cn } from "@/lib/utils";

export function Card({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn(
        "rounded-[28px] border border-border bg-card/90 p-6 shadow-[0_20px_80px_rgba(18,91,80,0.08)]",
        className,
      )}
      {...props}
    />
  );
}
