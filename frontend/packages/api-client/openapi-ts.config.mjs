// Generates the frontend Tadoku API client from the canonical contract. Run `pnpm --filter api-client generate`.

/** @type {import('@hey-api/openapi-ts').UserConfig} */
export default {
  input: '../../../services/tadoku-api/spec/openapi.yaml',
  output: { path: 'src/generated' },
  plugins: [
    { name: '@hey-api/client-fetch', baseUrl: false },
    '@hey-api/typescript',
    { name: 'zod', compatibilityVersion: 3 },
    { name: '@hey-api/sdk', validator: { request: false, response: 'zod' } },
  ],
}
