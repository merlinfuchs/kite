import { isExport } from "@/lib/env/client";

// The path of the link preview image of a template or bundle, or undefined in
// the export, which has no server for the API routes. The version makes a new
// version get a new URL, which crawlers fetch again.
export function ogImagePath(
  kind: "templates" | "bundles",
  id: string,
  version: number
) {
  if (isExport) return undefined;
  return `/api/og/${kind}/${encodeURIComponent(id)}?v=${version}`;
}
