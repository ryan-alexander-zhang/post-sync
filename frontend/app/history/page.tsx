import Link from "next/link";
import { CheckCircle2, Clock3, History, TriangleAlert } from "lucide-react";

import { Shell } from "@/components/layout/shell";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { Table, TD, TH } from "@/components/ui/table";
import { getPublishJobs } from "@/lib/api";

export default async function HistoryPage() {
  const jobs = await getPublishJobs();
  const success = jobs.items.filter((job) => job.status === "success").length;
  const active = jobs.items.filter((job) => ["pending", "running", "processing"].includes(job.status)).length;
  const risky = jobs.items.filter((job) => job.failedCount > 0).length;

  return (
    <Shell>
      <section className="grid gap-6 xl:grid-cols-[0.82fr_1.18fr]">
        <Card className="bg-[linear-gradient(160deg,rgba(10,24,41,0.96),rgba(7,18,31,0.88))]">
          <Badge className="border-primary/30 bg-primary/10 text-primary">Execution trail</Badge>
          <h2 className="mt-4 text-3xl font-semibold sm:text-4xl">Inspect publish history and delivery outcomes.</h2>
          <p className="mt-4 max-w-xl text-sm leading-7 text-foreground/68">
            Every publish job records a status fan-out so operators can audit routing, identify
            failures, and retry specific deliveries without losing the overall history.
          </p>
          <div className="mt-6 grid gap-3">
            {[
              { icon: History, label: "Total jobs", value: jobs.items.length, tone: "text-foreground" },
              { icon: CheckCircle2, label: "Successful", value: success, tone: "text-primary" },
              { icon: Clock3, label: "Active", value: active, tone: "text-accent" },
              { icon: TriangleAlert, label: "With failures", value: risky, tone: "text-[#ff8a8a]" },
            ].map(({ icon: Icon, label, value, tone }) => (
              <div key={label} className="flex items-center justify-between rounded-[22px] border border-border bg-background/30 px-4 py-4">
                <div className="flex items-center gap-3">
                  <Icon className={`size-4 ${tone}`} />
                  <span className="text-sm text-foreground/68">{label}</span>
                </div>
                <span className="font-mono text-lg text-foreground">{value}</span>
              </div>
            ))}
          </div>
        </Card>

        <Card>
          <div className="flex flex-wrap items-center justify-between gap-4">
            <div>
              <Badge>History</Badge>
              <h3 className="mt-3 text-3xl font-semibold">Publish jobs</h3>
            </div>
            <p className="max-w-md text-sm leading-7 text-foreground/62">
              Open a job to inspect per-target delivery records and retry only the failed items.
            </p>
          </div>
          <div className="mt-6 overflow-x-auto">
            <Table>
              <thead>
                <tr>
                  <TH>Job</TH>
                  <TH>Status</TH>
                  <TH>Success</TH>
                  <TH>Failed</TH>
                  <TH>Skipped</TH>
                  <TH>Created</TH>
                </tr>
              </thead>
              <tbody>
                {jobs.items.map((job) => (
                  <tr key={job.id} className="group">
                    <TD>
                      <Link className="font-medium text-primary transition hover:text-primary/85" href={`/history/${job.id}`}>
                        {job.id}
                      </Link>
                    </TD>
                    <TD>
                      <Badge
                        className={
                          job.status === "success"
                            ? "border-primary/30 bg-primary/10 text-primary"
                            : job.failedCount > 0
                              ? "border-[#ff8a8a]/30 bg-[#ff8a8a]/10 text-[#ff8a8a]"
                              : "border-accent/20 bg-accent/10 text-accent"
                        }
                      >
                        {job.status}
                      </Badge>
                    </TD>
                    <TD className="font-mono">{job.successCount}</TD>
                    <TD className="font-mono">{job.failedCount}</TD>
                    <TD className="font-mono">{job.skippedCount}</TD>
                    <TD className="text-foreground/68">{new Date(job.createdAt).toLocaleString()}</TD>
                  </tr>
                ))}
              </tbody>
            </Table>
          </div>
        </Card>
      </section>
    </Shell>
  );
}
