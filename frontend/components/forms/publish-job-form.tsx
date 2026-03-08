"use client";

import { useRef, useState, useTransition } from "react";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";
import { Select } from "@/components/ui/select";
import { ChannelAccount, ChannelTarget, Content } from "@/lib/types";

const API_BASE =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api/v1";

export function PublishJobForm({
  contents,
  accounts,
  targets,
}: {
  contents: Content[];
  accounts: ChannelAccount[];
  targets: ChannelTarget[];
}) {
  const router = useRouter();
  const formRef = useRef<HTMLFormElement>(null);
  const [message, setMessage] = useState("");
  const [isPending, startTransition] = useTransition();
  const accountMap = new Map(accounts.map((account) => [account.id, account]));

  function submit(formData: FormData) {
    const selectedTargets = formData.getAll("targetIds");
    startTransition(async () => {
      const response = await fetch(`${API_BASE}/publish-jobs`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          contentId: formData.get("contentId"),
          targetIds: selectedTargets,
          templateName: "default",
        }),
      });

      const payload = await response.json();
      if (!response.ok) {
        setMessage(payload.message ?? "Create publish job failed");
        return;
      }

      setMessage(`Job ${payload.jobId} created`);
      formRef.current?.reset();
      router.push(`/history/${payload.jobId}`);
      router.refresh();
    });
  }

  return (
    <form ref={formRef} action={submit} className="grid gap-5">
      <Select name="contentId" required defaultValue="">
        <option disabled value="">
          Select content
        </option>
        {contents.map((content) => (
          <option key={content.id} value={content.id}>
            {content.title || content.sourceFilename}
          </option>
        ))}
      </Select>
      <div className="grid gap-3 rounded-[24px] border border-border bg-background/30 p-4">
        <p className="text-sm font-medium">Select targets</p>
        {targets.map((target) => {
          const account = accountMap.get(target.channelAccountId);
          const channelLabel = account
            ? `${account.name} · ${account.channelType}`
            : target.channelAccountId;

          return (
            <label
              key={target.id}
              className="grid cursor-pointer grid-cols-[auto_1fr] items-start gap-3 rounded-2xl border border-transparent px-3 py-3 text-sm transition hover:border-primary/20 hover:bg-white/[0.02]"
            >
              <input className="mt-0.5 size-4 accent-[#39d98a]" name="targetIds" type="checkbox" value={target.id} />
              <span className="grid gap-1">
                <span className="font-medium">{target.targetName}</span>
                <span className="text-xs text-foreground/55">{channelLabel}</span>
                <span className="text-xs text-foreground/45">{target.targetKey}</span>
              </span>
            </label>
          );
        })}
      </div>
      <div className="flex items-center justify-between gap-3">
        <p className="text-sm text-foreground/65">{message || "Duplicate body on the same target will be skipped automatically."}</p>
        <Button disabled={isPending || contents.length === 0 || targets.length === 0} type="submit">
          {isPending ? "Submitting..." : "Create job"}
        </Button>
      </div>
    </form>
  );
}
