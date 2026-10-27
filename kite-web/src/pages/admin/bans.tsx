import AdminPage from "@/components/app/AdminPage";
import LoadingButton from "@/components/common/LoadingButton";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { useUserBanDeleteMutation } from "@/lib/api/mutations";
import { useUserBans } from "@/lib/hooks/api";
import { UserBan } from "@/lib/types/wire.gen";
import { toast } from "sonner";

// AdminBansPage is where admins see who is banned from features, like the
// marketplace, and lift bans. Users are banned from the reviews and reports.
export default function AdminBansPage() {
  return (
    <AdminPage
      title="Bans"
      description="Users banned from features, newest first. Ban the owner of a template from its review or report."
      className="container max-w-3xl py-12"
    >
      <BanList />
    </AdminPage>
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
