import { useAppDuplicateMutation } from "@/lib/api/mutations";
import { App } from "@/lib/types/wire.gen";
import { CheckIcon, ExternalLinkIcon, XIcon } from "lucide-react";
import { useRouter } from "next/router";
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

const copied = [
  "Commands and event listeners",
  "Message templates",
  "Variables, without their values",
  "Plugins and their settings",
  "Secrets and your custom status",
];

const notCopied = [
  "Variable values, logs and usage",
  "Collaborators",
  "Integrations, connect them again",
  "Premium, the copy starts on the free plan",
];

export default function AppDuplicateDialog({
  app,
  children,
}: {
  app: App;
  children: ReactNode;
}) {
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [token, setToken] = useState("");
  const [error, setError] = useState<string | null>(null);

  const duplicateMutation = useAppDuplicateMutation(app.id);

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

    duplicateMutation.mutate(
      { discord_token: discordToken },
      {
        onSuccess(res) {
          if (res.success) {
            toast.success(`${app.name} was copied to ${res.data.name}!`);
            onOpenChange(false);
            router.push(`/apps/${res.data.id}`);
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
            <DialogTitle>Duplicate App</DialogTitle>
            <DialogDescription>
              The copy needs its own bot. Create a new application in the
              Discord Developer Portal, click on <strong>Reset Token</strong>{" "}
              under <strong>Bot</strong> and paste its token here. It can&apos;t
              be the token of this app&apos;s bot.
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 sm:grid-cols-2 text-sm">
            <div className="space-y-2">
              <div className="font-medium">Copied</div>
              {copied.map((item) => (
                <div key={item} className="flex gap-2 text-muted-foreground">
                  <CheckIcon className="h-4 w-4 flex-none mt-0.5 text-green-600 dark:text-green-400" />
                  <span>{item}</span>
                </div>
              ))}
            </div>
            <div className="space-y-2">
              <div className="font-medium">Not copied</div>
              {notCopied.map((item) => (
                <div key={item} className="flex gap-2 text-muted-foreground">
                  <XIcon className="h-4 w-4 flex-none mt-0.5" />
                  <span>{item}</span>
                </div>
              ))}
            </div>
          </div>

          <Button variant="outline" asChild>
            <a
              href="https://discord.com/developers/applications"
              target="_blank"
              rel="noreferrer"
            >
              Open Developer Portal
              <ExternalLinkIcon className="h-4 w-4 ml-2" />
            </a>
          </Button>
          <div className="space-y-2">
            <Label htmlFor="duplicate-discord-token">New Bot Token</Label>
            <Input
              id="duplicate-discord-token"
              type="password"
              autoComplete="off"
              value={token}
              onChange={(e) => setToken(e.target.value)}
            />
            {error && <p className="text-sm text-destructive">{error}</p>}
          </div>
          <DialogFooter>
            <LoadingButton type="submit" loading={duplicateMutation.isPending}>
              Duplicate app
            </LoadingButton>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
