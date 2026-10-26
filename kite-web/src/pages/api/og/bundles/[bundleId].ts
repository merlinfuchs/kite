import { fetchBundle } from "@/lib/api/client";
import { templateFromBundle } from "@/lib/flow/templates";
import { handleTemplateImage } from "@/lib/og/handler";
import type { NextApiRequest, NextApiResponse } from "next";

export default function handler(req: NextApiRequest, res: NextApiResponse) {
  const bundleId = String(req.query.bundleId);
  return handleTemplateImage(req, res, "bundles", bundleId, async () => {
    const bundle = await fetchBundle(bundleId);
    return bundle ? templateFromBundle(bundle) : undefined;
  });
}
