/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  reactStrictMode: true,
  // Needs to be disabled until swc is fixed
  swcMinify: false,
}

const withTM = require('next-transpile-modules')(['ui']);

module.exports = withTM(nextConfig)
