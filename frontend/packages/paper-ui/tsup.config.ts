import { defineConfig } from 'tsup'

export default defineConfig({
  entry: {
    index: 'src/index.ts',
    icons: 'src/icons/index.ts',
    catalog: 'src/catalog/index.ts',
  },
  format: ['esm'],
  dts: true,
  clean: true,
  sourcemap: true,
  splitting: false,
  target: 'es2020',
  external: ['react', 'react-dom', 'react-hook-form', /^@base-ui\/react(?:\/.*)?$/],
})
