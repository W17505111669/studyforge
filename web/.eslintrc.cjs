/* eslint-env node */
module.exports = {
  root: true,
  env: {
    browser: true,
    es2022: true,
    node: true
  },
  parser: 'vue-eslint-parser',
  parserOptions: {
    ecmaVersion: 'latest',
    sourceType: 'module'
  },
  extends: [
    'plugin:vue/vue3-recommended',
    'eslint:recommended',
    'plugin:prettier/recommended'
  ],
  rules: {
    // Vue rules
    'vue/multi-word-component-names': 'off',
    'vue/no-v-html': 'off',
    'vue/require-default-prop': 'off',
    'vue/require-prop-types': 'off',
    'vue/no-unused-vars': 'warn',
    'vue/no-template-shadow': 'warn',
    'vue/attribute-hyphenation': 'off',
    'vue/v-on-event-hyphenation': 'off',

    // General rules
    'no-console': 'warn',
    'no-debugger': 'warn',
    'no-unused-vars': ['warn', { argsIgnorePattern: '^_', varsIgnorePattern: '^_' }],
    'no-undef': 'off', // Vue SFC globals are not recognized
    'no-empty': ['error', { allowEmptyCatch: true }],
    'no-constant-condition': 'warn',
    'prefer-const': 'warn',
    'no-var': 'error'
  },
  globals: {
    defineProps: 'readonly',
    defineEmits: 'readonly',
    defineExpose: 'readonly',
    withDefaults: 'readonly',
    __WIDGET_DATA__: 'readonly'
  }
}
