import { defineConfig } from 'tsup'
import { readFile } from 'node:fs/promises'
import { resolve } from 'node:path'

export default defineConfig({
  entry: {
    index: 'src/index.ts',
    icons: 'src/icons/index.ts',
    catalog: 'src/catalog/index.ts',
    // Rewritten by scripts/build-assets.mjs; the entry supplies its declarations.
    'tokens.stylex': 'src/tokens/tokens.stylex.ts',
  },
  format: ['esm'],
  esbuildPlugins: [
    {
      // StyleX hashes variable names by file, so every entry shares one dist/tokens.stylex.js.
      name: 'stylex-tokens',
      setup(build) {
        build.onResolve({ filter: /\/tokens\.stylex$/ }, () => ({
          path: './tokens.stylex.js',
          external: true,
        }))
      },
    },
    {
      name: 'example-source',
      setup(build) {
        build.onResolve({ filter: /\.tsx\?raw$/ }, ({ path, resolveDir }) => ({
          path: resolve(resolveDir, path.slice(0, -4)),
          namespace: 'example-source',
        }))
        build.onLoad(
          { filter: /.*/, namespace: 'example-source' },
          async ({ path }) => ({
            contents: await readFile(path, 'utf8'),
            loader: 'text',
          }),
        )
      },
    },
  ],
  dts: true,
  clean: true,
  sourcemap: true,
  splitting: false,
  target: 'es2020',
  external: [
    'react',
    'react-dom',
    '@tanstack/react-form',
    '@stylexjs/stylex',
    /^@base-ui\/react(?:\/.*)?$/,
  ],
})
