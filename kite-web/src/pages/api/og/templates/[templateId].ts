import { getTemplate } from "@/lib/flow/templateFiles";
import { handleTemplateImage } from "@/lib/og/handler";
import type { NextApiRequest, NextApiResponse } from "next";

export default function handler(req: NextApiRequest, res: NextApiResponse) {
  return handleTemplateImage(req, res, async () =>
    getTemplate(String(req.query.templateId))
  );
}
