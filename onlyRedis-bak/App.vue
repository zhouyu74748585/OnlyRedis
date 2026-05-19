<template>
  <router-view />
</template>

<script setup lang="ts">
import type { GlobalThemeOverrides } from 'naive-ui'

const themeOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: '#3b82f6',
    primaryColorHover: '#2563eb',
    primaryColorPressed: '#1d4ed8',
    primaryColorSuppl: '#3b82f6',
    bodyColor: '#f5f6f8',
    cardColor: '#ffffff',
    modalColor: '#ffffff',
    popoverColor: '#ffffff',
    borderColor: 'rgba(0,0,0,0.08)',
    textColor1: '#1f2937',
    textColor2: '#6b7280',
    textColor3: '#9ca3af',
    placeholderColor: '#9ca3af',
    inputColor: '#f9fafb',
    fontSize: '14px',
    fontSizeMini: '12px',
    fontSizeTiny: '12px',
    fontSizeSmall: '13px',
    fontSizeMedium: '14px',
    fontSizeLarge: '16px',
  },
}
</script>

<style>
.n-config-provider {
  height: 100%;
}
</style>
