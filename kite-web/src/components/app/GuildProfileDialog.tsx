import Link from "next/link";
import { ChangeEvent, useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import LoadingButton from "../common/LoadingButton";
import { useAppStateGuildProfileUpdateMutation } from "@/lib/api/mutations";
import { useAppFeature } from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { Guild, StateGuildProfileUpdateRequest } from "@/lib/types/wire.gen";

// Requests are sent as JSON with the images as data URIs, which the API's body
// limit has to fit.
const maxImageSize = 5 * 1024 * 1024;
const imageTypes = "image/png,image/jpeg,image/gif,image/webp";

function readDataUrl(file: File) {
  return new Promise<string>((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(reader.result as string);
    reader.onerror = () => reject(reader.error);
    reader.readAsDataURL(file);
  });
}

// Discord doesn't return the bot's server profile, so the fields start empty
// and only the ones filled in are changed.
export default function GuildProfileDialog({
  guild,
  open,
  onOpenChange,
}: {
  guild: Guild;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const appId = useAppId();
  const available = !!useAppFeature((f) => f.server_profiles);
  const mutation = useAppStateGuildProfileUpdateMutation(appId);

  const [nick, setNick] = useState("");
  const [bio, setBio] = useState("");
  const [avatar, setAvatar] = useState<string>();
  const [banner, setBanner] = useState<string>();

  function onImageChange(set: (v: string | undefined) => void) {
    return async (e: ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0];
      if (!file) return set(undefined);
      if (file.size > maxImageSize) {
        toast.error("Images can be at most 5 MB");
        e.target.value = "";
        return set(undefined);
      }
      set(await readDataUrl(file));
    };
  }

  async function save(requests: StateGuildProfileUpdateRequest[]) {
    for (const req of requests) {
      const res = await mutation.mutateAsync({ guildId: guild.id, ...req });
      if (!res.success) {
        toast.error(
          `Failed to update profile: ${res.error.message} (${res.error.code})`
        );
        return;
      }
    }

    toast.success("Server profile updated");
    setNick("");
    setBio("");
    setAvatar(undefined);
    setBanner(undefined);
    onOpenChange(false);
  }

  function onSave() {
    const req: StateGuildProfileUpdateRequest = {
      nick: nick || undefined,
      bio: bio || undefined,
      avatar,
    };
    // Each image goes in a request of its own to stay below the body limit.
    save(banner ? [req, { banner }] : [req]);
  }

  function onReset() {
    if (!confirm("Reset the bot's profile in this server to its own?")) return;
    save([{ nick: "", bio: "", avatar: "", banner: "" }]);
  }

  const changed = !!(nick || bio || avatar || banner);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Server profile in {guild.name}</DialogTitle>
          <DialogDescription>
            Change how the bot looks in this server. Fields you leave empty stay
            as they are.
          </DialogDescription>
        </DialogHeader>

        {available ? (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="profile-nick">Nickname</Label>
              <Input
                id="profile-nick"
                value={nick}
                maxLength={32}
                onChange={(e) => setNick(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="profile-avatar">Avatar</Label>
              <Input
                id="profile-avatar"
                type="file"
                accept={imageTypes}
                onChange={onImageChange(setAvatar)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="profile-banner">Banner</Label>
              <Input
                id="profile-banner"
                type="file"
                accept={imageTypes}
                onChange={onImageChange(setBanner)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="profile-bio">About Me</Label>
              <Textarea
                id="profile-bio"
                value={bio}
                maxLength={190}
                onChange={(e) => setBio(e.target.value)}
              />
            </div>
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">
            Server profiles are part of{" "}
            <Link
              href={`/apps/${appId}/premium`}
              className="text-primary hover:underline"
            >
              Premium
            </Link>
            .
          </p>
        )}

        {available && (
          <DialogFooter className="gap-2 sm:justify-between">
            <Button
              variant="outline"
              onClick={onReset}
              disabled={mutation.isPending}
            >
              Reset
            </Button>
            <LoadingButton
              onClick={onSave}
              loading={mutation.isPending}
              disabled={!changed}
            >
              Save
            </LoadingButton>
          </DialogFooter>
        )}
      </DialogContent>
    </Dialog>
  );
}
