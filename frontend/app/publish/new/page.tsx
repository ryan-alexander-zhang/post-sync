import { Shell } from "@/components/layout/shell";
import { PublishJobForm } from "@/components/forms/publish-job-form";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { getChannelAccounts, getChannelTargets, getContents } from "@/lib/api";

export default async function PublishPage() {
  const [contents, accounts, targets] = await Promise.all([
    getContents(),
    getChannelAccounts(),
    getChannelTargets(),
  ]);

  return (
    <Shell>
      <Card>
        <Badge>Publish</Badge>
        <h2 className="mt-3 text-3xl font-semibold">Create publish job</h2>
        <p className="mt-3 max-w-2xl text-sm leading-7 text-foreground/65">
          Jobs are created immediately and executed in the background. Each target gets its own delivery record with success, failure, or duplicate-skip state.
        </p>
        <div className="mt-6">
          <PublishJobForm
            contents={contents.items}
            accounts={accounts.items.filter((account) => account.enabled)}
            targets={targets.items.filter((target) => target.enabled)}
          />
        </div>
      </Card>
    </Shell>
  );
}
