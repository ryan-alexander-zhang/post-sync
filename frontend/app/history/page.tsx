import Link from "next/link";

import { Shell } from "@/components/layout/shell";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { Table, TD, TH } from "@/components/ui/table";
import { getPublishJobs } from "@/lib/api";

export default async function HistoryPage() {
  const jobs = await getPublishJobs();

  return (
    <Shell>
      <Card>
        <div className="flex items-center justify-between gap-4">
          <div>
            <Badge>History</Badge>
            <h2 className="mt-3 text-3xl font-semibold">Publish jobs</h2>
          </div>
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
                <tr key={job.id}>
                  <TD>
                    <Link className="font-medium text-primary" href={`/history/${job.id}`}>
                      {job.id}
                    </Link>
                  </TD>
                  <TD>{job.status}</TD>
                  <TD>{job.successCount}</TD>
                  <TD>{job.failedCount}</TD>
                  <TD>{job.skippedCount}</TD>
                  <TD>{new Date(job.createdAt).toLocaleString()}</TD>
                </tr>
              ))}
            </tbody>
          </Table>
        </div>
      </Card>
    </Shell>
  );
}
