/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  transpilePackages: ["lucide-react"],
  output: process.env.OUTPUT === "export" ? "export" : "standalone",
  // Lets pages check for the export, which has no server for API routes.
  env: { OUTPUT: process.env.OUTPUT ?? "" },
  // Keeps server.js at the top of the standalone output, even when a parent
  // directory of the checkout has a lockfile.
  outputFileTracingRoot: import.meta.dirname,
  // Nothing uses next/image, so the server doesn't need its image optimizer.
  images: { unoptimized: true },
  poweredByHeader: false,
  // sharp is only used by the image optimizer, and its native binary would be
  // for the machine that builds, not the server.
  outputFileTracingExcludes: {
    "*": ["**/node_modules/@img/**", "**/node_modules/sharp/**"],
  },
  // Read by the link preview images.
  outputFileTracingIncludes: {
    "/api/og/**": ["./src/lib/og/*.ttf", "./public/logo.svg"],
  },
  experimental: {
    // Pages rebuilt by ISR stay in memory. The server's directory is read-only,
    // and after a restart the pages of the build are fresh enough.
    isrFlushToDisk: false,
  },
  // Don't write a kite-web/AGENTS.md on next dev, the repo has its own.
  agentRules: false,
  // Templates moved to the marketplace. The export for self-hosting has no
  // server to redirect, so its old links, like bookmarks of the app's
  // Templates page, stop working.
  ...(process.env.OUTPUT !== "export" && {
    // The bundles users listed change, so this sitemap is generated.
    rewrites: async () => [
      {
        source: "/sitemap-marketplace.xml",
        destination: "/api/sitemap-marketplace",
      },
    ],
    redirects: async () => [
      {
        source: "/templates/:path*",
        destination: "/marketplace/:path*",
        permanent: true,
      },
      {
        source: "/apps/:appId/templates/:path*",
        destination: "/apps/:appId/marketplace/:path*",
        permanent: true,
      },
    ],
  }),
};

export default nextConfig;
