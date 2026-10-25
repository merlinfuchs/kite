import env from "@/lib/env/client";
import { APIResponse } from "./api/response";
import { StatsGetResponse } from "./types/wire.gen";

// The number of apps that added each template, by its ID.
export type TemplateInstalls = Record<string, number> | null;

// Fewer installs look like nobody uses a template, so they aren't shown.
const minShownInstalls = 10;

export function shownInstalls(installs: TemplateInstalls, templateId: string) {
  const count = installs?.[templateId] ?? 0;
  return count >= minShownInstalls ? count : null;
}

// Adds the install counts to the props of getStaticProps. On kite.onl the page
// is rebuilt with new counts every hour. The export for self-hosting has no
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
    const res = await fetch(`${env.NEXT_PUBLIC_API_PUBLIC_BASE_URL}/v1/stats`, {
      signal: AbortSignal.timeout(5000),
    });
    if (!res.ok) {
      throw new Error(`The API responded with ${res.status}`);
    }
    const body = (await res.json()) as APIResponse<StatsGetResponse>;
    if (!body.success) {
      throw new Error(body.error.message);
    }
    return body.data.template_installs ?? null;
  } catch (err) {
    // Builds, like the ones of CI, don't need the API. When rebuilding a page
    // later, throwing keeps the previous one with its counts.
    if (process.env.NEXT_PHASE === "phase-production-build") {
      console.warn(`Failed to get the template installs: ${err}`);
      return null;
    }
    throw err;
  }
}
