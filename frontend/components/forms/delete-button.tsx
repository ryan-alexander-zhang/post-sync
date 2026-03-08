"use client";

import { useState, useTransition } from "react";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";

const API_BASE =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api/v1";

export function DeleteButton({
  path,
  label,
}: {
  path: string;
  label: string;
}) {
  const router = useRouter();
  const [message, setMessage] = useState("");
  const [isPending, startTransition] = useTransition();

  return (
    <div className="flex items-center gap-2">
      <Button
        className="border-red-700 bg-red-700 hover:bg-red-700/90"
        disabled={isPending}
        onClick={() => {
          if (!window.confirm(`Delete ${label}? This cannot be undone.`)) {
            return;
          }

          startTransition(async () => {
            const response = await fetch(`${API_BASE}${path}`, {
              method: "DELETE",
            });

            if (!response.ok) {
              const payload = await response.json();
              setMessage(payload.message ?? `Delete ${label} failed`);
              return;
            }

            setMessage("");
            router.refresh();
          });
        }}
        type="button"
      >
        {isPending ? "Deleting..." : "Delete"}
      </Button>
      {message ? <span className="text-xs text-red-700">{message}</span> : null}
    </div>
  );
}
