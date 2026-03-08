import { Shell } from "@/components/layout/shell";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { getContent } from "@/lib/api";

export default async function ContentDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const content = await getContent(id);

  return (
    <Shell>
      <Card>
        <Badge>Content detail</Badge>
        <h2 className="mt-3 text-3xl font-semibold">{content.title || content.sourceFilename}</h2>
        <div className="mt-6 grid gap-4 md:grid-cols-2">
          <div className="rounded-[24px] border border-border bg-background/30 p-5">
            <p className="font-mono text-[11px] uppercase tracking-[0.24em] text-foreground/45">Metadata</p>
            <pre className="mt-4 overflow-auto whitespace-pre-wrap text-sm leading-7 text-foreground/75">
              {content.frontmatterJson}
            </pre>
          </div>
          <div className="rounded-[24px] border border-border bg-background/30 p-5">
            <p className="font-mono text-[11px] uppercase tracking-[0.24em] text-foreground/45">Normalized body</p>
            <pre className="mt-4 overflow-auto whitespace-pre-wrap text-sm leading-7 text-foreground/75">
              {content.bodyMarkdown}
            </pre>
          </div>
        </div>
        <div className="mt-6 rounded-[24px] border border-border bg-muted/60 p-5">
          <p className="font-mono text-[11px] uppercase tracking-[0.24em] text-foreground/45">Body hash</p>
          <p className="mt-2 font-mono text-sm">{content.bodyHash}</p>
        </div>
      </Card>
    </Shell>
  );
}
