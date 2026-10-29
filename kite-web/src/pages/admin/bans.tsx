import AdminPage from "@/components/app/AdminPage";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import LoadingButton from "@/components/common/LoadingButton";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import {
  useUserBanCreateMutation,
  useUserBanDeleteMutation,
} from "@/lib/api/mutations";
import { useUserBans } from "@/lib/hooks/api";
import { UserBan } from "@/lib/types/wire.gen";
import { useState } from "react";
import { toast } from "sonner";

// AdminBansPage is where admins see who is banned from features, like the
// marketplace, and lift bans. Users are also banned from the marketplace
// from the reviews and reports.
export default function AdminBansPage() {
  return (
    <AdminPage
      title="Bans"
      description="Users banned from features, newest first. The owner of a template can also be banned from its review or report."
      className="container max-w-3xl py-12"
    >
      <div className="space-y-8">
        <BanForm />
        <BanList />
      </div>
    </AdminPage>
  );
}

// banFeatures are what admins can ban a user from here, with what it does.
const banFeatures = {
  kite: {
    label: "Kite",
    description:
      "They can't use Kite anymore and the apps they own are stopped until they're unbanned.",
    confirm: "This logs them out everywhere and stops every app they own.",
  },
  marketplace: {
    label: "the marketplace",
    description:
      "They can't submit templates and new versions of theirs aren't reviewed. What they listed stays listed.",
    confirm: "This rejects the templates they have in review.",
  },
};

// BanForm bans a user, found by their Discord ID or that of one of their
// bots, like a bot someone reported.
function BanForm() {
  const banMutation = useUserBanCreateMutation();
  const [feature, setFeature] = useState<keyof typeof banFeatures>("kite");
  const [discordId, setDiscordId] = useState("");
  const [reason, setReason] = useState("");
  const { label, description, confirm } = banFeatures[feature];

  async function ban() {
    const res = await banMutation.mutateAsync({
      discord_id: discordId.trim(),
      feature,
      reason: reason.trim(),
    });
    if (!res.success) {
      toast.error(`Failed to ban: ${res.error.message} (${res.error.code})`);
      return;
    }
    toast.success(`Banned ${res.data.user?.display_name} from ${label}`);
    setDiscordId("");
    setReason("");
  }

  return (
    <Card className="p-5 space-y-3">
      <div>
        <div className="font-semibold">Ban a user</div>
        <p className="text-sm text-muted-foreground">{description}</p>
      </div>
      <Select
        value={feature}
        onValueChange={(v) => setFeature(v as keyof typeof banFeatures)}
      >
        <SelectTrigger aria-label="Ban from">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="kite">Kite</SelectItem>
          <SelectItem value="marketplace">Marketplace</SelectItem>
        </SelectContent>
      </Select>
      <Input
        value={discordId}
        inputMode="numeric"
        placeholder="Discord ID of the user or of one of their bots"
        onChange={(e) => setDiscordId(e.target.value)}
      />
      <Textarea
        value={reason}
        maxLength={1000}
        rows={3}
        placeholder="Why they're banned, they see this"
        onChange={(e) => setReason(e.target.value)}
      />
      <ConfirmDialog
        title={`Ban from ${label}?`}
        description={confirm}
        onConfirm={ban}
      >
        <LoadingButton
          variant="destructive"
          loading={banMutation.isPending}
          disabled={!/^\d+$/.test(discordId.trim()) || !reason.trim()}
        >
          Ban from {label}
        </LoadingButton>
      </ConfirmDialog>
    </Card>
  );
}

function BanList() {
  const bans = useUserBans();
  if (bans && !bans.length) {
    return <p className="text-sm text-muted-foreground">No one is banned.</p>;
  }
  return (
    <div className="space-y-4">
      {bans?.map((b) => b && <Ban key={`${b.user_id}/${b.feature}`} ban={b} />)}
    </div>
  );
}

function Ban({ ban }: { ban: UserBan }) {
  const deleteMutation = useUserBanDeleteMutation();

  async function unban() {
    const res = await deleteMutation.mutateAsync({
      userId: ban.user_id,
      feature: ban.feature,
    });
    if (!res.success) {
      toast.error(`Failed to unban: ${res.error.message} (${res.error.code})`);
      return;
    }
    toast.success(`Unbanned ${ban.user?.display_name ?? "the user"}`);
  }

  return (
    <Card className="p-5 space-y-3">
      <div className="flex flex-wrap items-center gap-2">
        <span className="font-semibold">{ban.user?.display_name}</span>
        <span className="text-sm text-muted-foreground">
          @{ban.user?.discord_username}
        </span>
        <Badge variant="outline">{ban.feature}</Badge>
      </div>
      <blockquote className="border-l-2 pl-3 text-sm whitespace-pre-wrap">
        {ban.reason}
      </blockquote>
      <div className="text-xs text-muted-foreground">
        Banned {new Date(ban.created_at).toLocaleString()}
      </div>
      <LoadingButton
        variant="outline"
        loading={deleteMutation.isPending}
        onClick={unban}
      >
        Unban
      </LoadingButton>
    </Card>
  );
}
