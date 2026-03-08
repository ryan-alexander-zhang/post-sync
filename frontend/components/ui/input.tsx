import { InputHTMLAttributes } from "react";

import { cn } from "@/lib/utils";

export function Input({
  className,
  ...props
}: InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      className={cn(
        "w-full rounded-2xl border border-border bg-background/45 px-4 py-3 text-sm text-foreground outline-none ring-0 transition placeholder:text-foreground/35 focus:border-primary/45 focus:bg-background/60",
        className,
      )}
      {...props}
    />
  );
}
