import AdminPage from "@/components/app/AdminPage";
import { TemplateDetails } from "@/components/app/TemplateDetails";
import LoadingButton from "@/components/common/LoadingButton";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { Textarea } from "@/components/ui/textarea";
import { useBundleReviewDecideMutation } from "@/lib/api/mutations";
import { usePendingBundleReviewQuery } from "@/lib/api/queries";
import { TemplateParameter, templateFromBundle } from "@/lib/flow/templates";
import { usePendingBundleReviews } from "@/lib/hooks/api";
import { cn } from "@/lib/utils";
import { PendingBundleReview } from "@/lib/types/wire.gen";
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
      <ReviewList />
    </AdminPage>
  );
}

// ReviewList lists the reviews without their content and shows the one that's
// open, the oldest until another is clicked. A decided one leaves the list,
// which opens the next.
function ReviewList() {
  const reviews = usePendingBundleReviews();
  const [openId, setOpenId] = useState<string>();
  const pending = useMemo(
    () => reviews?.flatMap((r) => (r ? [r] : [])) ?? [],
    [reviews]
  );
  const open = pending.find((r) => r.id === openId) ?? pending[0];

  if (reviews && !pending.length) {
    return <p className="text-sm text-muted-foreground">Nothing to review.</p>;
  }
  return (
    <div className="space-y-12">
      <div className="space-y-2">
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

function ReviewBadge({ review }: { review: PendingBundleReview }) {
  return (
    <Badge variant="outline">
      {review.listed_version !== null
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
  const [reason, setReason] = useState("");

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
  return (
    <TemplateDetails
      template={template}
      aside={
        <div className="space-y-4">
          <ParameterDefaultsCard parameters={template.parameters} />
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
                loading={decideMutation.isPending}
                onClick={() => decide(true)}
              >
                Approve
              </LoadingButton>
              <LoadingButton
                variant="outline"
                className="flex-1"
                loading={decideMutation.isPending}
                disabled={!reason.trim()}
                onClick={() => decide(false)}
              >
                Reject
              </LoadingButton>
            </div>
          </Card>
        </div>
      }
    />
  );
}

// ParameterDefaultsCard shows the values the template shares for its
// parameters, which apps start with. They aren't on its page, but they're
// part of what's reviewed.
function ParameterDefaultsCard({
  parameters,
}: {
  parameters: TemplateParameter[];
}) {
  const withDefault = parameters.filter((p) => p.default);
  if (!withDefault.length) return null;

  return (
    <Card className="p-5 space-y-3">
      <div className="font-semibold">Default values</div>
      {withDefault.map((p) => (
        <div key={p.key} className="space-y-1">
          <div className="text-sm font-medium">{p.label || p.key}</div>
          <div className="text-xs text-muted-foreground whitespace-pre-wrap break-words">
            {p.default}
          </div>
        </div>
      ))}
    </Card>
  );
}
