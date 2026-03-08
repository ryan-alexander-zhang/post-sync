import { TextareaHTMLAttributes } from "react";

import { cn } from "@/lib/utils";

export function Textarea({
  className,
  ...props
}: TextareaHTMLAttributes<HTMLTextAreaElement>) {
  return (
    <textarea
      className={cn(
        "min-h-28 w-full rounded-2xl border border-border bg-background/45 px-4 py-3 text-sm text-foreground outline-none transition placeholder:text-foreground/35 focus:border-primary/45 focus:bg-background/60",
        className,
      )}
      {...props}
    />
  );
}
