import { useAppTokenUpdateMutation } from "@/lib/api/mutations";
import { App } from "@/lib/types/wire.gen";
import { ExternalLinkIcon } from "lucide-react";
import { FormEvent, ReactNode, useState } from "react";
import { toast } from "sonner";
import LoadingButton from "../common/LoadingButton";
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "../ui/alert-dialog";
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
import { track } from "@/lib/analytics";
import AppTokenConsent from "./AppTokenConsent";

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
  // The other bot the token belongs to, while the owner confirms moving the
  // app to it.
  const [switchTo, setSwitchTo] = useState<string | null>(null);

  const updateMutation = useAppTokenUpdateMutation(app.id);

  function onOpenChange(open: boolean) {
    setOpen(open);
    if (!open) {
      setToken("");
      setError(null);
    }
  }

  function save(changeApp: boolean) {
    const discordToken = token.trim();
    if (!discordToken) return;

    updateMutation.mutate(
      { discord_token: discordToken, change_app: changeApp },
      {
        onSuccess(res) {
          if (res.success) {
            track("token_updated", { from });
            if (changeApp) {
              toast.success(
                `Your app now runs ${res.data.name}. Add it to your servers to use it there.`
              );
            } else {
              toast.success(
                app.no_credits
                  ? "Token updated. Your bot is starting again, and its commands and events run once it has credits."
                  : "Token updated. Your bot is starting again."
              );
            }
            onOpenChange(false);
          } else if (res.error.code === "discord_app_changed") {
            setSwitchTo(res.error.data.name);
          } else {
            setError(res.error.message);
          }
        },
        onError(err) {
          setError(String(err));
        },
        onSettled() {
          if (changeApp) setSwitchTo(null);
        },
      }
    );
  }

  function onSubmit(e: FormEvent) {
    e.preventDefault();
    save(false);
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent>
        <form onSubmit={onSubmit} className="space-y-5">
          <DialogHeader>
            <DialogTitle>Update token</DialogTitle>
            <DialogDescription>
              Open your bot in the Discord Developer Portal, click on{" "}
              <strong>Reset token</strong> under <strong>Bot</strong> and copy
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
            <Label htmlFor="discord-token">Bot token</Label>
            <Input
              id="discord-token"
              type="password"
              autoComplete="off"
              value={token}
              onChange={(e) => setToken(e.target.value)}
            />
            {error && <p className="text-sm text-destructive">{error}</p>}
          </div>
          <AppTokenConsent action="updating the token" />
          <DialogFooter>
            <LoadingButton type="submit" loading={updateMutation.isPending}>
              Update token
            </LoadingButton>
          </DialogFooter>
        </form>
      </DialogContent>

      <AlertDialog
        open={switchTo !== null}
        onOpenChange={(open) => !open && setSwitchTo(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Switch to {switchTo}?</AlertDialogTitle>
            <AlertDialogDescription>
              This token belongs to a different bot. Your app can move to it
              with its commands, event listeners, message templates, stored
              variables, and plan. It also takes the new bot&apos;s name and
              description.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <ul className="list-disc pl-5 space-y-1 text-sm text-muted-foreground">
            <li>Add {switchTo} to your servers again.</li>
            <li>
              Buttons and select menus on messages the old bot sent stop
              working. Send those messages again.
            </li>
            <li>
              The new bot can&apos;t edit or delete the old bot&apos;s messages.
            </li>
            <li>Custom emojis of the old bot don&apos;t work anymore.</li>
            <li>The old bot goes offline and its commands are removed.</li>
          </ul>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <LoadingButton
              loading={updateMutation.isPending}
              onClick={() => save(true)}
            >
              Switch to {switchTo}
            </LoadingButton>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </Dialog>
  );
}
