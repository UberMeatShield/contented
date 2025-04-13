/// <reference types="vitest" />
import angular from '@analogjs/vite-plugin-angular';

import { defineConfig } from 'vite';

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  return {
    plugins: [
      angular({
        tsconfig: "src/tsconfig.spec.json",
        workspaceRoot: __dirname,
      })
    ],
    test: {
      globals: true,
      environment: 'jsdom',
      setupFiles: ['src/test-setup.ts'],
      include: ['**/*.spec.ts', '**/*.d.ts'],
      reporters: ['default'],
    },
    define: {
      'import.meta.vitest': mode !== 'production',
      global: {},
      VITEST: true
    },
  };
});
