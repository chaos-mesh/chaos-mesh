import react from '@vitejs/plugin-react-swc'
import path from 'path'
import { defineConfig, loadEnv } from 'vite'
import svgr from 'vite-plugin-svgr'

// https://vite.dev/config/
export default defineConfig(({ command, mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  // For production builds, default to a relative base ('./') so the bundle is
  // relocatable under any subpath (e.g. dashboard hosted at /chaos-mesh/ behind
  // a prefix-rewriting ingress). Downstream builders that need an absolute base
  // (e.g. assets on a different origin) can override with VITE_BASE.
  // Dev mode keeps the default '/' so HMR and /@vite/client work as expected.
  // See https://vite.dev/guide/build.html#public-base-path
  const base = command === 'build' ? env.VITE_BASE || './' : '/'

  return {
    base,
    resolve: {
      alias: {
        // https://github.com/react-dnd/react-dnd/issues/3416
        'react/jsx-runtime.js': 'react/jsx-runtime',
        'react/jsx-dev-runtime.js': 'react/jsx-dev-runtime',

        '@': path.resolve(__dirname, './src'),
      },
    },
    plugins: [react(), svgr()],
    server: {
      proxy: {
        '/api': {
          target: env.VITE_API_BASE_URL,
        },
      },
    },
  }
})
