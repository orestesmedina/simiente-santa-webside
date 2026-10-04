/// <reference types="vitest/config" />
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [react(), tailwindcss()],
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: './src/test/setup.ts',
    css: false,
    // Las pruebas e2e de Playwright viven en `e2e/` y las ejecuta
    // `make e2e`; excluirlas evita que Vitest intente cargar `@playwright/test`
    // (sucesos con el mismo glob por defecto `**/*.spec.ts`).
    exclude: ['node_modules', 'dist', 'e2e'],
  },
});
