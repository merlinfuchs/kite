/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  transpilePackages: ["lucide-react"],
  output: process.env.OUTPUT === "export" ? "export" : "standalone",

  async rewrites() {
    return [
      {
        source: "/pl.js",
        destination:
          "https://insights.xenon.bot/js/script.outbound-links.pageview-props.revenue.tagged-events.js",
      },
      {
        source: "/api/event",
        destination: "https://insights.xenon.bot/api/event",
      },
    ];
  },
};

export default nextConfig;
