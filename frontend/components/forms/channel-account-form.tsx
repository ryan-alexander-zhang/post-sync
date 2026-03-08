"use client";

import { useRef, useState, useTransition } from "react";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

const API_BASE =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api/v1";

export function ChannelAccountForm() {
  const router = useRouter();
  const formRef = useRef<HTMLFormElement>(null);
  const [message, setMessage] = useState("");
  const [isPending, startTransition] = useTransition();

  function submit(formData: FormData) {
    startTransition(async () => {
      const response = await fetch(`${API_BASE}/channel-accounts`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          channelType: "telegram",
          name: formData.get("name"),
          secretRef: formData.get("secretRef"),
          config: {},
        }),
      });

      const payload = await response.json();
      if (!response.ok) {
        setMessage(payload.message ?? "Create channel account failed");
        return;
      }

      setMessage(`Created account ${payload.name}`);
      formRef.current?.reset();
      router.refresh();
    });
  }

  return (
    <form ref={formRef} action={submit} className="grid gap-3">
      <Input name="name" placeholder="Account name" required />
      <Input
        autoCapitalize="off"
        autoCorrect="off"
        defaultValue="TELEGRAM_BOT_TOKEN"
        name="secretRef"
        placeholder="TELEGRAM_BOT_TOKEN"
        required
      />
      <div className="flex items-center justify-between gap-3">
        <p className="text-sm text-foreground/65">
          {message || "Telegram accounts should usually use TELEGRAM_BOT_TOKEN as secretRef."}
        </p>
        <Button disabled={isPending} type="submit">
          {isPending ? "Saving..." : "Add account"}
        </Button>
      </div>
    </form>
  );
}
