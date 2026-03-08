import Link from "next/link";
import { Archive, FileText, Hash, Upload } from "lucide-react";

import { DeleteButton } from "@/components/forms/delete-button";
import { Shell } from "@/components/layout/shell";
import { ContentUploadForm } from "@/components/forms/content-upload-form";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { Table, TD, TH } from "@/components/ui/table";
import { getContents } from "@/lib/api";

export default async function ContentsPage() {
  const data = await getContents();
  const latest = data.items[0];

  return (
    <Shell>
      <section className="grid gap-6 xl:grid-cols-[0.95fr_1.05fr]">
        <Card className="bg-[linear-gradient(160deg,rgba(10,24,41,0.96),rgba(7,18,31,0.88))]">
          <Badge className="border-primary/30 bg-primary/10 text-primary">Content intake</Badge>
          <h2 className="mt-4 text-3xl font-semibold sm:text-4xl">Upload and normalize Markdown sources.</h2>
          <p className="mt-4 max-w-xl text-sm leading-7 text-foreground/68">
            Each file is parsed into a stable content model so downstream publish jobs can reuse
            the same source with dedup-aware delivery.
          </p>
          <div className="mt-6 grid gap-3 sm:grid-cols-3">
            <div className="rounded-[22px] border border-border bg-background/30 p-4">
              <p className="font-mono text-[11px] uppercase tracking-[0.24em] text-foreground/45">Stored</p>
              <p className="mt-3 text-2xl font-semibold">{data.items.length}</p>
            </div>
            <div className="rounded-[22px] border border-border bg-background/30 p-4">
              <p className="font-mono text-[11px] uppercase tracking-[0.24em] text-foreground/45">Latest source</p>
              <p className="mt-3 truncate text-sm font-semibold text-foreground/78">
                {latest?.sourceFilename ?? "No content yet"}
              </p>
            </div>
            <div className="rounded-[22px] border border-border bg-background/30 p-4">
              <p className="font-mono text-[11px] uppercase tracking-[0.24em] text-foreground/45">Next step</p>
              <p className="mt-3 text-sm font-semibold text-foreground/78">
                {data.items.length === 0 ? "Upload first file" : "Review then publish"}
              </p>
            </div>
          </div>
        </Card>

        <Card>
          <div className="flex items-center gap-3">
            <span className="inline-flex size-11 items-center justify-center rounded-2xl border border-primary/25 bg-primary/10 text-primary">
              <Upload className="size-5" />
            </span>
            <div>
              <p className="font-mono text-[11px] uppercase tracking-[0.24em] text-foreground/45">Upload panel</p>
              <h3 className="mt-1 text-2xl font-semibold">Add Markdown file</h3>
            </div>
          </div>
          <div className="mt-6">
            <ContentUploadForm />
          </div>
          <div className="mt-6 grid gap-3 sm:grid-cols-3">
            {[
              { icon: FileText, label: "Frontmatter parsed" },
              { icon: Hash, label: "Body hash generated" },
              { icon: Archive, label: "Stored for reuse" },
            ].map(({ icon: Icon, label }) => (
              <div key={label} className="flex items-center gap-3 rounded-2xl border border-border bg-background/30 px-4 py-4">
                <Icon className="size-4 text-accent" />
                <span className="text-sm text-foreground/70">{label}</span>
              </div>
            ))}
          </div>
        </Card>
      </section>

      <Card>
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <p className="font-mono text-[11px] uppercase tracking-[0.24em] text-foreground/45">Repository</p>
            <h3 className="mt-2 text-2xl font-semibold">Stored contents</h3>
          </div>
          <Badge className="border-accent/20 bg-accent/10 text-accent">{data.items.length} items</Badge>
        </div>
        <div className="mt-6 overflow-x-auto">
          <Table>
            <thead>
              <tr>
                <TH>Title</TH>
                <TH>Source</TH>
                <TH>Body hash</TH>
                <TH>Created</TH>
                <TH>Action</TH>
              </tr>
            </thead>
            <tbody>
              {data.items.map((item) => (
                <tr key={item.id} className="group">
                  <TD>
                    <Link className="font-medium text-primary transition hover:text-primary/85" href={`/contents/${item.id}`}>
                      {item.title || item.sourceFilename}
                    </Link>
                  </TD>
                  <TD className="text-foreground/72">{item.sourceFilename}</TD>
                  <TD className="font-mono text-xs text-foreground/62">{item.bodyHash}</TD>
                  <TD className="text-foreground/68">{new Date(item.createdAt).toLocaleString()}</TD>
                  <TD>
                    <DeleteButton
                      label={`content ${item.title || item.sourceFilename}`}
                      path={`/contents/${item.id}`}
                    />
                  </TD>
                </tr>
              ))}
            </tbody>
          </Table>
          {data.items.length === 0 ? (
            <p className="pt-4 text-sm text-foreground/55">No contents uploaded yet.</p>
          ) : null}
        </div>
      </Card>
    </Shell>
  );
}
