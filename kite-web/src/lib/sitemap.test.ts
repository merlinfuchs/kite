import { readdirSync, readFileSync, statSync } from "node:fs";
import { join, relative } from "node:path";
import { describe, expect, it } from "vitest";
import { templatePaths } from "./flow/templates";

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

// The pages the export writes for each dynamic route, from the same function
// as its getStaticPaths.
const dynamicRoutes: Record<string, () => string[]> = {
  "/templates/[templateId]": templatePaths,
};

const indexableRoutes = pageFiles(pagesDir)
  .filter((file) => !readFileSync(file, "utf8").includes("noindex"))
  .map(routeOf);

const indexablePaths = indexableRoutes.flatMap((route) =>
  route.includes("[") ? dynamicRoutes[route]?.() ?? [] : [route]
);

describe("sitemap.xml", () => {
  it("only lists pages that can be indexed", () => {
    for (const path of sitemapPaths) {
      expect(indexablePaths, path).toContain(path);
    }
  });

  it("lists every page that can be indexed", () => {
    for (const path of indexablePaths) {
      expect(sitemapPaths, path).toContain(path);
    }
  });

  it("knows the pages of every dynamic route", () => {
    for (const route of indexableRoutes.filter((r) => r.includes("["))) {
      expect(Object.keys(dynamicRoutes), route).toContain(route);
    }
  });
});
