import { apiRequest } from "./api/client";
import { StatsGetResponse } from "./types/wire.gen";

// The number of apps that added each template, by its ID.
export type TemplateInstalls = Record<string, number> | null;

// Adds the install counts to the props of getStaticProps. On kite.onl the page
// is rebuilt every hour, with counts the API caches for up to an hour. The export for self-hosting has no
// server to rebuild it, and the counts of kite.onl don't belong there.
export async function withTemplateInstalls<P>(props: P) {
  if (process.env.OUTPUT === "export") {
    return { props: { ...props, installs: null as TemplateInstalls } };
  }
  return {
    props: { ...props, installs: await fetchTemplateInstalls() },
    revalidate: 60 * 60,
  };
}

async function fetchTemplateInstalls(): Promise<TemplateInstalls> {
  try {
    const res = await apiRequest<StatsGetResponse>("/v1/stats", {
      signal: AbortSignal.timeout(5000),
    });
    if (!res.success) {
      throw new Error(res.error.message);
    }
    return res.data.template_installs ?? null;
  } catch (err) {
    // Builds, like the ones of CI, and next dev don't need the API. When the
    // server rebuilds a page, throwing keeps the previous one with its counts.
    if (
      process.env.NODE_ENV !== "production" ||
      process.env.NEXT_PHASE === "phase-production-build"
    ) {
      console.warn(`Failed to get the template installs: ${err}`);
      return null;
    }
    throw err;
  }
}
