import { Shell } from "@/components/layout/shell";
import { ChannelAccountForm } from "@/components/forms/channel-account-form";
import { ChannelTargetForm } from "@/components/forms/channel-target-form";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { Table, TD, TH } from "@/components/ui/table";
import { getChannelAccounts, getChannelTargets } from "@/lib/api";

export default async function ChannelsPage() {
  const [accounts, targets] = await Promise.all([
    getChannelAccounts(),
    getChannelTargets(),
  ]);

  return (
    <Shell>
      <section className="grid gap-6 lg:grid-cols-2">
        <Card>
          <Badge>Telegram account</Badge>
          <h2 className="mt-3 text-2xl font-semibold">Add bot account</h2>
          <div className="mt-6">
            <ChannelAccountForm />
          </div>
        </Card>
        <Card>
          <Badge>Target</Badge>
          <h2 className="mt-3 text-2xl font-semibold">Add group target</h2>
          <div className="mt-6">
            <ChannelTargetForm accounts={accounts.items} />
          </div>
        </Card>
      </section>

      <Card>
        <h3 className="text-xl font-semibold">Accounts</h3>
        <div className="mt-5 overflow-x-auto">
          <Table>
            <thead>
              <tr>
                <TH>Name</TH>
                <TH>Channel</TH>
                <TH>Secret ref</TH>
                <TH>Enabled</TH>
              </tr>
            </thead>
            <tbody>
              {accounts.items.map((account) => (
                <tr key={account.id}>
                  <TD>{account.name}</TD>
                  <TD>{account.channelType}</TD>
                  <TD className="font-mono text-xs">{account.secretRef}</TD>
                  <TD>{String(account.enabled)}</TD>
                </tr>
              ))}
            </tbody>
          </Table>
        </div>
      </Card>

      <Card>
        <h3 className="text-xl font-semibold">Targets</h3>
        <div className="mt-5 overflow-x-auto">
          <Table>
            <thead>
              <tr>
                <TH>Name</TH>
                <TH>Chat ID</TH>
                <TH>Account</TH>
                <TH>Enabled</TH>
              </tr>
            </thead>
            <tbody>
              {targets.items.map((target) => (
                <tr key={target.id}>
                  <TD>{target.targetName}</TD>
                  <TD className="font-mono text-xs">{target.targetKey}</TD>
                  <TD>{target.channelAccountId}</TD>
                  <TD>{String(target.enabled)}</TD>
                </tr>
              ))}
            </tbody>
          </Table>
        </div>
      </Card>
    </Shell>
  );
}
