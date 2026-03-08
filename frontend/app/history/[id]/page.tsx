import { Shell } from "@/components/layout/shell";
import { RetryDeliveryButton } from "@/components/forms/retry-delivery-button";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { Table, TD, TH } from "@/components/ui/table";
import { getPublishJob } from "@/lib/api";

export default async function HistoryDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const { job, deliveries } = await getPublishJob(id);

  return (
    <Shell>
      <Card>
        <Badge>{job.status}</Badge>
        <h2 className="mt-3 text-3xl font-semibold">Publish job detail</h2>
        <div className="mt-6 grid gap-4 md:grid-cols-4">
          <div className="rounded-2xl border border-border bg-background/30 p-4">
            <p className="font-mono text-[11px] uppercase tracking-[0.24em] text-foreground/45">Total</p>
            <p className="mt-2 text-3xl font-semibold">{job.totalDeliveries}</p>
          </div>
          <div className="rounded-2xl border border-border bg-background/30 p-4">
            <p className="font-mono text-[11px] uppercase tracking-[0.24em] text-foreground/45">Success</p>
            <p className="mt-2 text-3xl font-semibold">{job.successCount}</p>
          </div>
          <div className="rounded-2xl border border-border bg-background/30 p-4">
            <p className="font-mono text-[11px] uppercase tracking-[0.24em] text-foreground/45">Failed</p>
            <p className="mt-2 text-3xl font-semibold">{job.failedCount}</p>
          </div>
          <div className="rounded-2xl border border-border bg-background/30 p-4">
            <p className="font-mono text-[11px] uppercase tracking-[0.24em] text-foreground/45">Skipped</p>
            <p className="mt-2 text-3xl font-semibold">{job.skippedCount}</p>
          </div>
        </div>
      </Card>

      <Card>
        <h3 className="text-xl font-semibold">Deliveries</h3>
        <div className="mt-5 overflow-x-auto">
          <Table>
            <thead>
              <tr>
                <TH>Target</TH>
                <TH>Status</TH>
                <TH>Attempts</TH>
                <TH>Error</TH>
                <TH>Action</TH>
              </tr>
            </thead>
            <tbody>
              {deliveries.map((delivery) => (
                <tr key={delivery.id}>
                  <TD>
                    <div className="font-medium">{delivery.targetKey}</div>
                    <div className="font-mono text-xs text-foreground/55">{delivery.id}</div>
                  </TD>
                  <TD>{delivery.status}</TD>
                  <TD>{delivery.attemptCount}</TD>
                  <TD className="max-w-sm whitespace-pre-wrap text-xs text-foreground/65">
                    {delivery.errorCode || delivery.errorMessage || "-"}
                  </TD>
                  <TD>
                    {delivery.status === "FAILED" ? (
                      <RetryDeliveryButton deliveryId={delivery.id} />
                    ) : (
                      "-"
                    )}
                  </TD>
                </tr>
              ))}
            </tbody>
          </Table>
        </div>
      </Card>
    </Shell>
  );
}
