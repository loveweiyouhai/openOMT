<template>
  <header class="header">
    <div class="logo">
      <div class="logo-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
          <polyline points="3.27 6.96 12 12.01 20.73 6.96"/>
          <line x1="12" y1="22.08" x2="12" y2="12"/>
        </svg>
      </div>
      <span class="logo-name">openOMT</span>
    </div>
    <div class="header-center">
      <div v-if="connected" class="status-indicator online">
        <span class="status-dot"></span>
        <span class="status-text">{{ connectionName || '已连接' }}</span>
      </div>
    </div>
    <div class="header-actions">
      <div class="theme-switcher">
        <button 
          class="theme-btn" 
          :class="{ active: theme === 'dark' }"
          @click="setTheme('dark')"
          title="暗色主题"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
          </svg>
        </button>
        <button 
          class="theme-btn" 
          :class="{ active: theme === 'light' }"
          @click="setTheme('light')"
          title="亮色主题"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="5"/>
            <line x1="12" y1="1" x2="12" y2="3"/>
            <line x1="12" y1="21" x2="12" y2="23"/>
            <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/>
            <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/>
            <line x1="1" y1="12" x2="3" y2="12"/>
            <line x1="21" y1="12" x2="23" y2="12"/>
            <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/>
            <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/>
          </svg>
        </button>
      </div>
    </div>
  </header>
</template>

<script setup>
import { ref, onMounted } from 'vue'

defineProps({
  connected: Boolean,
  connectionName: String,
})

const theme = ref('dark')

const setTheme = (newTheme) => {
  theme.value = newTheme
  document.documentElement.setAttribute('data-theme', newTheme)
  localStorage.setItem('theme', newTheme)
}

onMounted(() => {
  const savedTheme = localStorage.getItem('theme') || 'dark'
  theme.value = savedTheme
  document.documentElement.setAttribute('data-theme', savedTheme)
})
</script>

<style lang="scss" scoped>
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  height: var(--header-height);
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border);
  backdrop-filter: blur(12px);
  position: relative;
  z-index: 100;
}

.logo {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logo-icon {
  width: 38px;
  height: 38px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #58a6ff 0%, #a371f7 100%);
  border-radius: 10px;
  box-shadow: 0 4px 12px rgba(88, 166, 255, 0.25);
  
  svg {
    width: 20px;
    height: 20px;
    color: white;
  }
}

.logo-name {
  font-weight: 700;
  font-size: 17px;
  letter-spacing: -0.3px;
  background: linear-gradient(90deg, var(--text-primary), var(--accent));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.header-center {
  flex: 1;
  display: flex;
  justify-content: center;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 16px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: 20px;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  transition: all 0.3s ease;
  
  &.online {
    background: var(--success-light);
    border-color: rgba(63, 185, 80, 0.3);
    color: var(--success);
    
    .status-dot {
      background: var(--success);
      box-shadow: 0 0 8px var(--success), 0 0 16px rgba(63, 185, 80, 0.3);
      animation: pulse 2s infinite;
    }
  }
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--text-muted);
  transition: all 0.3s ease;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.theme-switcher {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: 8px;
}

.theme-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: transparent;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  color: var(--text-muted);
  
  svg {
    width: 18px;
    height: 18px;
  }
  
  &:hover {
    background: var(--bg-hover);
    color: var(--text-primary);
  }
  
  &.active {
    background: var(--accent);
    color: white;
    box-shadow: 0 2px 8px rgba(88, 166, 255, 0.3);
  }
}
</style>
