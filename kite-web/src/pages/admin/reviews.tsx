import { TemplateDetails } from "@/components/app/TemplateDetails";
import BaseLayout from "@/components/common/BaseLayout";
import LoadingButton from "@/components/common/LoadingButton";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Textarea } from "@/components/ui/textarea";
import { useBundleReviewDecideMutation } from "@/lib/api/mutations";
import { usePendingBundleReviewsQuery } from "@/lib/api/queries";
import { templateFromBundle } from "@/lib/flow/templates";
import { useUser } from "@/lib/hooks/api";
import { PendingBundleReview } from "@/lib/types/wire.gen";
import { useState } from "react";
import { toast } from "sonner";

// AdminReviewsPage is where admins approve users' bundles for the marketplace,
// or reject them with a reason for their owner.
export default function AdminReviewsPage() {
  const user = useUser();
  const { data } = usePendingBundleReviewsQuery(!!user?.is_admin);
  const reviews = data?.success
    ? data.data.flatMap((r) => (r ? [r] : []))
    : undefined;

  return (
    <BaseLayout title="Reviews" noindex>
      <div className="container max-w-5xl py-12">
        <h1 className="text-lg font-semibold md:text-2xl mb-1">Reviews</h1>
        <p className="text-muted-foreground text-sm">
          Bundles waiting to be listed in the marketplace, and new versions of
          listed ones, oldest first.
        </p>
        <Separator className="my-8" />
        {user && !user.is_admin && (
          <p className="text-sm text-muted-foreground">
            You don&apos;t have access to this page.
          </p>
        )}
        {reviews && !reviews.length && (
          <p className="text-sm text-muted-foreground">Nothing to review.</p>
        )}
        <div className="space-y-16">
          {reviews?.map((r) => (
            <Review key={r.id} review={r} />
          ))}
        </div>
      </div>
    </BaseLayout>
  );
}

function Review({ review }: { review: PendingBundleReview }) {
  const decideMutation = useBundleReviewDecideMutation();
  const [reason, setReason] = useState("");

  const template = review.bundle && templateFromBundle(review.bundle);

  async function decide(approve: boolean) {
    const res = await decideMutation.mutateAsync({
      reviewId: review.id,
      req: { approve, reason: reason.trim() },
    });
    if (!res.success) {
      toast.error(`Failed to decide: ${res.error.message} (${res.error.code})`);
      return;
    }
    toast.success(
      `${approve ? "Approved" : "Rejected"} ${template?.name ?? "the bundle"}`
    );
  }

  return (
    <div>
      <div className="flex flex-wrap items-center gap-2 mb-4 text-sm text-muted-foreground">
        <Badge variant="outline">
          {review.listed_version !== null
            ? `Version ${review.version}, listed is ${review.listed_version}`
            : `New listing, version ${review.version}`}
        </Badge>
        <span>
          Submitted {new Date(review.created_at).toLocaleString()} by{" "}
          {review.bundle?.owner?.display_name} (@
          {review.bundle?.owner?.discord_username})
        </span>
      </div>
      {template ? (
        <TemplateDetails
          template={template}
          aside={
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
          }
        />
      ) : (
        <p className="text-sm text-destructive">
          The content of this version can&apos;t be shown.
        </p>
      )}
    </div>
  );
}
