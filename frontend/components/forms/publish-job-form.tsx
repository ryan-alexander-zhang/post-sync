"use client";

import { useRef, useState, useTransition } from "react";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";
import { Select } from "@/components/ui/select";
import { ChannelTarget, Content } from "@/lib/types";

const API_BASE =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api/v1";

export function PublishJobForm({
  contents,
  targets,
}: {
  contents: Content[];
  targets: ChannelTarget[];
}) {
  const router = useRouter();
  const formRef = useRef<HTMLFormElement>(null);
  const [message, setMessage] = useState("");
  const [isPending, startTransition] = useTransition();

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
      <div className="grid gap-3 rounded-[24px] border border-border bg-white/60 p-4">
        <p className="text-sm font-medium">Select targets</p>
        {targets.map((target) => (
          <label key={target.id} className="flex items-center gap-3 text-sm">
            <input className="size-4 accent-[#125b50]" name="targetIds" type="checkbox" value={target.id} />
            <span>{target.targetName}</span>
            <span className="text-foreground/45">{target.targetKey}</span>
          </label>
        ))}
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
