import { useAppTokenUpdateMutation } from "@/lib/api/mutations";
import { App } from "@/lib/types/wire.gen";
import { ExternalLinkIcon } from "lucide-react";
import { FormEvent, ReactNode, useState } from "react";
import { toast } from "sonner";
import LoadingButton from "../common/LoadingButton";
import { Button } from "../ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../ui/dialog";
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import { useTrack } from "@/lib/analytics";

export default function AppTokenUpdateDialog({
  app,
  from,
  children,
}: {
  app: App;
  // Where the dialog was opened, for analytics.
  from: "popup" | "list" | "settings";
  children: ReactNode;
}) {
  const [open, setOpen] = useState(false);
  const [token, setToken] = useState("");
  const [error, setError] = useState<string | null>(null);

  const updateMutation = useAppTokenUpdateMutation(app.id);
  const track = useTrack();

  function onOpenChange(open: boolean) {
    setOpen(open);
    if (!open) {
      setToken("");
      setError(null);
    }
  }

  function onSubmit(e: FormEvent) {
    e.preventDefault();
    const discordToken = token.trim();
    if (!discordToken) return;

    updateMutation.mutate(
      { discord_token: discordToken },
      {
        onSuccess(res) {
          if (res.success) {
            track("token_updated", { from });
            toast.success("Token updated! Your bot is starting again.");
            onOpenChange(false);
          } else {
            setError(res.error.message);
          }
        },
        onError(err) {
          setError(String(err));
        },
      }
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent>
        <form onSubmit={onSubmit} className="space-y-5">
          <DialogHeader>
            <DialogTitle>Update Token</DialogTitle>
            <DialogDescription>
              Open your bot in the Discord Developer Portal, click on{" "}
              <strong>Reset Token</strong> under <strong>Bot</strong> and copy
              the new token.
            </DialogDescription>
          </DialogHeader>
          <Button variant="outline" asChild>
            <a
              href={`https://discord.com/developers/applications/${app.discord_id}/bot`}
              target="_blank"
              rel="noreferrer"
            >
              Open Developer Portal
              <ExternalLinkIcon className="h-4 w-4 ml-2" />
            </a>
          </Button>
          <div className="space-y-2">
            <Label htmlFor="discord-token">Bot Token</Label>
            <Input
              id="discord-token"
              type="password"
              autoComplete="off"
              value={token}
              onChange={(e) => setToken(e.target.value)}
            />
            {error && <p className="text-sm text-destructive">{error}</p>}
          </div>
          <DialogFooter>
            <LoadingButton type="submit" loading={updateMutation.isPending}>
              Update token
            </LoadingButton>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
