import { isExport } from "@/lib/env/client";
import { Template } from "@/lib/flow/templates";

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

// templateImageVersion is the version of the link preview of a template. Its
// name, description and icon change without a new version, so they're part
// of it.
export function templateImageVersion(template: Template) {
  return `${template.version}-${hash(
    [template.name, template.description, template.icon].join("\n")
  )}`;
}

// creatorImageVersion is the version of the link preview of a creator's page.
// It changes when they list or remove a template or one of them changes.
export function creatorImageVersion(templates: Template[]) {
  return hash(
    templates
      .map((t) => `${t.id}:${templateImageVersion(t)}`)
      .sort()
      .join(",")
  );
}

function hash(key: string) {
  let hash = 0;
  for (let i = 0; i < key.length; i++) {
    hash = (Math.imul(hash, 31) + key.charCodeAt(i)) | 0;
  }
  return (hash >>> 0).toString(36);
}
