import path from "path";
import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  turbopack: {
    // Pin the project root to this folder so Next.js doesn't pick the repo root
    // because of a stray package-lock.json there.
    root: path.resolve(__dirname),
  },
};

export default nextConfig;
