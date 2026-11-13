import AdminPage from "@/components/app/AdminPage";
import { UserBadges } from "@/components/badges/Badges";
import LoadingButton from "@/components/common/LoadingButton";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  useUserBadgeGrantCreateMutation,
  useUserBadgeGrantDeleteMutation,
} from "@/lib/api/mutations";
import { useUserBadgeGrants } from "@/lib/hooks/api";
import { UserBadgeGrant } from "@/lib/types/wire.gen";
import { useState } from "react";
import { toast } from "sonner";

// AdminBadgesPage is where admins give users the staff and contributor badges.
// The other badges follow from what users did, like subscribing or listing
// templates.
export default function AdminBadgesPage() {
  return (
    <AdminPage
      title="Badges"
      description="Users with the Staff or Contributor badge, newest first. The other badges come from plans and templates on their own."
      className="container max-w-3xl py-12"
    >
      <div className="space-y-8">
        <GrantForm />
        <GrantList />
      </div>
    </AdminPage>
  );
}

// grantableBadges are the badges admins give here, with who they're for.
const grantableBadges = {
  staff: {
    title: "Staff",
    description:
      "For the Kite team, including the staff of the Kite Discord server.",
  },
  contributor: {
    title: "Contributor",
    description: "For people who helped build Kite.",
  },
};

function GrantForm() {
  const grantMutation = useUserBadgeGrantCreateMutation();
  const [kind, setKind] = useState<keyof typeof grantableBadges>("staff");
  const [discordId, setDiscordId] = useState("");
  const badge = grantableBadges[kind];

  async function grant() {
    const res = await grantMutation.mutateAsync({
      discord_id: discordId.trim(),
      kind,
    });
    if (!res.success) {
      toast.error(
        `Failed to give it: ${res.error.message} (${res.error.code})`
      );
      return;
    }
    toast.success(
      `Gave ${res.data.user?.display_name} the ${badge.title} badge`
    );
    setDiscordId("");
  }

  return (
    <Card className="p-5 space-y-3">
      <div>
        <div className="font-semibold">Give a badge</div>
        <p className="text-sm text-muted-foreground">
          {badge.description} It shows on their creator page and in their
          dashboard.
        </p>
      </div>
      <Select
        value={kind}
        onValueChange={(v) => setKind(v as keyof typeof grantableBadges)}
      >
        <SelectTrigger aria-label="Badge">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {Object.entries(grantableBadges).map(([value, b]) => (
            <SelectItem key={value} value={value}>
              {b.title}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <Input
        value={discordId}
        inputMode="numeric"
        placeholder="Discord ID of the user"
        onChange={(e) => setDiscordId(e.target.value)}
      />
      <LoadingButton
        loading={grantMutation.isPending}
        disabled={!/^\d+$/.test(discordId.trim())}
        onClick={grant}
      >
        Give badge
      </LoadingButton>
    </Card>
  );
}

function GrantList() {
  const grants = useUserBadgeGrants();
  if (grants && !grants.length) {
    return (
      <p className="text-sm text-muted-foreground">
        No one has a badge from here yet.
      </p>
    );
  }
  return (
    <div className="space-y-4">
      {grants?.map(
        (g) => g && <Grant key={`${g.user_id}/${g.kind}`} grant={g} />
      )}
    </div>
  );
}

function Grant({ grant }: { grant: UserBadgeGrant }) {
  const deleteMutation = useUserBadgeGrantDeleteMutation();

  async function revoke() {
    const res = await deleteMutation.mutateAsync({
      userId: grant.user_id,
      kind: grant.kind,
    });
    if (!res.success) {
      toast.error(
        `Failed to take it away: ${res.error.message} (${res.error.code})`
      );
      return;
    }
    toast.success(
      `Took the badge from ${grant.user?.display_name ?? "the user"}`
    );
  }

  return (
    <Card className="p-5 flex flex-wrap items-center gap-3">
      <UserBadges badges={[{ kind: grant.kind }]} size="sm" />
      <span className="font-semibold">{grant.user?.display_name}</span>
      <span className="text-sm text-muted-foreground">
        @{grant.user?.discord_username}
      </span>
      <span className="text-xs text-muted-foreground">
        since {new Date(grant.created_at).toLocaleDateString()}
      </span>
      <LoadingButton
        variant="outline"
        className="ml-auto"
        loading={deleteMutation.isPending}
        onClick={revoke}
      >
        Take away
      </LoadingButton>
    </Card>
  );
}
