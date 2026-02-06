declare module 'prettier/standalone' {
  import { Options } from 'prettier'
  export function format(source: string, options: Options): string
  export default { format }
}

declare module 'prettier/parser-markdown' {
  const plugin: import('prettier').Plugin
  export default plugin
}

declare module 'prettier/parser-html' {
  const plugin: import('prettier').Plugin
  export default plugin
}
