import Link from "next/link";
import {
  Activity,
  Archive,
  ArrowUpRight,
  CheckCircle2,
  Clock3,
  Send,
  Split,
  Zap,
} from "lucide-react";

import { Shell } from "@/components/layout/shell";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { getChannelTargets, getContents, getPublishJobs, getSystemInfo } from "@/lib/api";

const cards = [
  { label: "Contents", key: "contents", icon: Archive, hint: "Ready to publish" },
  { label: "Targets", key: "targets", icon: Split, hint: "Distribution endpoints" },
  { label: "Jobs", key: "jobs", icon: Send, hint: "Execution history" },
] as const;

export default async function HomePage() {
  const [systemInfo, jobs, contents, targets] = await Promise.all([
    getSystemInfo(),
    getPublishJobs(),
    getContents(),
    getChannelTargets(),
  ]);

  const latestJob = jobs.items[0];
  const enabledTargets = targets.items.filter((item) => item.enabled).length;
  const successfulJobs = jobs.items.filter((item) => item.status === "success").length;
  const queuedJobs = jobs.items.filter((item) =>
    ["pending", "running", "processing"].includes(item.status),
  ).length;
  const failedJobs = jobs.items.filter((item) => item.failedCount > 0).length;
  const healthTone = systemInfo.status === "ok" ? "text-primary" : "text-[#ff8a8a]";

  return (
    <Shell>
      <section className="grid gap-6 lg:grid-cols-[1.35fr_0.95fr]">
        <Card className="relative overflow-hidden border-primary/15 bg-[linear-gradient(145deg,rgba(10,24,41,0.96),rgba(7,18,31,0.88))]">
          <div className="pointer-events-none absolute inset-y-0 right-0 w-1/2 bg-[radial-gradient(circle_at_center,rgba(57,217,138,0.12),transparent_58%)]" />
          <div className="relative space-y-8">
            <div className="flex flex-wrap items-center gap-3">
              <Badge className="border-primary/30 bg-primary/10 text-primary">Live control surface</Badge>
              <div className={`flex items-center gap-2 text-xs uppercase tracking-[0.24em] ${healthTone}`}>
                <Activity className="size-3.5" />
                {systemInfo.status}
              </div>
            </div>
            <div className="space-y-4">
              <h2 className="max-w-4xl text-4xl font-semibold leading-tight text-foreground sm:text-5xl">
                Route Markdown into a reliable multi-channel publishing pipeline.
              </h2>
              <p className="max-w-2xl text-base leading-8 text-foreground/70 sm:text-lg">
                This console centralizes content intake, target routing, delivery execution,
                and post-send inspection so publishing stays operational instead of manual.
              </p>
            </div>
            <div className="grid gap-3 sm:grid-cols-3">
              <div className="rounded-[24px] border border-border bg-white/[0.03] p-4">
                <p className="font-mono text-[11px] uppercase tracking-[0.24em] text-foreground/45">
                  Database
                </p>
                <p className="mt-3 text-lg font-semibold">{systemInfo.database}</p>
              </div>
              <div className="rounded-[24px] border border-border bg-white/[0.03] p-4">
                <p className="font-mono text-[11px] uppercase tracking-[0.24em] text-foreground/45">
                  Enabled targets
                </p>
                <p className="mt-3 text-lg font-semibold">{enabledTargets}</p>
              </div>
              <div className="rounded-[24px] border border-border bg-white/[0.03] p-4">
                <p className="font-mono text-[11px] uppercase tracking-[0.24em] text-foreground/45">
                  Latest job
                </p>
                <p className="mt-3 truncate text-lg font-semibold">
                  {latestJob ? latestJob.status : "No jobs"}
                </p>
              </div>
            </div>
            <div className="flex flex-wrap gap-3">
              <Link
                href="/publish/new"
                className="inline-flex min-h-11 items-center gap-2 rounded-2xl border border-primary/30 bg-primary px-5 py-3 text-sm font-semibold text-slate-950 transition hover:-translate-y-0.5 hover:bg-primary/90"
              >
                Create publish job
                <ArrowUpRight className="size-4" />
              </Link>
              <Link
                href="/contents"
                className="inline-flex min-h-11 items-center gap-2 rounded-2xl border border-border bg-white/[0.03] px-5 py-3 text-sm font-semibold text-foreground/75 transition hover:border-primary/30 hover:text-foreground"
              >
                Upload content
                <Archive className="size-4" />
              </Link>
            </div>
          </div>
        </Card>

        <Card className="bg-[linear-gradient(180deg,rgba(8,21,37,0.98),rgba(6,16,28,0.9))]">
          <div className="flex items-center justify-between gap-3">
            <div>
              <p className="font-mono text-[11px] uppercase tracking-[0.3em] text-accent/80">
                Pipeline snapshot
              </p>
              <h3 className="mt-2 text-2xl font-semibold">Operational status</h3>
            </div>
            <Zap className="size-5 text-accent" />
          </div>
          <div className="mt-6 grid gap-3">
            <div className="flex items-center justify-between rounded-2xl border border-border bg-white/[0.03] px-4 py-4">
              <div className="flex items-center gap-3">
                <CheckCircle2 className="size-4 text-primary" />
                <span className="text-sm text-foreground/68">Successful jobs</span>
              </div>
              <span className="font-mono text-lg text-foreground">{successfulJobs}</span>
            </div>
            <div className="flex items-center justify-between rounded-2xl border border-border bg-white/[0.03] px-4 py-4">
              <div className="flex items-center gap-3">
                <Clock3 className="size-4 text-accent" />
                <span className="text-sm text-foreground/68">Queued or running</span>
              </div>
              <span className="font-mono text-lg text-foreground">{queuedJobs}</span>
            </div>
            <div className="flex items-center justify-between rounded-2xl border border-border bg-white/[0.03] px-4 py-4">
              <div className="flex items-center gap-3">
                <Activity className="size-4 text-[#ff8a8a]" />
                <span className="text-sm text-foreground/68">Jobs with failures</span>
              </div>
              <span className="font-mono text-lg text-foreground">{failedJobs}</span>
            </div>
          </div>
          <div className="mt-6 rounded-[24px] border border-border bg-background/35 p-4">
            <p className="font-mono text-[11px] uppercase tracking-[0.24em] text-foreground/45">
              Recommended next action
            </p>
            <p className="mt-3 text-sm leading-7 text-foreground/70">
              {contents.items.length === 0
                ? "Start by uploading a Markdown source so the pipeline has content to distribute."
                : enabledTargets === 0
                  ? "Configure at least one enabled channel target before triggering a publish job."
                  : "The workspace is ready. Create a publish job and inspect delivery details from history."}
            </p>
          </div>
        </Card>
      </section>

      <section className="grid gap-4 md:grid-cols-3">
        {cards.map(({ label, key, icon: Icon, hint }) => {
          const value =
            key === "contents"
              ? contents.items.length
              : key === "targets"
                ? targets.items.length
                : jobs.items.length;

          return (
            <Card key={key} className="flex items-center justify-between gap-4">
              <div className="space-y-2">
                <p className="font-mono text-[11px] uppercase tracking-[0.24em] text-foreground/45">
                  {label}
                </p>
                <p className="text-3xl font-semibold">{value}</p>
                <p className="text-sm text-foreground/60">{hint}</p>
              </div>
              <div className="rounded-2xl border border-primary/20 bg-primary/10 p-4">
                <Icon className="size-6 text-primary" />
              </div>
            </Card>
          );
        })}
      </section>

      <section className="grid gap-6 xl:grid-cols-[1.35fr_0.65fr]">
        <Card>
          <div className="flex items-center justify-between gap-4">
            <div>
              <p className="font-mono text-[11px] uppercase tracking-[0.24em] text-foreground/45">
                Recent jobs
              </p>
              <h3 className="mt-2 text-2xl font-semibold">Latest publish activity</h3>
            </div>
          </div>
          <div className="mt-6 grid gap-3">
            {jobs.items.slice(0, 5).map((job) => (
              <div key={job.id} className="rounded-[22px] border border-border bg-white/[0.03] px-4 py-4">
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <div className="space-y-1">
                    <p className="font-medium text-foreground">{job.id}</p>
                    <p className="text-sm text-foreground/58">content: {job.contentId}</p>
                  </div>
                  <div className="flex flex-wrap items-center gap-2">
                    <Badge className="border-accent/20 bg-accent/10 text-accent">{job.status}</Badge>
                    <span className="font-mono text-xs text-foreground/45">
                      {new Date(job.createdAt).toLocaleString()}
                    </span>
                  </div>
                </div>
                <div className="mt-4 grid gap-2 sm:grid-cols-3">
                  <div className="rounded-2xl border border-border bg-background/35 px-3 py-2 text-sm text-foreground/65">
                    success: <span className="font-mono text-foreground">{job.successCount}</span>
                  </div>
                  <div className="rounded-2xl border border-border bg-background/35 px-3 py-2 text-sm text-foreground/65">
                    failed: <span className="font-mono text-foreground">{job.failedCount}</span>
                  </div>
                  <div className="rounded-2xl border border-border bg-background/35 px-3 py-2 text-sm text-foreground/65">
                    skipped: <span className="font-mono text-foreground">{job.skippedCount}</span>
                  </div>
                </div>
              </div>
            ))}
            {jobs.items.length === 0 ? (
              <p className="text-sm text-foreground/55">No publish jobs yet.</p>
            ) : null}
          </div>
        </Card>

        <Card className="bg-[linear-gradient(180deg,rgba(8,21,37,0.98),rgba(6,16,28,0.9))]">
          <p className="font-mono text-[11px] uppercase tracking-[0.24em] text-foreground/45">
            Operator checklist
          </p>
          <div className="mt-5 grid gap-3">
            {[
              {
                label: "Content intake",
                done: contents.items.length > 0,
                detail: "Upload and parse at least one Markdown source.",
              },
              {
                label: "Target routing",
                done: enabledTargets > 0,
                detail: "Keep one or more enabled targets ready for delivery.",
              },
              {
                label: "Execution trail",
                done: jobs.items.length > 0,
                detail: "Run a job to verify end-to-end publishing state.",
              },
            ].map((item) => (
              <div key={item.label} className="rounded-[22px] border border-border bg-white/[0.03] p-4">
                <div className="flex items-center justify-between gap-3">
                  <p className="text-sm font-semibold text-foreground">{item.label}</p>
                  <Badge
                    className={
                      item.done
                        ? "border-primary/30 bg-primary/10 text-primary"
                        : "border-border bg-muted text-foreground/68"
                    }
                  >
                    {item.done ? "ready" : "pending"}
                  </Badge>
                </div>
                <p className="mt-2 text-sm leading-7 text-foreground/62">{item.detail}</p>
              </div>
            ))}
          </div>
        </Card>
      </section>
    </Shell>
  );
}
