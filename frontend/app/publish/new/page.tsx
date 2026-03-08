import { CheckCheck, Send, Split, Text } from "lucide-react";

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

  const enabledAccounts = accounts.items.filter((account) => account.enabled);
  const enabledTargets = targets.items.filter((target) => target.enabled);

  return (
    <Shell>
      <section className="grid gap-6 xl:grid-cols-[0.82fr_1.18fr]">
        <Card className="bg-[linear-gradient(160deg,rgba(10,24,41,0.96),rgba(7,18,31,0.88))]">
          <Badge className="border-primary/30 bg-primary/10 text-primary">Dispatch</Badge>
          <h2 className="mt-4 text-3xl font-semibold sm:text-4xl">Create a publish job from normalized content.</h2>
          <p className="mt-4 max-w-xl text-sm leading-7 text-foreground/68">
            Publish jobs execute immediately in the background. Each selected target gets an
            independent delivery record so retries stay granular and auditable.
          </p>
          <div className="mt-6 grid gap-3">
            {[
              { icon: Text, label: "Available content", value: contents.items.length },
              { icon: CheckCheck, label: "Enabled accounts", value: enabledAccounts.length },
              { icon: Split, label: "Enabled targets", value: enabledTargets.length },
            ].map(({ icon: Icon, label, value }) => (
              <div key={label} className="flex items-center justify-between rounded-[22px] border border-border bg-background/30 px-4 py-4">
                <div className="flex items-center gap-3">
                  <Icon className="size-4 text-accent" />
                  <span className="text-sm text-foreground/68">{label}</span>
                </div>
                <span className="font-mono text-lg text-foreground">{value}</span>
              </div>
            ))}
          </div>
        </Card>

        <Card>
          <div className="flex items-center gap-3">
            <span className="inline-flex size-11 items-center justify-center rounded-2xl border border-primary/25 bg-primary/10 text-primary">
              <Send className="size-5" />
            </span>
            <div>
              <Badge>Publish</Badge>
              <h3 className="mt-2 text-3xl font-semibold">Create publish job</h3>
            </div>
          </div>
          <p className="mt-4 max-w-2xl text-sm leading-7 text-foreground/65">
            Duplicate body on the same target will be skipped automatically. Disabled accounts and
            targets are excluded from this form.
          </p>
          <div className="mt-6">
            <PublishJobForm
              contents={contents.items}
              accounts={enabledAccounts}
              targets={enabledTargets}
            />
          </div>
        </Card>
      </section>
    </Shell>
  );
}
