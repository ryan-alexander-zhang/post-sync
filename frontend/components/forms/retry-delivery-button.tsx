"use client";

import { useState, useTransition } from "react";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";

const API_BASE =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api/v1";

export function RetryDeliveryButton({ deliveryId }: { deliveryId: string }) {
  const router = useRouter();
  const [message, setMessage] = useState("");
  const [isPending, startTransition] = useTransition();

  return (
    <div className="flex items-center gap-3">
      <Button
        className="min-h-10 rounded-xl px-3 py-2 text-xs"
        disabled={isPending}
        onClick={() =>
          startTransition(async () => {
            const response = await fetch(`${API_BASE}/delivery-tasks/${deliveryId}/retry`, {
              method: "POST",
            });
            const payload = await response.json();
            if (!response.ok) {
              setMessage(payload.message ?? "Retry failed");
              return;
            }
            setMessage("Retry queued");
            router.refresh();
          })
        }
        type="button"
      >
        {isPending ? "Retrying..." : "Retry"}
      </Button>
      {message ? <span className="text-xs text-foreground/55">{message}</span> : null}
    </div>
  );
}
