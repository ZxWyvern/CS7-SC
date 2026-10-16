import eslintPluginAstro from 'eslint-plugin-astro';
import tseslint from 'typescript-eslint';
import eslintJs from '@eslint/js';

export default [
  {
    ignores: ['dist/', '.astro/', 'node_modules/', 'coverage/'],
  },
  eslintJs.configs.recommended,
  ...tseslint.configs.recommended,
  ...eslintPluginAstro.configs['flat/recommended'],
  {
    rules: {
      // frontend-only kelas: no-console di astro components kecuali warn/error
      'no-console': ['warn', { allow: ['warn', 'error'] }],
    },
  },
];
