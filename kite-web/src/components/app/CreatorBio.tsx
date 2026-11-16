import { useEffect, useState } from "react";
import { toast } from "sonner";
import { PencilIcon, PlusIcon } from "lucide-react";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "../ui/dialog";
import { Button } from "../ui/button";
import { Textarea } from "../ui/textarea";
import LoadingButton from "../common/LoadingButton";
import { useUserQuery } from "@/lib/api/queries";
import { useUserUpdateMutation } from "@/lib/api/mutations";
import { cn } from "@/lib/utils";

// The limits of the service, see MaxUserBioLength and MaxUserBioLines.
export const maxBioLength = 300;
export const maxBioLines = 5;

// bioLength counts characters like the service does, so an emoji is one.
export function bioLength(bio: string) {
  return Array.from(bio).length;
}

function bioLines(bio: string) {
  return bio.split("\n").length;
}

// CreatorBio shows the bio of a creator on their page, and lets them change
// it. The page is rendered on the server and cached, so a bio they saved is
// shown from here until the page is rendered again.
export default function CreatorBio({
  creatorId,
  bio: initialBio,
  className,
}: {
  creatorId: string;
  bio?: string;
  className?: string;
}) {
  const { data: user } = useUserQuery();
  const [bio, setBio] = useState(initialBio ?? "");
  const [open, setOpen] = useState(false);

  useEffect(() => setBio(initialBio ?? ""), [initialBio]);

  const isOwn = !!user?.success && user.data.discord_id === creatorId;

  if (!bio && !isOwn) return null;

  return (
    <div className={cn("max-w-2xl", className)}>
      {bio && (
        <p className="text-foreground/90 whitespace-pre-line break-words leading-relaxed">
          {bio}
        </p>
      )}
      {isOwn && (
        <Button
          variant="ghost"
          size="sm"
          className={cn("-ml-3 text-muted-foreground", bio && "mt-2")}
          onClick={() => setOpen(true)}
        >
          {bio ? (
            <PencilIcon className="mr-2 h-3.5 w-3.5" />
          ) : (
            <PlusIcon className="mr-2 h-3.5 w-3.5" />
          )}
          {bio ? "Edit bio" : "Add a bio"}
        </Button>
      )}
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          {open && (
            <CreatorBioForm
              bio={bio}
              onSaved={(b) => {
                setBio(b);
                setOpen(false);
              }}
            />
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}

function CreatorBioForm({
  bio: initialBio,
  onSaved,
}: {
  bio: string;
  onSaved: (bio: string) => void;
}) {
  const mutation = useUserUpdateMutation();
  const [bio, setBio] = useState(initialBio);

  const length = bioLength(bio);
  const lines = bioLines(bio);
  const tooLong = length > maxBioLength;
  const tooManyLines = lines > maxBioLines;

  async function save() {
    const res = await mutation.mutateAsync({ bio });
    if (!res.success) {
      toast.error(
        `Failed to save your bio: ${res.error.message} (${res.error.code})`
      );
      return;
    }
    toast.success(res.data.bio ? "Saved your bio" : "Removed your bio");
    onSaved(res.data.bio);
  }

  return (
    <>
      <DialogHeader>
        <DialogTitle>Your bio</DialogTitle>
        <DialogDescription>
          Shown on your creator page in the Marketplace, to everyone. Say what
          you make or where people can find you.
        </DialogDescription>
      </DialogHeader>
      <div>
        <Textarea
          value={bio}
          rows={5}
          autoFocus
          placeholder="I make moderation and welcome templates for community servers."
          onChange={(e) => setBio(e.target.value)}
          className="resize-none"
        />
        <div className="flex justify-between mt-2 text-xs">
          <span
            className={cn(
              "text-muted-foreground",
              tooManyLines && "text-destructive"
            )}
          >
            {tooManyLines
              ? `Use at most ${maxBioLines} lines`
              : "Plain text, links aren't clickable"}
          </span>
          <span
            className={cn(
              "tabular-nums text-muted-foreground",
              tooLong && "text-destructive"
            )}
          >
            {length}/{maxBioLength}
          </span>
        </div>
      </div>
      <DialogFooter className="gap-2">
        <DialogClose asChild>
          <Button variant="outline">Cancel</Button>
        </DialogClose>
        <LoadingButton
          onClick={save}
          loading={mutation.isPending}
          disabled={tooLong || tooManyLines || bio === initialBio}
        >
          Save
        </LoadingButton>
      </DialogFooter>
    </>
  );
}
