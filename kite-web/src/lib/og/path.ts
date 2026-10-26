import { isExport } from "@/lib/env/client";

// The path of the link preview image of a template, bundle or creator, or
// undefined in the export, which has no server for the API routes. The version
// makes a new version get a new URL, which crawlers fetch again.
export function ogImagePath(
  kind: "templates" | "bundles" | "creators",
  id: string,
  version: number | string
) {
  if (isExport) return undefined;
  return `/api/og/${kind}/${encodeURIComponent(id)}?v=${version}`;
}

// creatorImageVersion is the version of the link preview of a creator's page.
// It changes when they list or remove a template or one gets a new version.
export function creatorImageVersion(
  templates: { id: string; version: number }[]
) {
  const key = templates
    .map((t) => `${t.id}:${t.version}`)
    .sort()
    .join(",");
  let hash = 0;
  for (let i = 0; i < key.length; i++) {
    hash = (Math.imul(hash, 31) + key.charCodeAt(i)) | 0;
  }
  return (hash >>> 0).toString(36);
}
