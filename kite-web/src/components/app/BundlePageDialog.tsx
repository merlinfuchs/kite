import env from "@/lib/env/client";
import { ReactNode, useEffect, useMemo, useState } from "react";
import { toast } from "sonner";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../ui/dialog";
import { Button } from "../ui/button";
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import { Textarea } from "../ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import { ToggleGroup, ToggleGroupItem } from "../ui/toggle-group";
import LoadingButton from "../common/LoadingButton";
import {
  useBundleListingSubmitMutation,
  useBundleUpdateMutation,
} from "@/lib/api/mutations";
import {
  inMarketplace,
  maxBundleDescription,
  maxBundleName,
  nameHasSymbols,
  normalizeTag,
} from "@/lib/flow/bundleListing";
import {
  iconName,
  lucideName,
  tagCounts,
  templateCategories,
  templateIcons,
} from "@/lib/flow/templates";
import TemplateIcon, { loadLucide } from "./TemplateIcon";
import { BundleListing, MaxBundleTags, UserBundle } from "@/lib/types/wire.gen";
import { XIcon } from "lucide-react";
import { useTemplates } from "@/lib/hooks/templates";
import { cn } from "@/lib/utils";

// BundlePageDialog edits the page of a bundle the user made. Its link shows
// changes right away, unless it's in the marketplace, where Kite reviews them
// first. With submit, it lists the bundle in the marketplace.
export default function BundlePageDialog({
  bundle,
  submit,
  children,
}: {
  bundle: UserBundle;
  submit?: boolean;
  children: ReactNode;
}) {
  const [open, setOpen] = useState(false);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="max-w-2xl max-h-[90dvh] overflow-y-auto">
        {open && (
          <BundlePageForm
            bundle={bundle}
            submit={submit}
            onDone={() => setOpen(false)}
          />
        )}
      </DialogContent>
    </Dialog>
  );
}

// The most the service allows: characters in a paragraph, and steps before
// starting and characters in each.
const maxParagraph = 1500;
const maxSteps = 5;
const maxStep = 500;

// Paragraphs are separated by an empty line.
function splitParagraphs(text: string) {
  return text
    .split(/\n\s*\n/)
    .map((p) => p.trim())
    .filter(Boolean);
}

function BundlePageForm({
  bundle,
  submit,
  onDone,
}: {
  bundle: UserBundle;
  submit?: boolean;
  onDone: () => void;
}) {
  const updateMutation = useBundleUpdateMutation();
  const submitMutation = useBundleListingSubmitMutation();
  // Listing it, and changes to a bundle in the marketplace, are reviewed.
  const mode = submit ? "list" : inMarketplace(bundle) ? "review" : "save";

  const initial = {
    name: bundle.page.name,
    description: bundle.page.description,
    icon: bundle.page.icon,
    category: bundle.page.category,
    about: bundle.page.about.join("\n\n"),
    setup: bundle.page.setup.join("\n"),
  };
  const [form, setForm] = useState(initial);
  const set = (key: keyof typeof initial) => (value: string) =>
    setForm((f) => ({ ...f, [key]: value }));
  const initialTags = bundle.page.tags ?? [];
  const [tags, setTags] = useState<string[]>(initialTags);
  const changed =
    Object.entries(initial).some(
      ([key, value]) => form[key as keyof typeof initial] !== value
    ) ||
    tags.length !== initialTags.length ||
    tags.some((t, i) => t !== initialTags[i]);

  const paragraphs = splitParagraphs(form.about);
  const tooLong = paragraphs.some((p) => p.length > maxParagraph);
  // One step per line.
  const steps = form.setup
    .split("\n")
    .map((s) => s.trim())
    .filter(Boolean);
  const stepTooLong = steps.some((s) => s.length > maxStep);
  const valid =
    !!form.name.trim() &&
    !nameHasSymbols(form.name) &&
    form.description.length <= maxBundleDescription &&
    paragraphs.length <= 5 &&
    !tooLong &&
    steps.length <= maxSteps &&
    !stepTooLong &&
    (mode === "save" ||
      (!!form.icon && !!form.category && !!paragraphs.length));

  async function save() {
    const page: BundleListing = {
      name: form.name.trim(),
      description: form.description.trim(),
      icon: form.icon,
      category: form.category,
      tags,
      about: paragraphs,
      faq: [],
      setup: steps,
    };
    const res =
      mode === "save"
        ? await updateMutation.mutateAsync({ bundleId: bundle.id, page })
        : await submitMutation.mutateAsync({
            bundleId: bundle.id,
            listing: page,
          });
    if (!res.success) {
      toast.error(
        `Failed to ${mode === "save" ? "save" : "submit"} ${page.name}: ${
          res.error.message
        } (${res.error.code})`
      );
      return;
    }
    toast.success(
      mode === "save"
        ? `Saved ${page.name}`
        : mode === "review"
        ? "Submitted your changes, Kite reviews them soon"
        : `Submitted ${page.name}, Kite reviews it soon`
    );
    onDone();
  }

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          {mode === "list"
            ? `List ${bundle.name} in the marketplace`
            : `Edit the page of ${bundle.name}`}
        </DialogTitle>
        <DialogDescription>
          {mode === "list" ? (
            <>
              Kite reviews version {bundle.version} and this page before
              it&apos;s listed. Once it is, its link leads to the marketplace,
              and new versions are reviewed before apps get them.{" "}
              <a
                href={`${env.NEXT_PUBLIC_DOCS_LINK}/guides/publishing-templates`}
                target="_blank"
                className="text-primary hover:underline"
              >
                How to get it approved
              </a>
            </>
          ) : mode === "review" &&
            bundle.listed_version !== null &&
            bundle.listed_version !== bundle.version ? (
            `Kite reviews your changes with version ${bundle.version}, which apps get once it's approved.`
          ) : mode === "review" ? (
            "Kite reviews your changes before the marketplace shows them. Apps that added it don't get an update for them."
          ) : (
            "Anyone with its link sees your changes right away. Apps that added it don't get an update for them."
          )}
        </DialogDescription>
      </DialogHeader>

      <div className="space-y-5">
        <div className="space-y-1.5">
          <Label htmlFor="page-name">Name</Label>
          <Input
            id="page-name"
            value={form.name}
            maxLength={maxBundleName}
            onChange={(e) => set("name")(e.target.value)}
          />
          {nameHasSymbols(form.name) && (
            <p className="text-sm text-destructive">
              Names can&apos;t have emojis or symbols. Pick an icon for the
              marketplace instead.
            </p>
          )}
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="page-description">Description</Label>
          <Input
            id="page-description"
            value={form.description}
            placeholder="What it does, in one sentence"
            onChange={(e) => set("description")(e.target.value)}
          />
          {form.description.length > maxBundleDescription && (
            <p className="text-xs text-destructive">
              {`At most ${maxBundleDescription} characters, this has ${form.description.length}.`}
            </p>
          )}
        </div>
        <div className="space-y-1.5">
          <Label>Icon</Label>
          <IconPicker value={form.icon} onChange={set("icon")} />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="page-category">Category</Label>
          <Select value={form.category} onValueChange={set("category")}>
            <SelectTrigger id="page-category">
              <SelectValue placeholder="Choose a category" />
            </SelectTrigger>
            <SelectContent>
              {templateCategories.map((c) => (
                <SelectItem key={c.value} value={c.value}>
                  {c.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="page-tags">Tags</Label>
          <TagInput id="page-tags" value={tags} onChange={setTags} />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="page-about">How it works</Label>
          <Textarea
            id="page-about"
            value={form.about}
            rows={6}
            placeholder="What it does and how members use it. Leave an empty line between paragraphs."
            onChange={(e) => set("about")(e.target.value)}
          />
          {paragraphs.length > 5 && (
            <p className="text-xs text-destructive">
              At most 5 paragraphs, this has {paragraphs.length}.
            </p>
          )}
          {tooLong && (
            <p className="text-xs text-destructive">
              A paragraph can have at most {maxParagraph} characters.
            </p>
          )}
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="page-setup">Before you start</Label>
          <Textarea
            id="page-setup"
            value={form.setup}
            rows={2}
            placeholder="Steps outside of Kite it needs, like turning on an intent. One per line, optional."
            onChange={(e) => set("setup")(e.target.value)}
          />
          {steps.length > maxSteps && (
            <p className="text-xs text-destructive">
              At most {maxSteps} lines, this has {steps.length}.
            </p>
          )}
          {stepTooLong && (
            <p className="text-xs text-destructive">
              A line can have at most {maxStep} characters.
            </p>
          )}
        </div>
      </div>

      <DialogFooter className="gap-2">
        <DialogClose asChild>
          <Button variant="outline">Cancel</Button>
        </DialogClose>
        <LoadingButton
          onClick={save}
          loading={updateMutation.isPending || submitMutation.isPending}
          // Listing it doesn't need changes, but a review of nothing does.
          disabled={!valid || (mode !== "list" && !changed)}
        >
          {mode === "save" ? "Save" : "Submit for review"}
        </LoadingButton>
      </DialogFooter>
    </>
  );
}

// suggestedIcons are Lucide components that suit Discord bots, which the
// picker shows after Kite's icons and before all others.
const suggestedIcons = [
  "Shield",
  "ShieldBan",
  "Ban",
  "Lock",
  "Key",
  "Eye",
  "EyeOff",
  "Flag",
  "TriangleAlert",
  "UserX",
  "UserCheck",
  "Hammer",
  "Trash2",
  "Scale",
  "Fingerprint",
  "Users",
  "UserPlus",
  "Heart",
  "HandHeart",
  "PartyPopper",
  "Cake",
  "Trophy",
  "Medal",
  "Crown",
  "Sparkles",
  "Smile",
  "ThumbsUp",
  "Megaphone",
  "Bell",
  "Calendar",
  "CalendarDays",
  "Ticket",
  "Vote",
  "Handshake",
  "MessageCircle",
  "MessagesSquare",
  "Inbox",
  "AtSign",
  "Hash",
  "Mic",
  "Headphones",
  "Radio",
  "Gamepad2",
  "Dices",
  "Puzzle",
  "Swords",
  "Target",
  "Rocket",
  "Ghost",
  "Skull",
  "Cat",
  "Dog",
  "Music",
  "Film",
  "Image",
  "Camera",
  "Palette",
  "Pizza",
  "Coffee",
  "Joystick",
  "Spade",
  "CircleHelp",
  "Book",
  "BookOpen",
  "ClipboardList",
  "ListChecks",
  "ListTodo",
  "FileText",
  "NotebookPen",
  "Search",
  "Link",
  "Globe",
  "Clock",
  "Timer",
  "Hourglass",
  "Calculator",
  "ChartColumn",
  "TrendingUp",
  "Database",
  "Server",
  "Code",
  "Terminal",
  "Settings",
  "Wrench",
  "Zap",
  "Bot",
  "Cpu",
  "Cloud",
  "Sun",
  "Bookmark",
  "Tag",
  "Package",
  "ShoppingCart",
  "Coins",
  "Wallet",
  "Gem",
  "Map",
  "MapPin",
  "Compass",
  "Plane",
  "Car",
  "House",
  "GraduationCap",
  "Briefcase",
  "Leaf",
  "Flame",
  "Snowflake",
  "Repeat",
  "Shuffle",
  "Send",
  "Reply",
  "Quote",
  "Languages",
];

// pickerLimit is how many icons the picker shows at once, Kite's and the
// suggested ones before searching. Searching finds the others.
const pickerLimit = Object.keys(templateIcons).length + suggestedIcons.length;

// IconPicker picks any Lucide icon, found by its name. Kite's template icons
// come first, then the suggested ones.
function IconPicker({
  value,
  onChange,
}: {
  value: string;
  onChange: (icon: string) => void;
}) {
  const [query, setQuery] = useState("");
  const [names, setNames] = useState<string[]>();

  useEffect(() => {
    let live = true;
    loadLucide().then((icons) => {
      if (!live) return;
      // Kite's icons keep their names, like arrow-up-1-0 for ArrowUp10.
      const own = Object.keys(templateIcons);
      const components = new Set([
        ...suggestedIcons.filter((c) => c in icons),
        ...Object.keys(icons),
      ]);
      for (const name of own) components.delete(lucideName(name));
      setNames([...own, ...[...components].map(iconName)]);
    });
    return () => {
      live = false;
    };
  }, []);

  const words = query.trim().toLowerCase().split(/\s+/);
  const matches = (names ?? Object.keys(templateIcons)).filter((name) =>
    words.every((w) => name.includes(w))
  );
  const shown = matches.slice(0, pickerLimit);
  // The picked icon stays visible, also when a search hides it.
  if (value && !shown.includes(value)) shown.unshift(value);

  return (
    <div className="space-y-2">
      <Input
        value={query}
        placeholder="Search icons, like shield or music"
        onChange={(e) => setQuery(e.target.value)}
      />
      <ToggleGroup
        type="single"
        variant="outline"
        value={value}
        onValueChange={(v) => v && onChange(v)}
        className="flex-wrap justify-start max-h-40 overflow-y-auto p-px"
      >
        {shown.map((name) => (
          <ToggleGroupItem
            key={name}
            value={name}
            aria-label={name}
            title={name.replaceAll("-", " ")}
            className="data-[state=on]:border-primary data-[state=on]:bg-primary/15 data-[state=on]:text-primary"
          >
            <TemplateIcon name={name} className="h-4 w-4" />
          </ToggleGroupItem>
        ))}
      </ToggleGroup>
      <p className="text-sm text-muted-foreground">
        {!matches.length
          ? "No icon matches."
          : matches.length > pickerLimit
          ? `Showing ${pickerLimit} of ${matches.length} icons, search to find others.`
          : null}
      </p>
    </div>
  );
}

// TagInput edits the tags of a bundle. A tag is added with Enter or a comma,
// and the last one is removed with Backspace in the empty input. It suggests
// the tags other templates use, so the same thing gets the same tag.
function TagInput({
  id,
  value,
  onChange,
}: {
  id: string;
  value: string[];
  onChange: (value: string[]) => void;
}) {
  const [text, setText] = useState("");
  const [focused, setFocused] = useState(false);
  // The suggestion the arrow keys picked, -1 for what was typed.
  const [highlight, setHighlight] = useState(-1);
  const templates = useTemplates();
  const full = value.length >= MaxBundleTags;

  const typed = normalizeTag(text);
  const used = useMemo(() => tagCounts(templates ?? []), [templates]);
  // The ones that start with what was typed first, then the ones that have
  // it. The most used without typing.
  const suggestions = used
    .filter(({ tag }) => !value.includes(tag) && tag !== typed)
    .filter(({ tag }) => !typed || tag.includes(typed))
    .sort(
      (a, b) =>
        Number(b.tag.startsWith(typed)) - Number(a.tag.startsWith(typed))
    )
    .slice(0, maxTagSuggestions);
  const open = focused && !full && !!suggestions.length;

  const addTag = (tag: string) => {
    setText("");
    setHighlight(-1);
    if (tag.length < 2 || value.includes(tag) || full) return;
    onChange([...value, tag]);
  };

  return (
    <div className="space-y-2">
      {!!value.length && (
        <div className="flex flex-wrap gap-1.5">
          {value.map((tag) => (
            <span
              key={tag}
              className="inline-flex items-center gap-1 rounded-full border bg-muted px-2.5 py-0.5 text-xs"
            >
              #{tag}
              <button
                type="button"
                aria-label={`Remove ${tag}`}
                className="text-muted-foreground hover:text-foreground"
                onClick={() => onChange(value.filter((t) => t !== tag))}
              >
                <XIcon className="h-3 w-3" />
              </button>
            </span>
          ))}
        </div>
      )}
      <div className="relative">
        <Input
          id={id}
          value={text}
          disabled={full}
          autoComplete="off"
          role="combobox"
          aria-expanded={open}
          aria-controls={`${id}-suggestions`}
          placeholder={
            full
              ? `At most ${MaxBundleTags} tags`
              : "Like roblox or economy, press Enter to add"
          }
          onChange={(e) => {
            setHighlight(-1);
            // A comma adds what's before it, like when pasting a list.
            if (e.target.value.includes(",")) {
              const parts = e.target.value.split(",");
              const last = parts.pop() ?? "";
              const added = parts
                .map(normalizeTag)
                .filter((t, i, all) => t.length >= 2 && all.indexOf(t) === i)
                .filter((t) => !value.includes(t));
              onChange([...value, ...added].slice(0, MaxBundleTags));
              setText(last);
            } else {
              setText(e.target.value);
            }
          }}
          onKeyDown={(e) => {
            if (e.key === "ArrowDown" && open) {
              e.preventDefault();
              setHighlight((h) => Math.min(h + 1, suggestions.length - 1));
            } else if (e.key === "ArrowUp" && open) {
              e.preventDefault();
              setHighlight((h) => Math.max(h - 1, -1));
            } else if (e.key === "Enter") {
              e.preventDefault();
              addTag(
                open && highlight >= 0 ? suggestions[highlight]!.tag : typed
              );
            } else if (e.key === "Escape" && open) {
              // Closes the suggestions, not the dialog.
              e.preventDefault();
              e.stopPropagation();
              setFocused(false);
            } else if (e.key === "Backspace" && !text && value.length) {
              onChange(value.slice(0, -1));
            }
          }}
          onFocus={() => setFocused(true)}
          onBlur={() => {
            setFocused(false);
            addTag(typed);
          }}
        />
        {open && (
          <div
            id={`${id}-suggestions`}
            role="listbox"
            className="absolute z-50 mt-1 w-full rounded-md border bg-popover p-1 shadow-md"
          >
            {!typed && (
              <div className="px-2 py-1 text-xs text-muted-foreground">
                Used by other templates
              </div>
            )}
            {suggestions.map(({ tag, count }, i) => (
              <button
                key={tag}
                type="button"
                role="option"
                aria-selected={i === highlight}
                // Before the input loses focus, which would add what was
                // typed instead.
                onMouseDown={(e) => {
                  e.preventDefault();
                  addTag(tag);
                }}
                onMouseEnter={() => setHighlight(i)}
                className={cn(
                  "flex w-full items-center justify-between rounded-sm px-2 py-1.5 text-sm",
                  i === highlight && "bg-accent text-accent-foreground"
                )}
              >
                <span>#{tag}</span>
                <span className="text-xs text-muted-foreground">
                  {count} {count === 1 ? "template" : "templates"}
                </span>
              </button>
            ))}
          </div>
        )}
      </div>
      <p className="text-sm text-muted-foreground">
        {`Up to ${MaxBundleTags} words on what it does, so people find it. Pick one others use when it fits.`}
      </p>
    </div>
  );
}

// How many tags TagInput suggests at once.
const maxTagSuggestions = 6;
