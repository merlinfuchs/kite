import { getTemplate } from "@/lib/flow/templateFiles";
import { handleTemplateImage } from "@/lib/og/handler";
import type { NextApiRequest, NextApiResponse } from "next";

export default function handler(req: NextApiRequest, res: NextApiResponse) {
  const templateId = String(req.query.templateId);
  return handleTemplateImage(req, res, "templates", templateId, () =>
    getTemplate(templateId)
  );
}
