import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import { join, relative } from "node:path";
import { describe, expect, it } from "vitest";
import { getTemplate, getTemplates } from "./flow/templates";

const pagesDir = join(__dirname, "../pages");
const sitemap = readFileSync(
  join(__dirname, "../../public/sitemap.xml"),
  "utf8"
);
const sitemapPaths = [
  ...sitemap.matchAll(/<loc>https:\/\/kite\.onl([^<]*)<\/loc>/g),
].map((m) => m[1] || "/");

// The dashboard, embeds and API routes are never indexed.
const ignoredDirs = ["apps", "embed", "api"];

function pageFiles(dir: string): string[] {
  return readdirSync(dir).flatMap((name) => {
    const path = join(dir, name);
    if (statSync(path).isDirectory()) {
      return ignoredDirs.includes(relative(pagesDir, path))
        ? []
        : pageFiles(path);
    }
    return name.endsWith(".tsx") && !name.startsWith("_") ? [path] : [];
  });
}

function routeOf(file: string) {
  return "/" + relative(pagesDir, file).replace(/(\/?index)?\.tsx$/, "");
}

// Template pages are exported from templates/[templateId].tsx.
const templatePath = /^\/templates\/([^/]+)$/;

describe("sitemap.xml", () => {
  it("only lists existing pages", () => {
    for (const path of sitemapPaths) {
      const templateId = path.match(templatePath)?.[1];
      if (templateId) {
        expect(getTemplate(templateId), path).toBeDefined();
        continue;
      }
      const base = join(pagesDir, path === "/" ? "index" : path);
      expect(
        existsSync(base + ".tsx") || existsSync(join(base, "index.tsx")),
        path
      ).toBe(true);
    }
  });

  it("lists every page that can be indexed", () => {
    for (const file of pageFiles(pagesDir)) {
      if (file.includes("[")) continue;
      if (readFileSync(file, "utf8").includes("noindex")) continue;
      expect(sitemapPaths, routeOf(file)).toContain(routeOf(file));
    }
  });

  it("lists every template", () => {
    for (const template of getTemplates()) {
      expect(sitemapPaths).toContain(`/templates/${template.id}`);
    }
  });
});
