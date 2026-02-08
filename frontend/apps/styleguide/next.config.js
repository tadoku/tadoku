/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  reactStrictMode: true,
  swcMinify: true,
  transpilePackages: ['ui'],
  webpack: config => {
    config.module.rules.push({
      resourceQuery: /raw/,
      type: 'asset/source',
    })
    return config
  },
}

module.exports = nextConfig
