import { UserBundle } from "../types/wire.gen";

// The most characters the service allows in the name and description of a
// bundle. The description is one sentence for the template cards.
export const maxBundleName = 100;
export const maxBundleDescription = 200;

// nameHasSymbols is whether a name has emojis or other symbols, which the
// service rejects. The marketplace shows the icon next to the name instead.
export function nameHasSymbols(name: string) {
  return /\p{So}/u.test(name);
}

// inMarketplace is whether a bundle is listed or in review for it.
export function inMarketplace(bundle: UserBundle) {
  return (
    bundle.listed_version !== null || bundle.latest_review?.status === "pending"
  );
}

// newVersionsReviewed is whether a new version of a bundle goes to the
// marketplace review, like the service decides when it's published. A
// rejected one is only submitted again with a complete page.
export function newVersionsReviewed(bundle: UserBundle) {
  const { page } = bundle;
  return (
    bundle.removed_reason === null &&
    (inMarketplace(bundle) ||
      (bundle.latest_review?.status === "rejected" &&
        !!page.icon &&
        !!page.category &&
        !!page.about.length))
  );
}
