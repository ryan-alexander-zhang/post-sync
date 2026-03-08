"use client";

import { useRef, useState, useTransition } from "react";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import {
  buildChannelAccountPayload,
  CHANNEL_OPTIONS,
  SupportedChannelType,
} from "@/lib/channels";

const API_BASE =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api/v1";

export function ChannelAccountForm() {
  const router = useRouter();
  const formRef = useRef<HTMLFormElement>(null);
  const [message, setMessage] = useState("");
  const [channelType, setChannelType] = useState<SupportedChannelType>("telegram");
  const [isPending, startTransition] = useTransition();

  function submit(formData: FormData) {
    startTransition(async () => {
      const response = await fetch(`${API_BASE}/channel-accounts`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(buildChannelAccountPayload(channelType, formData)),
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
      <Select
        name="channelType"
        onChange={(event) => setChannelType(event.target.value as SupportedChannelType)}
        value={channelType}
      >
        {CHANNEL_OPTIONS.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </Select>
      <Input name="name" placeholder="Account name" required />
      {channelType !== "personal_feishu" ? (
        <Input
          autoCapitalize="off"
          autoCorrect="off"
          defaultValue={
            CHANNEL_OPTIONS.find((option) => option.value === channelType)?.defaultSecretRef
          }
          key={channelType}
          name="secretRef"
          placeholder="Secret env ref"
          required
        />
      ) : null}
      {channelType === "feishu" ? (
        <>
          <Input
            autoCapitalize="off"
            autoCorrect="off"
            defaultValue="FEISHU_APP_ID"
            name="appIdEnv"
            placeholder="FEISHU_APP_ID"
            required
          />
          <Input
            autoCapitalize="off"
            autoCorrect="off"
            defaultValue="FEISHU_TENANT_ACCESS_TOKEN"
            name="tokenEnv"
            placeholder="FEISHU_TENANT_ACCESS_TOKEN (optional)"
          />
          <Input
            autoCapitalize="off"
            autoCorrect="off"
            name="baseUrl"
            placeholder="https://open.feishu.cn (optional)"
          />
        </>
      ) : null}
      {channelType === "personal_feishu" ? (
        <Input
          autoCapitalize="off"
          autoCorrect="off"
          name="webhookUrl"
          placeholder="https://open.feishu.cn/open-apis/bot/v2/hook/..."
          required
        />
      ) : null}
      <div className="flex items-center justify-between gap-3">
        <p className="text-sm text-foreground/65">
          {message ||
            (channelType === "feishu"
              ? "Enterprise Feishu uses FEISHU_APP_ID + FEISHU_APP_SECRET by default, with optional FEISHU_TENANT_ACCESS_TOKEN override."
              : channelType === "personal_feishu"
                ? "Personal Feishu only needs a webhook URL from the custom bot."
                : "Telegram accounts should usually use TELEGRAM_BOT_TOKEN as secretRef.")}
        </p>
        <Button disabled={isPending} type="submit">
          {isPending ? "Saving..." : "Add account"}
        </Button>
      </div>
    </form>
  );
}
