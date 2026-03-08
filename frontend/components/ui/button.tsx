import { ButtonHTMLAttributes } from "react";

import { cn } from "@/lib/utils";

const baseClassName =
  "inline-flex items-center justify-center rounded-2xl border px-4 py-2 text-sm font-medium transition hover:-translate-y-0.5";

export function Button({
  className,
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <button
      className={cn(
        baseClassName,
        "border-primary bg-primary text-white hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-60",
        className,
      )}
      {...props}
    />
  );
}
