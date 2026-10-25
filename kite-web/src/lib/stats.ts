import { apiRequest } from "./api/client";
import { StatsGetResponse } from "./types/wire.gen";

// The totals of the homepage and the installs of each template, or null when
// the page was built without them.
export type Stats = StatsGetResponse | null;

// Adds the stats to the props of getStaticProps. On kite.onl the page is
// rebuilt every hour, with stats the API caches for up to an hour. The export
// for self-hosting has no server to rebuild it, and the stats of kite.onl
// don't belong there.
export async function withStats<P>(props: P) {
  if (process.env.OUTPUT === "export") {
    return { props: { ...props, stats: null as Stats } };
  }
  return {
    props: { ...props, stats: await fetchStats() },
    revalidate: 60 * 60,
  };
}

async function fetchStats(): Promise<Stats> {
  try {
    const res = await apiRequest<StatsGetResponse>("/v1/stats", {
      signal: AbortSignal.timeout(5000),
    });
    if (!res.success) {
      throw new Error(res.error.message);
    }
    return res.data;
  } catch (err) {
    // Builds, like the ones of CI, and next dev don't need the API. When the
    // server rebuilds a page, throwing keeps the previous one with its stats.
    if (
      process.env.NODE_ENV !== "production" ||
      process.env.NEXT_PHASE === "phase-production-build"
    ) {
      console.warn(`Failed to get the stats: ${err}`);
      return null;
    }
    throw err;
  }
}
