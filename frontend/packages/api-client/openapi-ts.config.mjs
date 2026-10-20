// Generates the frontend Tadoku API client from the canonical contract. Run `pnpm --filter api-client generate`.

/** Response validators check the JSON shape: types, required properties, nullability and enum values. */
const shapeOnly = ctx => ctx.nodes.base(ctx)

/** Objects the contract closes with `additionalProperties: false` reject unknown properties. */
const closedObject = ctx =>
  ctx.schema.additionalProperties?.type === 'never' ? ctx.nodes.base(ctx).attr('strict').call() : undefined

/** @type {import('@hey-api/openapi-ts').UserConfig} */
export default {
  input: '../../../services/tadoku-api/spec/openapi.yaml',
  output: { path: 'src/generated' },
  parser: {
    filters: {
      // Oathkeeper callbacks; browsers and frontend servers never call them.
      operations: { exclude: ['POST /authz/internal/v1/proxy/admin-check', 'GET /identity/internal/v1/session'] },
    },
    patch: {
      operations: {
        // The API requires JSON on every identity write; the contract documents logout without a body.
        'POST /identity/logout': operation => {
          operation.requestBody = { required: true, content: { 'application/json': { schema: { type: 'object' } } } }
        },
      },
      schemas: {
        // The API answers null for a contest without a language allow list; the contract documents an array.
        ImmersionContest: schema => {
          schema.allOf[1].properties.language_code_allow_list.nullable = true
        },
        ImmersionContestView: schema => {
          schema.allOf[1].properties.allowed_languages.nullable = true
        },
        // Test tenants evaluate flags in the Flipt test environment.
        ImmersionFeatureAccessResponse: schema => {
          schema.properties.environment.enum.push('test')
        },
      },
    },
  },
  plugins: [
    { name: '@hey-api/client-fetch', baseUrl: false },
    '@hey-api/typescript',
    {
      name: 'zod',
      // Legacy Next.js applications compile with TypeScript 4.9, which cannot read Zod 4 declarations.
      compatibilityVersion: 3,
      // Lengths, formats, patterns, ranges and item counts describe accepted input. JavaScript cannot check
      // them with the API's semantics (it counts UTF-16 units, not characters), and stored data predates
      // some limits, so response validators do not reject on them.
      '~resolvers': { array: shapeOnly, number: shapeOnly, object: closedObject, string: shapeOnly },
    },
    // The API validates requests; responses are validated before callers see them.
    { name: '@hey-api/sdk', validator: { request: false, response: 'zod' } },
  ],
}
