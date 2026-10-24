/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  transpilePackages: ["lucide-react"],
  output: process.env.OUTPUT === "export" ? "export" : "standalone",
  // Don't write a kite-web/AGENTS.md on next dev, the repo has its own.
  agentRules: false,
};

export default nextConfig;
