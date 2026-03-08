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
          <div className="rounded-[24px] border border-border bg-white/65 p-5">
            <p className="text-sm uppercase tracking-[0.2em] text-foreground/55">Metadata</p>
            <pre className="mt-4 overflow-auto whitespace-pre-wrap text-sm leading-7 text-foreground/75">
              {content.frontmatterJson}
            </pre>
          </div>
          <div className="rounded-[24px] border border-border bg-white/65 p-5">
            <p className="text-sm uppercase tracking-[0.2em] text-foreground/55">Normalized body</p>
            <pre className="mt-4 overflow-auto whitespace-pre-wrap text-sm leading-7 text-foreground/75">
              {content.bodyMarkdown}
            </pre>
          </div>
        </div>
        <div className="mt-6 rounded-[24px] border border-border bg-muted/60 p-5">
          <p className="text-sm font-medium">Body hash</p>
          <p className="mt-2 font-mono text-sm">{content.bodyHash}</p>
        </div>
      </Card>
    </Shell>
  );
}
