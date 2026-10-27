import { ReactNode, useEffect, useState } from "react";
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
import { useBundleListingSubmitMutation } from "@/lib/api/mutations";
import {
  iconName,
  lucideName,
  templateCategories,
  templateIcons,
} from "@/lib/flow/templates";
import TemplateIcon, { loadLucide } from "./TemplateIcon";
import { UserBundle } from "@/lib/types/wire.gen";

// BundleListingDialog submits a shared bundle to the marketplace, with how
// it's shown there. Kite reviews it before it's listed.
export default function BundleListingDialog({
  bundle,
  children,
}: {
  bundle: UserBundle;
  children: ReactNode;
}) {
  const [open, setOpen] = useState(false);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="max-w-2xl max-h-[90dvh] overflow-y-auto">
        {open && (
          <BundleListingForm bundle={bundle} onDone={() => setOpen(false)} />
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

function BundleListingForm({
  bundle,
  onDone,
}: {
  bundle: UserBundle;
  onDone: () => void;
}) {
  const submitMutation = useBundleListingSubmitMutation();
  // The last submission is the starting point, also when it was rejected.
  const last = bundle.latest_review?.listing;

  const [icon, setIcon] = useState(last?.icon ?? "");
  const [category, setCategory] = useState(last?.category ?? "");
  const [about, setAbout] = useState(last?.about.join("\n\n") ?? "");
  const [setup, setSetup] = useState(last?.setup.join("\n") ?? "");

  const paragraphs = splitParagraphs(about);
  const tooLong = paragraphs.some((p) => p.length > maxParagraph);
  // One step per line.
  const steps = setup
    .split("\n")
    .map((s) => s.trim())
    .filter(Boolean);
  const stepTooLong = steps.some((s) => s.length > maxStep);

  async function submit() {
    const res = await submitMutation.mutateAsync({
      bundleId: bundle.id,
      listing: {
        icon,
        category,
        about: paragraphs,
        faq: [],
        setup: steps,
      },
    });
    if (!res.success) {
      toast.error(
        `Failed to submit ${bundle.name}: ${res.error.message} (${res.error.code})`
      );
      return;
    }
    toast.success(`Submitted ${bundle.name}, Kite reviews it soon`);
    onDone();
  }

  return (
    <>
      <DialogHeader>
        <DialogTitle>List {bundle.name} in the marketplace</DialogTitle>
        <DialogDescription>
          Kite reviews version {bundle.version} and this page before it&apos;s
          listed. Once it is, its link leads to the marketplace, and new
          versions are reviewed before apps get them.
        </DialogDescription>
      </DialogHeader>

      <div className="space-y-5">
        <div className="space-y-1.5">
          <Label>Icon</Label>
          <IconPicker value={icon} onChange={setIcon} />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="listing-category">Category</Label>
          <Select value={category} onValueChange={setCategory}>
            <SelectTrigger id="listing-category">
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
          <Label htmlFor="listing-about">How it works</Label>
          <Textarea
            id="listing-about"
            value={about}
            rows={6}
            placeholder="What it does and how members use it. Leave an empty line between paragraphs."
            onChange={(e) => setAbout(e.target.value)}
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
          <Label htmlFor="listing-setup">Before you start</Label>
          <Textarea
            id="listing-setup"
            value={setup}
            rows={2}
            placeholder="Steps outside of Kite it needs, like turning on an intent. One per line, optional."
            onChange={(e) => setSetup(e.target.value)}
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
          onClick={submit}
          loading={submitMutation.isPending}
          disabled={
            !icon ||
            !category ||
            !paragraphs.length ||
            paragraphs.length > 5 ||
            tooLong ||
            steps.length > maxSteps ||
            stepTooLong
          }
        >
          Submit for review
        </LoadingButton>
      </DialogFooter>
    </>
  );
}

// pickerLimit is how many icons the picker shows at once. Searching finds
// the others.
const pickerLimit = 96;

// IconPicker picks any Lucide icon, found by its name. Kite's template icons
// come first.
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
      const ownComponents = new Set(own.map(lucideName));
      const others = Object.keys(icons)
        .filter((c) => !ownComponents.has(c))
        .map(iconName);
      setNames([...own, ...others]);
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
