import { apiRequest } from "@/lib/api/client";
import { templateFromBundle } from "@/lib/flow/templates";
import { handleTemplateImage } from "@/lib/og/handler";
import { BundleGetResponse } from "@/lib/types/wire.gen";
import type { NextApiRequest, NextApiResponse } from "next";

export default function handler(req: NextApiRequest, res: NextApiResponse) {
  return handleTemplateImage(req, res, async () => {
    const bundleId = encodeURIComponent(String(req.query.bundleId));
    const bundle = await apiRequest<BundleGetResponse>(
      `/v1/bundles/${bundleId}`,
      { signal: AbortSignal.timeout(5000) }
    );
    return bundle.success ? templateFromBundle(bundle.data) : undefined;
  });
}
