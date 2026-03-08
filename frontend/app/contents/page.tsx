import Link from "next/link";

import { DeleteButton } from "@/components/forms/delete-button";
import { Shell } from "@/components/layout/shell";
import { ContentUploadForm } from "@/components/forms/content-upload-form";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { Table, TD, TH } from "@/components/ui/table";
import { getContents } from "@/lib/api";

export default async function ContentsPage() {
  const data = await getContents();

  return (
    <Shell>
      <Card>
        <div className="flex items-center justify-between gap-4">
          <div>
            <Badge>Content</Badge>
            <h2 className="mt-3 text-3xl font-semibold">Upload Markdown files</h2>
          </div>
        </div>
        <div className="mt-6">
          <ContentUploadForm />
        </div>
      </Card>

      <Card>
        <h3 className="text-xl font-semibold">Stored contents</h3>
        <div className="mt-5 overflow-x-auto">
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
                <tr key={item.id}>
                  <TD>
                    <Link className="font-medium text-primary" href={`/contents/${item.id}`}>
                      {item.title || item.sourceFilename}
                    </Link>
                  </TD>
                  <TD>{item.sourceFilename}</TD>
                  <TD className="font-mono text-xs">{item.bodyHash}</TD>
                  <TD>{new Date(item.createdAt).toLocaleString()}</TD>
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
