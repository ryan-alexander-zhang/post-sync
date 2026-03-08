"use client";

import { useRef, useState, useTransition } from "react";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { ChannelAccount } from "@/lib/types";
import { buildChannelTargetPayload } from "@/lib/channels";

const API_BASE =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api/v1";

export function ChannelTargetForm({ accounts }: { accounts: ChannelAccount[] }) {
  const router = useRouter();
  const formRef = useRef<HTMLFormElement>(null);
  const [message, setMessage] = useState("");
  const [accountID, setAccountID] = useState("");
  const [isPending, startTransition] = useTransition();
  const selectedAccount = accounts.find((account) => account.id === accountID);

  function submit(formData: FormData) {
    startTransition(async () => {
      const response = await fetch(`${API_BASE}/channel-targets`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(buildChannelTargetPayload(selectedAccount, formData)),
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
      <Select
        defaultValue=""
        name="channelAccountId"
        onChange={(event) => setAccountID(event.target.value)}
        required
      >
        <option disabled value="">
          Select account
        </option>
        {accounts.map((account) => (
          <option key={account.id} value={account.id}>
            {account.name} ({account.channelType})
          </option>
        ))}
      </Select>
      <Input name="targetName" placeholder="Target name" required />
      <Input
        name="targetKey"
        placeholder={
          selectedAccount?.channelType === "feishu"
            ? "Feishu chat id"
            : "Telegram group chat id"
        }
        required
      />
      {selectedAccount?.channelType !== "feishu" ? (
        <>
          <Input name="topicName" placeholder="Topic name (optional)" />
          <Input name="topicId" placeholder="Topic id / message_thread_id (optional)" />
        </>
      ) : null}
      <div className="flex items-center justify-between gap-3">
        <p className="text-sm text-foreground/65">
          {message ||
            (selectedAccount?.channelType === "feishu"
              ? "Feishu targets currently support chat_id group delivery."
              : "Leave topic fields empty to publish to the group root, or fill topic id to target a specific topic.")}
        </p>
        <Button disabled={isPending || accounts.length === 0} type="submit">
          {isPending ? "Saving..." : "Add target"}
        </Button>
      </div>
    </form>
  );
}
