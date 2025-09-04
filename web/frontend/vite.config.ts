import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'node:path'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    host: true,
    proxy: {
      // Proxy API calls to Go backend
      '/api': {
        target: 'http://localhost:7878',
        changeOrigin: true,
        secure: false,
        configure: (proxy: { on: (event: string, handler: (...args: unknown[]) => void) => void }) => {
          proxy.on('error', (err: Error) => {
            console.log('proxy error', err);
          });
          proxy.on('proxyReq', (_proxyReq: unknown, req: { method: string; url: string }) => {
            console.log('Sending Request to the Target:', req.method, req.url);
          });
          proxy.on('proxyRes', (proxyRes: { statusCode: number }, req: { url: string }) => {
            console.log('Received Response from the Target:', proxyRes.statusCode, req.url);
          });
        },
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['react', 'react-dom'],
          redux: ['@reduxjs/toolkit', 'react-redux'],
          router: ['react-router-dom'],
        },
      },
    },
  },
  resolve: {
    alias: {
      '@': path.resolve(import.meta.dirname, './src'),
      '@components': path.resolve(import.meta.dirname, './src/components'),
      '@pages': path.resolve(import.meta.dirname, './src/pages'),
      '@hooks': path.resolve(import.meta.dirname, './src/hooks'),
      '@store': path.resolve(import.meta.dirname, './src/store'),
      '@utils': path.resolve(import.meta.dirname, './src/utils'),
      '@types': path.resolve(import.meta.dirname, './src/types'),
      '@styles': path.resolve(import.meta.dirname, './src/styles'),
      '@assets': path.resolve(import.meta.dirname, './src/assets'),
    },
  },
  css: {
    modules: {
      localsConvention: 'camelCase',
      generateScopedName: '[name]__[local]___[hash:base64:5]',
    },
    postcss: './postcss.config.js',
  },
})
