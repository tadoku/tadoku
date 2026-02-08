/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  reactStrictMode: true,
  swcMinify: true,
  transpilePackages: ['ui'],
  webpack: config => {
    // Add raw source loader for ?raw imports
    config.module.rules.push({
      resourceQuery: /raw/,
      type: 'asset/source',
    })

    // Exclude ?raw files from normal Next.js processing
    config.module.rules.forEach(rule => {
      if (rule.oneOf) {
        rule.oneOf.forEach(oneOfRule => {
          if (oneOfRule.resourceQuery === undefined) {
            oneOfRule.resourceQuery = { not: [/raw/] }
          }
        })
      }
    })

    return config
  },
}

module.exports = nextConfig
