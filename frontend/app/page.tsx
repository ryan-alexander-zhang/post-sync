import { Activity, Archive, Send, Split } from "lucide-react";

import { Shell } from "@/components/layout/shell";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { getChannelTargets, getContents, getPublishJobs, getSystemInfo } from "@/lib/api";

const cards = [
  { label: "Contents", key: "contents", icon: Archive },
  { label: "Targets", key: "targets", icon: Split },
  { label: "Jobs", key: "jobs", icon: Send },
];

export default async function HomePage() {
  const [systemInfo, jobs, contents, targets] = await Promise.all([
    getSystemInfo(),
    getPublishJobs(),
    getContents(),
    getChannelTargets(),
  ]);

  return (
    <Shell>
      <section className="grid gap-6 rounded-[32px] border border-border bg-card/90 p-8 shadow-[0_20px_80px_rgba(18,91,80,0.12)] md:grid-cols-[1.4fr_0.8fr]">
        <div className="space-y-5">
          <Badge>Live MVP</Badge>
          <h2 className="max-w-3xl text-5xl font-semibold leading-tight text-foreground">
            Publish one Markdown source to multiple Telegram targets with dedup and full delivery history.
          </h2>
          <p className="max-w-2xl text-lg leading-8 text-foreground/75">
            The backend runs async publish jobs, stores delivery status per target, and keeps channel credentials outside the database.
          </p>
        </div>
        <div className="rounded-[28px] bg-primary p-6 text-white">
          <div className="flex items-center gap-3">
            <Activity className="size-5" />
            <p className="text-sm uppercase tracking-[0.25em] text-white/80">
              System status
            </p>
          </div>
          <div className="mt-5 space-y-3 text-sm">
            <div className="rounded-2xl border border-white/10 bg-white/5 px-4 py-3">
              API: {systemInfo.status}
            </div>
            <div className="rounded-2xl border border-white/10 bg-white/5 px-4 py-3">
              DB driver: {systemInfo.database}
            </div>
          </div>
        </div>
      </section>

      <section className="grid gap-4 md:grid-cols-3">
        {cards.map(({ label, key, icon: Icon }) => {
          const value =
            key === "contents"
              ? contents.items.length
              : key === "targets"
                ? targets.items.length
                : jobs.items.length;

          return (
            <Card key={key} className="flex items-center justify-between">
              <div>
                <p className="text-sm text-foreground/60">{label}</p>
                <p className="mt-2 text-3xl font-semibold">{value}</p>
              </div>
              <div className="rounded-2xl bg-muted p-4">
                <Icon className="size-6 text-primary" />
              </div>
            </Card>
          );
        })}
      </section>

      <Card>
        <div className="flex items-center justify-between gap-4">
          <div>
            <p className="text-sm uppercase tracking-[0.2em] text-foreground/55">
              Recent jobs
            </p>
            <h3 className="mt-2 text-2xl font-semibold">Latest publish activity</h3>
          </div>
        </div>
        <div className="mt-6 grid gap-3">
          {jobs.items.slice(0, 5).map((job) => (
            <div key={job.id} className="rounded-2xl border border-border bg-white/70 px-4 py-4">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <p className="font-medium">{job.id}</p>
                  <p className="text-sm text-foreground/60">content: {job.contentId}</p>
                </div>
                <Badge>{job.status}</Badge>
              </div>
            </div>
          ))}
          {jobs.items.length === 0 ? (
            <p className="text-sm text-foreground/55">No publish jobs yet.</p>
          ) : null}
        </div>
      </Card>
    </Shell>
  );
}
