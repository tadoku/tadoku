import { defaultSerovalDeserializerPlugins } from '@tanstack/react-start'
import { fromJSON } from 'seroval'

const formContentType = /multipart\/form-data|application\/x-www-form-urlencoded/

/**
 * Whether TanStack Start's server function handler can decode `request`'s payload, read the way the handler reads it.
 * It consumes the body, so pass a clone taken before the handler runs. It leaves out Start's serialization adapters, so
 * a payload carrying an adapted value, such as a server function, does not decode; Webv3 sends none.
 */
export async function serverFunctionPayloadDecodes(request: Request) {
  const contentType = request.headers.get('Content-Type') ?? ''
  const plugins = defaultSerovalDeserializerPlugins
  try {
    if (formContentType.test(contentType)) {
      await request.formData()
    } else if (request.method === 'GET') {
      const payload = new URL(request.url).searchParams.get('payload')
      if (payload) fromJSON(JSON.parse(payload), { plugins })
    } else if (contentType.includes('application/json')) {
      fromJSON(await request.json(), { plugins })
    }
    return true
  } catch {
    return false
  }
}
