import { fetchBundles, fetchPlans, fetchStats } from "@/lib/apiData";
import env from "@/lib/env/client";
import { llmsTxt } from "@/lib/llmsTxt";
import type { NextApiRequest, NextApiResponse } from "next";

// Served as /llms.txt. The numbers change, so it's generated like the
// marketplace sitemap.
export default async function handler(
  _req: NextApiRequest,
  res: NextApiResponse
) {
  const [stats, plans, bundles] = await Promise.all([
    fetchStats().catch(() => null),
    fetchPlans().catch(() => null),
    fetchBundles().catch(() => null),
  ]);
  res.setHeader("Content-Type", "text/plain; charset=utf-8");
  res.setHeader("Cache-Control", "public, max-age=3600");
  res.send(
    llmsTxt({
      stats,
      plans: plans?.flatMap((p) => (p ? [p] : [])) ?? null,
      bundles,
      docsUrl: env.NEXT_PUBLIC_DOCS_LINK,
      githubUrl: env.NEXT_PUBLIC_GITHUB_LINK,
      today: new Date().toISOString().slice(0, 10),
    })
  );
}
