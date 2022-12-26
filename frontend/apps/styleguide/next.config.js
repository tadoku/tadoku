/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  reactStrictMode: true,
  swcMinify: true,
}

const withTM = require('next-transpile-modules')(['ui']);

module.exports = withTM(nextConfig)
