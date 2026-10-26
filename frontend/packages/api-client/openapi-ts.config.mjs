// Generates the frontend Tadoku API client from the canonical contract. Run `pnpm --filter api-client generate`.

// `net/mail` atext: RFC 5322 atext plus every non-ASCII character (RFC 6532).
const atext = "[!#-'*+\\-/-9=?A-Z^-~\\u{80}-\\u{d7ff}\\u{e000}-\\u{10ffff}]"
const dotAtom = `${atext}+(?:\\.${atext}+)*`
// `net.ParseIP`: dotted decimal without leading zeros, or RFC 4291 text without a zone.
const octet = '(?:25[0-5]|2[0-4]\\d|1\\d\\d|[1-9]?\\d)'
const ipv4 = `${octet}(?:\\.${octet}){3}`
const h16 = '[\\dA-Fa-f]{1,4}'
const ls32 = `(?:${h16}:${h16}|${ipv4})`
const ipv6 = [
  `(?:${h16}:){6}${ls32}`,
  `::(?:${h16}:){5}${ls32}`,
  `(?:${h16})?::(?:${h16}:){4}${ls32}`,
  `(?:(?:${h16}:){0,1}${h16})?::(?:${h16}:){3}${ls32}`,
  `(?:(?:${h16}:){0,2}${h16})?::(?:${h16}:){2}${ls32}`,
  `(?:(?:${h16}:){0,3}${h16})?::${h16}:${ls32}`,
  `(?:(?:${h16}:){0,4}${h16})?::${ls32}`,
  `(?:(?:${h16}:){0,5}${h16})?::${h16}`,
  `(?:(?:${h16}:){0,6}${h16})?::`,
].join('|')

/**
 * An address `net/mail` reads back unchanged: a dot-atom local part and a dot-atom or IP-literal domain. This is the
 * email feature's `validAddress`, which admits every stored recipient; its 320-byte bound is the schema's
 * `maxLength`, checked in UTF-8 bytes. Stored account emails are not `format: email`.
 */
const emailPattern = `^${dotAtom}@(?:${dotAtom}|\\[(?:${ipv4}|${ipv6})\\])$`

/** Go's RFC 3339 reader: two-digit fields, seconds below 60 and a `Z` or `±hh:mm` offset under 24 hours. */
const goTime = 'T(?:[01]\\d|2[0-3]):[0-5]\\d:[0-5]\\d(?:\\.\\d+)?(?:Z|[+-](?:[01]\\d|2[0-3]):[0-5]\\d)$'

/**
 * Strings keep their format, length and pattern constraints, with the API's semantics where Zod's differ: lengths
 * count characters (code points), not UTF-16 units; UUIDs use GUID validation; `format: email` follows `validAddress`,
 * including its UTF-8 byte bound; `format: date-time` follows Go's RFC 3339 reader.
 */
const string = ctx => {
  const { $, chain, nodes, schema } = ctx
  const constant = nodes.const(ctx)
  if (constant) return constant
  chain.current = nodes.base(ctx)
  if (schema.format === 'email') chain.current = chain.current.attr('regex').call($.regexp(emailPattern, 'u'))
  else if (schema.format === 'uuid') chain.current = chain.current.attr('guid').call()
  else if (schema.format === 'date-time') chain.current = nodes.format(ctx).attr('regex').call($.regexp(goTime))
  else chain.current = nodes.format(ctx) ?? chain.current
  if (schema.minLength !== undefined || schema.maxLength !== undefined) {
    const characters = `^[\\s\\S]{${schema.minLength ?? 0},${schema.maxLength ?? ''}}$`
    chain.current = chain.current.attr('regex').call($.regexp(characters, 'u'))
  }
  chain.current = nodes.pattern(ctx) ?? chain.current
  if (schema.format === 'email' && schema.maxLength !== undefined) {
    // No regular expression counts UTF-8 bytes.
    const bytes = $.new('TextEncoder').attr('encode').call('value').attr('length')
    const fits = $.func(f => f.param('value').do($.return($.binary(bytes, '<=', $.literal(schema.maxLength)))))
    chain.current = chain.current.attr('refine').call(fits)
  }
  return chain.current
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
      compatibilityVersion: 4,
      // Go writes times in RFC 3339 with the time's own offset.
      dates: { offset: true },
      '~resolvers': { object: closedObject, string },
    },
    // The API validates requests; responses are validated before callers see them.
    { name: '@hey-api/sdk', validator: { request: false, response: 'zod' } },
  ],
}
