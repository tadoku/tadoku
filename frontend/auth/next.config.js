/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  swcMinify: true,
  async rewrites() {
    return [
      {
        source: '/kratos/:path*',
        destination: 'http://localhost/kratos/:path*'
      }
    ]
  }
}

module.exports = nextConfig
