import { apiRequest } from "./api/client";
import { isExport } from "./env/client";
import {
  BillingPlanListResponse,
  Bundle,
  BundleListResponse,
  StatsGetResponse,
} from "./types/wire.gen";

// The totals of the homepage and the installs of each template, or null when
// the page was built without them.
export type Stats = StatsGetResponse | null;
// The plans of the pricing section, or null when the page was built without
// them.
export type Plans = BillingPlanListResponse | null;

// The result of getStaticProps for pages with data from the API. On kite.onl
// the page is rebuilt every hour by default, like the API caches the stats.
// Plans rarely change.
// The export for self-hosting has no server to rebuild it, and the data of
// kite.onl doesn't belong there.
export function withApiData<P>(props: P, revalidate = 60 * 60) {
  return { props, ...(isExport ? {} : { revalidate }) };
}

export function fetchStats(): Promise<Stats> {
  return fetchApiData("/v1/stats");
}

export function fetchPlans(): Promise<Plans> {
  return fetchApiData("/v1/billing/plans");
}

// fetchListedBundles are the users' bundles listed in the marketplace, with
// only what their cards show, or null when the page is built without them.
// The official templates come from the files. The API caches the list, so
// pages of creators that don't exist can load it too.
export async function fetchListedBundles(): Promise<Bundle[] | null> {
  const bundles = await fetchApiData<BundleListResponse>("/v1/bundles");
  return bundles?.flatMap((b) => (b && !b.official ? [b] : [])) ?? null;
}

async function fetchApiData<T>(path: string): Promise<T | null> {
  if (isExport) return null;
  try {
    const res = await apiRequest<T>(path, {
      signal: AbortSignal.timeout(5000),
    });
    if (!res.success) {
      throw new Error(res.error.message);
    }
    return res.data;
  } catch (err) {
    // Builds, like the ones of CI, and next dev don't need the API. When the
    // server rebuilds a page, throwing keeps the previous one with its data.
    if (
      process.env.NODE_ENV !== "production" ||
      process.env.NEXT_PHASE === "phase-production-build"
    ) {
      console.warn(`Failed to get ${path}: ${err}`);
      return null;
    }
    throw err;
  }
}
