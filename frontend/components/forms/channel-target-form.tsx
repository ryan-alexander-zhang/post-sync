"use client";

import { useRef, useState, useTransition } from "react";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { ChannelAccount } from "@/lib/types";

const API_BASE =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api/v1";

export function ChannelTargetForm({ accounts }: { accounts: ChannelAccount[] }) {
  const router = useRouter();
  const formRef = useRef<HTMLFormElement>(null);
  const [message, setMessage] = useState("");
  const [isPending, startTransition] = useTransition();

  function submit(formData: FormData) {
    startTransition(async () => {
      const response = await fetch(`${API_BASE}/channel-targets`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          channelAccountId: formData.get("channelAccountId"),
          targetType: "telegram_group",
          targetKey: formData.get("targetKey"),
          targetName: formData.get("targetName"),
          config: {
            disableNotification: false,
            disableWebPagePreview: false,
          },
        }),
      });

      const payload = await response.json();
      if (!response.ok) {
        setMessage(payload.message ?? "Create target failed");
        return;
      }

      setMessage(`Created target ${payload.targetName}`);
      formRef.current?.reset();
      router.refresh();
    });
  }

  return (
    <form ref={formRef} action={submit} className="grid gap-3">
      <Select name="channelAccountId" required defaultValue="">
        <option disabled value="">
          Select account
        </option>
        {accounts.map((account) => (
          <option key={account.id} value={account.id}>
            {account.name}
          </option>
        ))}
      </Select>
      <Input name="targetName" placeholder="Target name" required />
      <Input name="targetKey" placeholder="Telegram chat id" required />
      <div className="flex items-center justify-between gap-3">
        <p className="text-sm text-foreground/65">{message || "Use Telegram group chat_id as target key."}</p>
        <Button disabled={isPending || accounts.length === 0} type="submit">
          {isPending ? "Saving..." : "Add target"}
        </Button>
      </div>
    </form>
  );
}
