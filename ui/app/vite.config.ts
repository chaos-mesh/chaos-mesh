import react from '@vitejs/plugin-react-swc'
import path from 'path'
import { defineConfig, loadEnv } from 'vite'
import svgr from 'vite-plugin-svgr'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  return {
    base: './',
    legacy: {
      // MUI 5 icons and react-ace expose their components through CommonJS default exports.
      inconsistentCjsInterop: true,
    },
    resolve: {
      alias: {
        // https://github.com/react-dnd/react-dnd/issues/3416
        'react/jsx-runtime.js': 'react/jsx-runtime',
        'react/jsx-dev-runtime.js': 'react/jsx-dev-runtime',

        '@': path.resolve(__dirname, './src'),
      },
    },
    plugins: [react(), svgr()],
    build: {
      target: ['chrome107', 'edge107', 'firefox104', 'safari16'],
    },
    server: {
      proxy: {
        '/api': {
          target: env.VITE_API_BASE_URL,
        },
      },
    },
  }
})
