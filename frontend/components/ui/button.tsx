import { ButtonHTMLAttributes } from "react";

import { cn } from "@/lib/utils";

const baseClassName =
  "inline-flex min-h-11 items-center justify-center rounded-2xl border px-4 py-2 text-sm font-medium transition duration-200 disabled:cursor-not-allowed disabled:opacity-60";

export function Button({
  className,
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <button
      className={cn(
        baseClassName,
        "border-primary/30 bg-primary text-slate-950 shadow-[0_14px_30px_rgba(57,217,138,0.22)] hover:-translate-y-0.5 hover:bg-primary/90",
        className,
      )}
      {...props}
    />
  );
}
