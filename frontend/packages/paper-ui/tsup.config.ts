import { defineConfig } from 'tsup'
import { readFile } from 'node:fs/promises'
import { resolve } from 'node:path'

export default defineConfig({
  entry: {
    index: 'src/index.ts',
    icons: 'src/icons/index.ts',
    catalog: 'src/catalog/index.ts',
  },
  format: ['esm'],
  esbuildPlugins: [
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
    'react-hook-form',
    /^@base-ui\/react(?:\/.*)?$/,
  ],
})
