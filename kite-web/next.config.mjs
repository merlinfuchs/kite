/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  transpilePackages: ["lucide-react"],
  output: process.env.OUTPUT === "export" ? "export" : "standalone",
};

export default nextConfig;
