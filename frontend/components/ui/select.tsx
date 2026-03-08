import { SelectHTMLAttributes } from "react";

import { cn } from "@/lib/utils";

export function Select({
  className,
  ...props
}: SelectHTMLAttributes<HTMLSelectElement>) {
  return (
    <select
      className={cn(
        "w-full rounded-2xl border border-border bg-background/45 px-4 py-3 text-sm text-foreground outline-none transition focus:border-primary/45 focus:bg-background/60",
        className,
      )}
      {...props}
    />
  );
}
