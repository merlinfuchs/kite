import AdminPage from "@/components/app/AdminPage";
import { TemplateDetails } from "@/components/app/TemplateDetails";
import LoadingButton from "@/components/common/LoadingButton";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import {
  useBundleReviewDecideMutation,
  useBundleReviewPrescreenRunMutation,
  useBundleReviewReopenMutation,
  useUserBanCreateMutation,
} from "@/lib/api/mutations";
import { usePendingBundleReviewQuery } from "@/lib/api/queries";
import { TemplateParameter, templateFromBundle } from "@/lib/flow/templates";
import {
  useAutoRejectedBundleReviews,
  usePendingBundleReviews,
} from "@/lib/hooks/api";
import { cn } from "@/lib/utils";
import {
  BundleReviewPrescreen,
  PendingBundleReview,
} from "@/lib/types/wire.gen";
import { useMemo, useState } from "react";
import { toast } from "sonner";

// AdminReviewsPage is where admins approve users' bundles for the marketplace,
// or reject them with a reason for their owner.
export default function AdminReviewsPage() {
  return (
    <AdminPage
      title="Reviews"
      description="Templates waiting to be listed in the marketplace, and new versions of listed ones, oldest first."
    >
      <div className="space-y-16">
        <ReviewList />
        <AutoRejectedList />
      </div>
    </AdminPage>
  );
}

// ReviewList lists the reviews without their content and shows the one that's
// open, the oldest until another is clicked. A decided one leaves the list,
// which opens the next.
function ReviewList() {
  const reviews = usePendingBundleReviews();
  const [openId, setOpenId] = useState<string>();
  const [byRating, setByRating] = useState(false);
  const pending = useMemo(() => {
    const list = reviews?.flatMap((r) => (r ? [r] : [])) ?? [];
    // Unrated ones last, the order of the queue otherwise.
    return byRating
      ? [...list].sort(
          (a, b) => (b.prescreen.rating ?? -1) - (a.prescreen.rating ?? -1)
        )
      : list;
  }, [reviews, byRating]);
  const open = pending.find((r) => r.id === openId) ?? pending[0];

  if (reviews && !pending.length) {
    return <p className="text-sm text-muted-foreground">Nothing to review.</p>;
  }
  return (
    <div className="space-y-12">
      <div className="space-y-2">
        <Tabs
          value={byRating ? "rating" : "oldest"}
          onValueChange={(v) => setByRating(v === "rating")}
          className="flex justify-end"
        >
          <TabsList>
            <TabsTrigger value="oldest">Oldest first</TabsTrigger>
            <TabsTrigger value="rating">Highest rating</TabsTrigger>
          </TabsList>
        </Tabs>
        {pending.map((r) => (
          <button
            key={r.id}
            type="button"
            onClick={() => setOpenId(r.id)}
            className={cn(
              "w-full text-left rounded-lg border p-3 flex flex-wrap items-center gap-x-3 gap-y-1 text-sm hover:bg-muted/50",
              r.id === open?.id && "border-primary bg-muted/50"
            )}
          >
            <PrescreenBadge prescreen={r.prescreen} />
            <span className="font-medium">{r.bundle?.name}</span>
            <ReviewBadge review={r} />
            <span className="text-muted-foreground">
              {new Date(r.created_at).toLocaleString()} by{" "}
              {r.bundle?.owner?.display_name} (@
              {r.bundle?.owner?.discord_username})
            </span>
          </button>
        ))}
      </div>
      {open && <Review key={open.id} id={open.id} />}
    </div>
  );
}

// PrescreenBadge is the rating of the AI pre-screen, colored by how likely
// it's approved.
function PrescreenBadge({ prescreen }: { prescreen: BundleReviewPrescreen }) {
  if (!prescreen.screened_at) {
    return <Badge variant="outline">Screening…</Badge>;
  }
  if (prescreen.rating === null) {
    return (
      <Badge variant="outline">
        {prescreen.verdict === "skipped" ? "Too big" : "Not rated"}
      </Badge>
    );
  }
  const rating = prescreen.rating;
  return (
    <Badge
      variant="outline"
      className={cn(
        "border-transparent font-medium tabular-nums",
        rating >= 80
          ? "bg-green-500/15 text-green-600 dark:text-green-500"
          : rating >= 50
          ? "bg-amber-500/15 text-amber-600 dark:text-amber-500"
          : "bg-destructive/15 text-destructive"
      )}
    >
      {rating}
    </Badge>
  );
}

const severityClasses: Record<string, string> = {
  high: "bg-destructive",
  medium: "bg-amber-500",
  low: "bg-muted-foreground",
};

// PrescreenCard is what the AI pre-screen made of the review: a summary, what
// to look at and a rating. It can run again, like after it failed.
function PrescreenCard({ review }: { review: PendingBundleReview }) {
  const runMutation = useBundleReviewPrescreenRunMutation();
  const { prescreen } = review;

  async function run() {
    const res = await runMutation.mutateAsync(review.id);
    if (!res.success) {
      toast.error(`Failed: ${res.error.message} (${res.error.code})`);
    }
  }

  return (
    <Card className="p-5 space-y-3">
      <div className="flex items-center justify-between gap-3">
        <div className="font-semibold">AI pre-screen</div>
        <PrescreenBadge prescreen={prescreen} />
      </div>
      {prescreen.screened_at ? (
        <>
          {prescreen.summary && (
            <p className="text-sm text-muted-foreground">{prescreen.summary}</p>
          )}
          {!!prescreen.concerns.length && (
            <ul className="space-y-2">
              {prescreen.concerns.map((c, i) => (
                <li key={i} className="flex gap-2 text-sm">
                  <span
                    className={cn(
                      "mt-1.5 h-2 w-2 flex-none rounded-full",
                      severityClasses[c.severity] ?? severityClasses.low
                    )}
                    title={c.severity}
                  />
                  {c.text}
                </li>
              ))}
            </ul>
          )}
          <div className="flex items-center justify-between gap-3 text-xs text-muted-foreground">
            <span>
              {prescreen.model},{" "}
              {new Date(prescreen.screened_at).toLocaleString()}
            </span>
            <LoadingButton
              variant="ghost"
              size="sm"
              loading={runMutation.isPending}
              onClick={run}
            >
              Run again
            </LoadingButton>
          </div>
        </>
      ) : (
        <p className="text-sm text-muted-foreground">
          It runs within a minute of submitting.
        </p>
      )}
    </Card>
  );
}

// AutoRejectedList lists what the AI pre-screen rejected, to reopen what it
// got wrong.
function AutoRejectedList() {
  const reviews = useAutoRejectedBundleReviews();
  const reopenMutation = useBundleReviewReopenMutation();
  const rejected = reviews?.flatMap((r) => (r ? [r] : [])) ?? [];

  async function reopen(id: string) {
    const res = await reopenMutation.mutateAsync(id);
    if (!res.success) {
      toast.error(`Failed to reopen: ${res.error.message} (${res.error.code})`);
      return;
    }
    toast.success("Back in the queue");
  }

  if (!rejected.length) return null;
  return (
    <section className="space-y-3">
      <div>
        <h2 className="text-lg font-semibold">Rejected by the pre-screen</h2>
        <p className="text-sm text-muted-foreground">
          Newest first. Reopen one to decide it yourself.
        </p>
      </div>
      {rejected.map((r) => (
        <Card key={r.id} className="p-4 space-y-2">
          <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-sm">
            <PrescreenBadge prescreen={r.prescreen} />
            <span className="font-medium">{r.bundle?.name}</span>
            <span className="text-muted-foreground">
              {r.decided_at && new Date(r.decided_at).toLocaleString()} by{" "}
              {r.bundle?.owner?.display_name} (@
              {r.bundle?.owner?.discord_username})
            </span>
            <LoadingButton
              variant="outline"
              size="sm"
              className="ml-auto"
              loading={
                reopenMutation.isPending && reopenMutation.variables === r.id
              }
              disabled={reopenMutation.isPending}
              onClick={() => reopen(r.id)}
            >
              Reopen
            </LoadingButton>
          </div>
          <p className="text-sm">{r.reason}</p>
          {r.prescreen.summary && (
            <p className="text-sm text-muted-foreground">
              {r.prescreen.summary}
            </p>
          )}
        </Card>
      ))}
    </section>
  );
}

function ReviewBadge({ review }: { review: PendingBundleReview }) {
  return (
    <Badge variant="outline">
      {review.listed_version === review.version
        ? `Page changes, version ${review.version}`
        : review.listed_version !== null
        ? `Version ${review.version}, listed is ${review.listed_version}`
        : review.approved_version !== null
        ? `Listing again, version ${review.version}, apps have ${review.approved_version}`
        : `New listing, version ${review.version}`}
    </Badge>
  );
}

function Review({ id }: { id: string }) {
  const { data } = usePendingBundleReviewQuery(id);
  const decideMutation = useBundleReviewDecideMutation();
  const banMutation = useUserBanCreateMutation();
  const [reason, setReason] = useState("");
  const deciding = decideMutation.isPending || banMutation.isPending;

  const review = data?.success ? data.data : undefined;
  const template = useMemo(
    () => (review?.bundle ? templateFromBundle(review.bundle) : undefined),
    [review]
  );

  async function decide(approve: boolean) {
    const res = await decideMutation.mutateAsync({
      reviewId: id,
      req: { approve, reason: reason.trim() },
    });
    if (!res.success) {
      toast.error(`Failed to decide: ${res.error.message} (${res.error.code})`);
      return;
    }
    toast.success(
      `${approve ? "Approved" : "Rejected"} ${template?.name ?? "the template"}`
    );
  }

  // Banning the owner rejects this and the rest of what they have in review.
  async function rejectAndBan() {
    if (!review) return;
    const res = await banMutation.mutateAsync({
      user_id: review.owner_user_id,
      feature: "marketplace",
      reason: reason.trim(),
    });
    if (!res.success) {
      toast.error(`Failed to ban: ${res.error.message} (${res.error.code})`);
      return;
    }
    toast.success(
      `Rejected ${
        template?.name ?? "the template"
      } and banned its owner from the marketplace`
    );
  }

  if (data && !data.success) {
    return (
      <p className="text-sm text-destructive">
        {data.error.message} ({data.error.code})
      </p>
    );
  }
  if (!review) return null;
  if (!template) {
    return (
      <p className="text-sm text-destructive">
        The content of this version can&apos;t be shown.
      </p>
    );
  }
  const notes = template.versionNotes.find(
    (n) => n.version === template.version
  );
  return (
    <div className="space-y-12">
      <TemplateDetails
        template={template}
        starrable={false}
        aside={
          <div className="space-y-4">
            <PrescreenCard review={review} />
            {notes && (
              <Card className="p-5 space-y-2">
                <div className="font-semibold">
                  What&apos;s new in version {notes.version}
                </div>
                <p className="text-sm text-muted-foreground whitespace-pre-wrap break-words">
                  {notes.notes}
                </p>
              </Card>
            )}
            <Card className="p-5 space-y-3">
              <div className="font-semibold">Decision</div>
              <Textarea
                value={reason}
                maxLength={1000}
                rows={4}
                placeholder="Why it's rejected, its owner sees this"
                onChange={(e) => setReason(e.target.value)}
              />
              <div className="flex gap-2">
                <LoadingButton
                  className="flex-1"
                  loading={deciding}
                  onClick={() => decide(true)}
                >
                  Approve
                </LoadingButton>
                <LoadingButton
                  variant="outline"
                  className="flex-1"
                  loading={deciding}
                  disabled={!reason.trim()}
                  onClick={() => decide(false)}
                >
                  Reject
                </LoadingButton>
              </div>
              <LoadingButton
                variant="destructive"
                className="w-full"
                loading={deciding}
                disabled={!reason.trim()}
                onClick={rejectAndBan}
              >
                Reject and ban owner
              </LoadingButton>
            </Card>
          </div>
        }
      />
      <ParametersSection parameters={template.parameters} />
    </div>
  );
}

const parameterTypeLabels: Record<TemplateParameter["type"], string> = {
  text: "Text",
  textarea: "Long text",
  channel: "Channel",
  role: "Role",
  boolean: "Yes/No",
};

// ParametersSection shows what apps are asked when adding the template, and
// the values it shares for them, which apps start with. They aren't on its
// page, but they're part of what's reviewed.
function ParametersSection({
  parameters,
}: {
  parameters: TemplateParameter[];
}) {
  if (!parameters.length) return null;

  return (
    <section className="max-w-2xl">
      <h2 className="font-semibold mb-3">Parameters</h2>
      <div className="space-y-3">
        {parameters.map((p) => (
          <Card key={p.key} className="p-4 space-y-1">
            <div className="flex flex-wrap items-center gap-2 text-sm">
              <span className="font-medium">{p.label || p.key}</span>
              <Badge variant="outline">{parameterTypeLabels[p.type]}</Badge>
              {p.required && <Badge variant="outline">Required</Badge>}
            </div>
            {p.description && (
              <p className="text-sm text-muted-foreground">{p.description}</p>
            )}
            {p.default && (
              <div className="text-xs text-muted-foreground whitespace-pre-wrap break-words">
                <span className="font-medium text-foreground">Default:</span>{" "}
                {p.default}
              </div>
            )}
          </Card>
        ))}
      </div>
    </section>
  );
}
