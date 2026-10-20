// Generates the frontend Tadoku API client from the canonical contract. Run `pnpm --filter api-client generate`.

/**
 * Strings keep their format, length and pattern constraints, with the API's semantics where Zod's differ: the API
 * counts characters (code points) where Zod's `min` and `max` count UTF-16 units, and Go's `net/mail` accepts
 * addresses Zod's email rule rejects, such as a domain without a dot.
 */
const string = ctx => {
  const { $, chain, nodes, schema } = ctx
  const constant = nodes.const(ctx)
  if (constant) return constant
  chain.current = nodes.base(ctx)
  chain.current = schema.format === 'email'
    ? chain.current.attr('regex').call($.regexp('^.+@[^\\s@]+$'))
    : nodes.format(ctx) ?? chain.current
  if (schema.minLength !== undefined || schema.maxLength !== undefined) {
    const characters = `^[\\s\\S]{${schema.minLength ?? 0},${schema.maxLength ?? ''}}$`
    chain.current = chain.current.attr('regex').call($.regexp(characters, 'u'))
  }
  return nodes.pattern(ctx) ?? chain.current
}

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
  },
  plugins: [
    { name: '@hey-api/client-fetch', baseUrl: false },
    '@hey-api/typescript',
    {
      name: 'zod',
      // Legacy Next.js applications compile with TypeScript 4.9, which cannot read Zod 4 declarations.
      compatibilityVersion: 3,
      // Go writes times in RFC 3339 with the time's own offset.
      dates: { offset: true },
      '~resolvers': { object: closedObject, string },
    },
    // The API validates requests; responses are validated before callers see them.
    { name: '@hey-api/sdk', validator: { request: false, response: 'zod' } },
  ],
}
