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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import ColorPicker from "@/tools/common/components/ColorPicker";
import LoadingButton from "../common/LoadingButton";
import BotProfilePreview from "./BotProfilePreview";
import { nameStyleCss } from "./nameStyleFonts";
import { useAppStateGuildProfileUpdateMutation } from "@/lib/api/mutations";
import {
  useAppStateGuildProfileQuery,
  useAppStateUserQuery,
} from "@/lib/api/queries";
import {
  useApp,
  useAppFeature,
  useFeaturePlan,
  useResponseData,
} from "@/lib/hooks/api";
import { useAppId } from "@/lib/hooks/params";
import { profileImageTypes, readProfileImage } from "@/lib/profileImage";
import {
  colorsForEffect,
  defaultNameStyleColors,
  nameStyleEffects,
  nameStyleFonts,
} from "@/lib/nameStyle";
import {
  Guild,
  NameStyle,
  StateGuildProfileUpdateRequest,
} from "@/lib/types/wire.gen";

// A name style being edited, its colors are undefined while half typed.
type NameStyleDraft = Omit<NameStyle, "colors"> & {
  colors: (number | undefined)[];
};

// The fields start empty and only the ones filled in are changed. The preview
// shows the current profile with the changes.
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
  const plan = useFeaturePlan("server_profiles");
  const mutation = useAppStateGuildProfileUpdateMutation(appId);

  const app = useApp();
  const user = useResponseData(useAppStateUserQuery(appId));
  const current = useResponseData(
    useAppStateGuildProfileQuery(appId, open ? guild.id : "")
  );

  const [nick, setNick] = useState("");
  const [bio, setBio] = useState("");
  const [avatar, setAvatar] = useState<string>();
  const [banner, setBanner] = useState<string>();
  const [nameStyle, setNameStyle] = useState<NameStyleDraft>();

  function onImageChange(set: (v: string | undefined) => void) {
    return async (e: ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0];
      const image = file && (await readProfileImage(file));
      if (!image) e.target.value = "";
      set(image);
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
    onDialogOpenChange(false);
  }

  // Changes that weren't saved are dropped on close, as the file inputs are
  // empty when the dialog opens again.
  function onDialogOpenChange(open: boolean) {
    if (!open) {
      setNick("");
      setBio("");
      setAvatar(undefined);
      setBanner(undefined);
      setNameStyle(undefined);
    }
    onOpenChange(open);
  }

  const newNameStyle =
    nameStyle && nameStyle.colors.every((c) => c !== undefined)
      ? { ...nameStyle, colors: nameStyle.colors as number[] }
      : undefined;

  function onSave() {
    const req: StateGuildProfileUpdateRequest = {
      nick: nick || undefined,
      bio: bio || undefined,
      avatar,
      name_style: newNameStyle,
    };
    // Each image goes in a request of its own to stay below the body limit.
    save(banner ? [req, { banner }] : [req]);
  }

  function onReset() {
    if (!confirm("Reset the bot's profile in this server to its own?")) return;
    save([{ reset: true }]);
  }

  const changed = !!(nick || bio || avatar || banner || nameStyle);
  const shownNameStyle = newNameStyle || current?.name_style;

  return (
    <Dialog open={open} onOpenChange={onDialogOpenChange}>
      <DialogContent className="max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Server profile in {guild.name}</DialogTitle>
          <DialogDescription>
            Change how the bot looks in this server. Fields you leave empty stay
            as they are.
          </DialogDescription>
        </DialogHeader>

        <BotProfilePreview
          className="max-w-none"
          name={nick || current?.nick || user?.username || app?.name || ""}
          username={user?.username}
          avatarUrl={avatar || current?.avatar_url || user?.avatar_url}
          bannerUrl={banner || current?.banner_url || user?.banner_url}
          about={bio || app?.description || undefined}
          nameStyle={shownNameStyle && nameStyleCss(shownNameStyle)}
        />

        {available ? (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="profile-nick">Nickname</Label>
              <Input
                id="profile-nick"
                value={nick}
                placeholder={current?.nick || user?.username}
                maxLength={32}
                onChange={(e) => setNick(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="profile-avatar">Avatar</Label>
              <Input
                id="profile-avatar"
                type="file"
                accept={profileImageTypes}
                onChange={onImageChange(setAvatar)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="profile-banner">Banner</Label>
              <Input
                id="profile-banner"
                type="file"
                accept={profileImageTypes}
                onChange={onImageChange(setBanner)}
              />
            </div>
            <NameStyleFields
              value={nameStyle}
              current={current?.name_style}
              onChange={setNameStyle}
            />
            <div className="space-y-2">
              <Label htmlFor="profile-bio">About Me</Label>
              <p className="text-sm text-muted-foreground">
                Discord doesn&apos;t share the About Me of a server, so the
                preview shows your bot&apos;s own until you type one.
              </p>
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
              {plan ? `Kite ${plan.title}` : "a paid plan"}
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
              disabled={!changed || (!!nameStyle && !newNameStyle)}
            >
              Save
            </LoadingButton>
          </DialogFooter>
        )}
      </DialogContent>
    </Dialog>
  );
}

// The style starts as the current one once a font is picked, and "Keep
// current" drops the changes.
function NameStyleFields({
  value,
  current,
  onChange,
}: {
  value?: NameStyleDraft;
  current?: NameStyle;
  onChange: (value: NameStyleDraft | undefined) => void;
}) {
  function setFont(font: string) {
    if (font === "keep") {
      onChange(undefined);
      return;
    }
    onChange({
      effect: "solid",
      colors: [defaultNameStyleColors[0]],
      ...(value || current),
      font,
    });
  }

  function setEffect(effect: string) {
    if (!value) return;
    onChange({
      ...value,
      effect,
      colors: colorsForEffect(
        effect,
        value.colors,
        (i) => defaultNameStyleColors[i]
      ),
    });
  }

  function setColor(i: number, color: number | undefined) {
    if (!value) return;
    onChange({
      ...value,
      colors: value.colors.map((c, j) => (j === i ? color : c)),
    });
  }

  return (
    <div className="space-y-2">
      <Label htmlFor="profile-name-font">Name style</Label>
      <p className="text-sm text-muted-foreground">
        Servers show the font. The effect and colors show in the bot&apos;s
        profile.
      </p>
      <div className="flex gap-2">
        <Select value={value?.font || "keep"} onValueChange={setFont}>
          <SelectTrigger id="profile-name-font" aria-label="Font">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="keep">Keep current</SelectItem>
            {nameStyleFonts.map((f) => (
              <SelectItem key={f.value} value={f.value}>
                {f.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        {value && (
          <Select value={value.effect} onValueChange={setEffect}>
            <SelectTrigger aria-label="Effect">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {nameStyleEffects.map((e) => (
                <SelectItem key={e.value} value={e.value}>
                  {e.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}
      </div>
      {value?.colors.map((color, i) => (
        <ColorPicker key={i} value={color} onChange={(c) => setColor(i, c)} />
      ))}
    </div>
  );
}
