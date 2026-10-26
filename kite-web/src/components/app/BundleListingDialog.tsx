import { ReactNode, useState } from "react";
import { toast } from "sonner";
import { PlusIcon, Trash2Icon } from "lucide-react";
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
import { templateCategories, templateIcons } from "@/lib/flow/templates";
import { BundleFAQ, UserBundle } from "@/lib/types/wire.gen";

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

// maxParagraph is the most characters the service allows in a paragraph.
const maxParagraph = 1500;

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
  const [faq, setFaq] = useState<BundleFAQ[]>(last?.faq ?? []);

  const paragraphs = splitParagraphs(about);
  const tooLong = paragraphs.some((p) => p.length > maxParagraph);
  const filledFaq = faq.filter((f) => f.question.trim() || f.answer.trim());

  async function submit() {
    const res = await submitMutation.mutateAsync({
      bundleId: bundle.id,
      listing: {
        icon,
        category,
        about: paragraphs,
        faq: filledFaq.map((f) => ({
          question: f.question.trim(),
          answer: f.answer.trim(),
        })),
        setup: setup
          .split("\n")
          .map((s) => s.trim())
          .filter(Boolean),
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

  function setFaqEntry(i: number, entry: Partial<BundleFAQ>) {
    setFaq(faq.map((f, j) => (j === i ? { ...f, ...entry } : f)));
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
          <ToggleGroup
            type="single"
            variant="outline"
            value={icon}
            onValueChange={(v) => v && setIcon(v)}
            className="flex-wrap justify-start"
          >
            {Object.entries(templateIcons).map(([name, Icon]) => (
              <ToggleGroupItem key={name} value={name} aria-label={name}>
                <Icon className="h-4 w-4" />
              </ToggleGroupItem>
            ))}
          </ToggleGroup>
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
        </div>
        <div className="space-y-2">
          <Label>Questions and answers</Label>
          {faq.map((f, i) => (
            <div key={i} className="flex gap-2 items-start">
              <div className="flex-1 space-y-1.5">
                <Input
                  value={f.question}
                  maxLength={200}
                  placeholder="Question"
                  onChange={(e) => setFaqEntry(i, { question: e.target.value })}
                />
                <Textarea
                  value={f.answer}
                  maxLength={1500}
                  rows={2}
                  placeholder="Answer"
                  onChange={(e) => setFaqEntry(i, { answer: e.target.value })}
                />
              </div>
              <Button
                variant="ghost"
                size="icon"
                aria-label="Remove question"
                onClick={() => setFaq(faq.filter((_, j) => j !== i))}
              >
                <Trash2Icon className="h-4 w-4" />
              </Button>
            </div>
          ))}
          {faq.length < 10 && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => setFaq([...faq, { question: "", answer: "" }])}
            >
              <PlusIcon className="mr-2 h-4 w-4" />
              Add question
            </Button>
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
            filledFaq.some((f) => !f.question.trim() || !f.answer.trim())
          }
        >
          Submit for review
        </LoadingButton>
      </DialogFooter>
    </>
  );
}
