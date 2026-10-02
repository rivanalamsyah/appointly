// @ts-check
import { defineConfig } from 'astro/config';
import react from '@astrojs/react';
import tailwindcss from '@tailwindcss/vite';
import node from '@astrojs/node';

// https://astro.build/config
export default defineConfig({
  output: 'server',
  adapter: node({
    mode: 'standalone',
  }),
  integrations: [
    react(),
  ],
  server: {
    port: 4321,
    host: true,
  },
  vite: {
    plugins: [
      tailwindcss(),
    ],
    server: {
      proxy: {
        '/api': {
          target: process.env.API_URL || 'http://localhost:8080',
          changeOrigin: true,
        },
      },
    },
    define: {
      'import.meta.env.API_BASE_URL': JSON.stringify(
        process.env.API_URL || 'http://localhost:8080'
      ),
    },
  },
});
