import { createHash } from 'node:crypto'

// TanStack Start names the handler of `const x = createServerFn()` `x_createServerFn_handler`.
const handlerSuffix = '_createServerFn_handler'

/**
 * The ID a built server function answers at `/_serverFn/<id>`, from its source file relative to the app and the
 * variable it is assigned to. vite.config.ts passes this to TanStack Start, and the image smoke calls it with the same
 * file and variable to reach a function without reading the bundle.
 */
export function serverFunctionId(filename, name) {
  const variable = name.endsWith(handlerSuffix) ? name.slice(0, -handlerSuffix.length) : name
  return createHash('sha256').update(`${filename}--${variable}`).digest('hex')
}
