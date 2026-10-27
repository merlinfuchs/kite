import { UserBundle } from "../types/wire.gen";

// inMarketplace is whether a bundle is listed or in review for it.
export function inMarketplace(bundle: UserBundle) {
  return (
    bundle.listed_version !== null || bundle.latest_review?.status === "pending"
  );
}

// newVersionsReviewed is whether a new version of a bundle goes to the
// marketplace review, like the service decides when it's published.
export function newVersionsReviewed(bundle: UserBundle) {
  return (
    bundle.removed_reason === null &&
    (inMarketplace(bundle) || bundle.latest_review?.status === "rejected")
  );
}
