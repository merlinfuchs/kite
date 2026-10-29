import AdminPage from "@/components/app/AdminPage";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import LoadingButton from "@/components/common/LoadingButton";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
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
// marketplace, and lift bans. Users are banned from the marketplace from the
// reviews and reports, and from Kite here.
export default function AdminBansPage() {
  return (
    <AdminPage
      title="Bans"
      description="Users banned from features, newest first. Ban the owner of a template from its review or report."
      className="container max-w-3xl py-12"
    >
      <div className="space-y-8">
        <KiteBanForm />
        <BanList />
      </div>
    </AdminPage>
  );
}

// KiteBanForm bans a user from Kite, found by their Discord ID or that of one
// of their bots, like a bot someone reported.
function KiteBanForm() {
  const banMutation = useUserBanCreateMutation();
  const [discordId, setDiscordId] = useState("");
  const [reason, setReason] = useState("");

  async function ban() {
    const res = await banMutation.mutateAsync({
      discord_id: discordId.trim(),
      feature: "kite",
      reason: reason.trim(),
    });
    if (!res.success) {
      toast.error(`Failed to ban: ${res.error.message} (${res.error.code})`);
      return;
    }
    toast.success(`Banned ${res.data.user?.display_name} from Kite`);
    setDiscordId("");
    setReason("");
  }

  return (
    <Card className="p-5 space-y-3">
      <div>
        <div className="font-semibold">Ban from Kite</div>
        <p className="text-sm text-muted-foreground">
          They can&apos;t use Kite anymore and the apps they own are stopped
          until they&apos;re unbanned.
        </p>
      </div>
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
        title="Ban from Kite?"
        description="This logs them out everywhere and stops every app they own."
        onConfirm={ban}
      >
        <LoadingButton
          variant="destructive"
          loading={banMutation.isPending}
          disabled={!/^\d+$/.test(discordId.trim()) || !reason.trim()}
        >
          Ban from Kite
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
