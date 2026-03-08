import { KeyRound, Network, RadioTower, Route } from "lucide-react";

import { Shell } from "@/components/layout/shell";
import { ChannelAccountForm } from "@/components/forms/channel-account-form";
import { DeleteButton } from "@/components/forms/delete-button";
import { ChannelTargetForm } from "@/components/forms/channel-target-form";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { Table, TD, TH } from "@/components/ui/table";
import { getChannelAccounts, getChannelTargets } from "@/lib/api";
import { describeTarget, parseFeishuAccountConfig } from "@/lib/channels";

export default async function ChannelsPage() {
  const [accounts, targets] = await Promise.all([
    getChannelAccounts(),
    getChannelTargets(),
  ]);

  const enabledAccounts = accounts.items.filter((account) => account.enabled).length;
  const enabledTargets = targets.items.filter((target) => target.enabled).length;

  return (
    <Shell>
      <section className="grid gap-6 xl:grid-cols-[0.78fr_1.22fr]">
        <Card className="bg-[linear-gradient(160deg,rgba(10,24,41,0.96),rgba(7,18,31,0.88))]">
          <Badge className="border-primary/30 bg-primary/10 text-primary">Routing control</Badge>
          <h2 className="mt-4 text-3xl font-semibold sm:text-4xl">Manage accounts, targets, and channel reach.</h2>
          <p className="mt-4 max-w-xl text-sm leading-7 text-foreground/68">
            Accounts hold credentials. Targets define where each publish job lands. Keep them
            separated so channels stay reusable and delivery paths remain explicit.
          </p>
          <div className="mt-6 grid gap-3">
            {[
              { icon: KeyRound, label: "Accounts", value: accounts.items.length },
              { icon: RadioTower, label: "Enabled accounts", value: enabledAccounts },
              { icon: Route, label: "Enabled targets", value: enabledTargets },
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

        <section className="grid gap-6 lg:grid-cols-2">
          <Card>
            <div className="flex items-center gap-3">
              <span className="inline-flex size-11 items-center justify-center rounded-2xl border border-primary/25 bg-primary/10 text-primary">
                <Network className="size-5" />
              </span>
              <div>
                <Badge>Channel account</Badge>
                <h3 className="mt-2 text-2xl font-semibold">Add channel account</h3>
              </div>
            </div>
            <div className="mt-6">
              <ChannelAccountForm />
            </div>
          </Card>

          <Card>
            <div className="flex items-center gap-3">
              <span className="inline-flex size-11 items-center justify-center rounded-2xl border border-accent/25 bg-accent/10 text-accent">
                <Route className="size-5" />
              </span>
              <div>
                <Badge className="border-accent/20 bg-accent/10 text-accent">Target</Badge>
                <h3 className="mt-2 text-2xl font-semibold">Add channel target</h3>
              </div>
            </div>
            <div className="mt-6">
              <ChannelTargetForm accounts={accounts.items} />
            </div>
          </Card>
        </section>
      </section>

      <Card>
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <p className="font-mono text-[11px] uppercase tracking-[0.24em] text-foreground/45">Accounts</p>
            <h3 className="mt-2 text-2xl font-semibold">Credential registry</h3>
          </div>
          <Badge className="border-accent/20 bg-accent/10 text-accent">{accounts.items.length} accounts</Badge>
        </div>
        <div className="mt-5 overflow-x-auto">
          <Table>
            <thead>
              <tr>
                <TH>Name</TH>
                <TH>Channel</TH>
                <TH>Secret ref</TH>
                <TH>Config</TH>
                <TH>Enabled</TH>
                <TH>Action</TH>
              </tr>
            </thead>
            <tbody>
              {accounts.items.map((account) => {
                const feishu = parseFeishuAccountConfig(account.configJson);

                return (
                  <tr key={account.id} className="group">
                    <TD className="font-medium">{account.name}</TD>
                    <TD>
                      <Badge className="border-primary/20 bg-primary/10 text-primary">{account.channelType}</Badge>
                    </TD>
                    <TD className="font-mono text-xs text-foreground/62">{account.secretRef}</TD>
                    <TD className="text-xs text-foreground/60">
                      {account.channelType === "feishu"
                        ? [feishu.appIdEnv, feishu.tokenEnv].filter(Boolean).join(" / ") || "-"
                        : "-"}
                    </TD>
                    <TD>
                      <Badge
                        className={
                          account.enabled
                            ? "border-primary/30 bg-primary/10 text-primary"
                            : "border-border bg-muted text-foreground/68"
                        }
                      >
                        {account.enabled ? "enabled" : "disabled"}
                      </Badge>
                    </TD>
                    <TD>
                      <DeleteButton
                        label={`account ${account.name}`}
                        path={`/channel-accounts/${account.id}`}
                      />
                    </TD>
                  </tr>
                );
              })}
            </tbody>
          </Table>
        </div>
      </Card>

      <Card>
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <p className="font-mono text-[11px] uppercase tracking-[0.24em] text-foreground/45">Targets</p>
            <h3 className="mt-2 text-2xl font-semibold">Delivery routes</h3>
          </div>
          <Badge className="border-accent/20 bg-accent/10 text-accent">{targets.items.length} targets</Badge>
        </div>
        <div className="mt-5 overflow-x-auto">
          <Table>
            <thead>
              <tr>
                <TH>Name</TH>
                <TH>Channel</TH>
                <TH>Group</TH>
                <TH>Subtype</TH>
                <TH>Route key</TH>
                <TH>Account</TH>
                <TH>Enabled</TH>
                <TH>Action</TH>
              </tr>
            </thead>
            <tbody>
              {targets.items.map((target) => {
                const description = describeTarget(target);

                return (
                  <tr key={target.id} className="group">
                    <TD className="font-medium">{target.targetName}</TD>
                    <TD>
                      <Badge className="border-primary/20 bg-primary/10 text-primary">{description.channelLabel}</Badge>
                    </TD>
                    <TD className="font-mono text-xs text-foreground/62">{description.primary}</TD>
                    <TD className="text-foreground/68">{description.secondary}</TD>
                    <TD className="font-mono text-xs text-foreground/62">{target.targetKey}</TD>
                    <TD className="text-foreground/68">{target.channelAccountId}</TD>
                    <TD>
                      <Badge
                        className={
                          target.enabled
                            ? "border-primary/30 bg-primary/10 text-primary"
                            : "border-border bg-muted text-foreground/68"
                        }
                      >
                        {target.enabled ? "enabled" : "disabled"}
                      </Badge>
                    </TD>
                    <TD>
                      <DeleteButton
                        label={`target ${target.targetName}`}
                        path={`/channel-targets/${target.id}`}
                      />
                    </TD>
                  </tr>
                );
              })}
            </tbody>
          </Table>
        </div>
      </Card>
    </Shell>
  );
}
