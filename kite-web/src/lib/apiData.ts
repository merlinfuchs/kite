import { apiRequest } from "./api/client";
import { isExport } from "./env/client";
import { BillingPlanListResponse, StatsGetResponse } from "./types/wire.gen";

// The totals of the homepage and the installs of each template, or null when
// the page was built without them.
export type Stats = StatsGetResponse | null;
// The plans of the pricing section, or null when the page was built without
// them.
export type Plans = BillingPlanListResponse | null;

type ApiData = { stats: Stats; plans: Plans };

const paths: Record<keyof ApiData, string> = {
  stats: "/v1/stats",
  plans: "/v1/billing/plans",
};

// Adds data from the API to the props of getStaticProps. On kite.onl the page
// is rebuilt every hour, with stats the API caches for up to an hour. The
// export for self-hosting has no server to rebuild it, and the data of
// kite.onl doesn't belong there.
export async function withApiData<P, K extends keyof ApiData>(
  props: P,
  keys: K[]
) {
  const data = Object.fromEntries(
    await Promise.all(
      keys.map(async (key) => [key, isExport ? null : await fetchData(key)])
    )
  ) as Pick<ApiData, K>;
  return {
    props: { ...props, ...data },
    ...(isExport ? {} : { revalidate: 60 * 60 }),
  };
}

async function fetchData(key: keyof ApiData) {
  try {
    const res = await apiRequest<NonNullable<ApiData[typeof key]>>(paths[key], {
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
      console.warn(`Failed to get the ${key}: ${err}`);
      return null;
    }
    throw err;
  }
}
