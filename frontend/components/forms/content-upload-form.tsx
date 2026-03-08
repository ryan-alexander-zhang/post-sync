"use client";

import { useRef, useState, useTransition } from "react";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

const API_BASE =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api/v1";

export function ContentUploadForm() {
  const router = useRouter();
  const formRef = useRef<HTMLFormElement>(null);
  const [message, setMessage] = useState<string>("");
  const [isPending, startTransition] = useTransition();

  function handleSubmit(formData: FormData) {
    startTransition(async () => {
      const response = await fetch(`${API_BASE}/contents/upload`, {
        method: "POST",
        body: formData,
      });

      const payload = await response.json();
      if (!response.ok) {
        setMessage(payload.message ?? "Upload failed");
        return;
      }

      setMessage(`Uploaded ${payload.sourceFilename}`);
      formRef.current?.reset();
      router.refresh();
    });
  }

  return (
    <form
      ref={formRef}
      action={handleSubmit}
      className="grid gap-4 rounded-[24px] border border-dashed border-primary/40 bg-white/60 p-5"
    >
      <Input name="file" type="file" accept=".md,.markdown,text/markdown,text/plain" />
      <div className="flex items-center justify-between gap-3">
        <p className="text-sm text-foreground/65">{message || "Upload one Markdown file at a time."}</p>
        <Button disabled={isPending} type="submit">
          {isPending ? "Uploading..." : "Upload"}
        </Button>
      </div>
    </form>
  );
}
