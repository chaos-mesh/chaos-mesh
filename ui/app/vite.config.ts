import react from '@vitejs/plugin-react-swc'
import path from 'path'
import { defineConfig, loadEnv } from 'vite'
import svgr from 'vite-plugin-svgr'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  return {
    // Relative base so the built bundle is relocatable under any subpath
    // (e.g. dashboard hosted at /chaos-mesh/ behind a prefix-rewriting ingress).
    // See https://vite.dev/guide/build.html#public-base-path
    base: './',
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
