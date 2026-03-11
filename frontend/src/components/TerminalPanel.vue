<template>
  <div class="terminal-panel">
    <!-- 服务器连接标签栏 -->
    <div class="connection-bar">
      <div class="connection-tabs">
        <div 
          v-for="conn in connectionList" 
          :key="conn.id"
          class="connection-tab"
          :class="{ active: conn.id === activeConnId }"
          @click="$emit('switch-connection', conn.id)"
        >
          <el-icon><Monitor /></el-icon>
          <span>{{ conn.name || conn.host }}</span>
          <el-icon class="tab-close" @click.stop="$emit('close-connection', conn.id)"><Close /></el-icon>
        </div>
      </div>
    </div>

    <!-- 终端标签栏 -->
    <div class="terminal-header" v-if="activeConnId">
      <div class="terminal-tabs">
        <div 
          v-for="(shell, index) in currentShellTabs" 
          :key="shell.id"
          class="terminal-tab"
          :class="{ active: shell.id === currentActiveShellId }"
          @click="switchShell(shell.id)"
        >
          <el-icon><Monitor /></el-icon>
          <span>终端 {{ index + 1 }}</span>
          <el-icon class="tab-close" @click.stop="closeShell(shell.id)"><Close /></el-icon>
        </div>
        <div class="terminal-tab add-tab" @click="addShell">
          <el-icon><Plus /></el-icon>
        </div>
      </div>
      <div class="terminal-actions" v-if="currentActiveShellId">
        <el-button size="small" @click="clearTerminal">
          <el-icon><Delete /></el-icon>
          清屏
        </el-button>
      </div>
    </div>

    <div v-if="!activeConnId" class="terminal-empty">
      <el-icon :size="48"><Monitor /></el-icon>
      <p>请先连接服务器</p>
    </div>

    <div v-else-if="currentShellTabs.length === 0" class="terminal-empty">
      <el-icon :size="48"><Monitor /></el-icon>
      <p>点击 + 打开终端</p>
    </div>

    <div ref="terminalContainer" class="terminal-container" v-show="currentActiveShellId"></div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { Monitor, Plus, Close, Delete } from '@element-plus/icons-vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import '@xterm/xterm/css/xterm.css'

const props = defineProps({
  connections: Object,
  activeConnId: String,
})

const emit = defineEmits(['switch-connection', 'close-connection'])

const terminalContainer = ref()

// 每个连接的终端会话: { connId: { shells: [], activeShellId: null } }
const connectionShells = ref({})

let resizeHandler = null

const connectionList = computed(() => {
  if (!props.connections) return []
  return Object.values(props.connections)
})

const currentShellTabs = computed(() => {
  if (!props.activeConnId) return []
  return connectionShells.value[props.activeConnId]?.shells || []
})

const currentActiveShellId = computed(() => {
  if (!props.activeConnId) return null
  return connectionShells.value[props.activeConnId]?.activeShellId || null
})

const terminalThemes = {
  dark: {
    background: '#0d1117',
    foreground: '#c9d1d9',
    cursor: '#58a6ff',
    cursorAccent: '#0d1117',
    selectionBackground: '#3b5070',
    black: '#484f58',
    red: '#ff7b72',
    green: '#3fb950',
    yellow: '#d29922',
    blue: '#58a6ff',
    magenta: '#bc8cff',
    cyan: '#39c5cf',
    white: '#b1bac4',
    brightBlack: '#6e7681',
    brightRed: '#ffa198',
    brightGreen: '#56d364',
    brightYellow: '#e3b341',
    brightBlue: '#79c0ff',
    brightMagenta: '#d2a8ff',
    brightCyan: '#56d4dd',
    brightWhite: '#f0f6fc',
  },
  light: {
    background: '#ffffff',
    foreground: '#1f2328',
    cursor: '#0969da',
    cursorAccent: '#ffffff',
    selectionBackground: '#add6ff',
    black: '#1f2328',
    red: '#cf222e',
    green: '#1a7f37',
    yellow: '#9a6700',
    blue: '#0969da',
    magenta: '#8250df',
    cyan: '#1b7c83',
    white: '#6e7781',
    brightBlack: '#57606a',
    brightRed: '#a40e26',
    brightGreen: '#2da44e',
    brightYellow: '#bf8700',
    brightBlue: '#218bff',
    brightMagenta: '#a475f9',
    brightCyan: '#3192aa',
    brightWhite: '#8c959f',
  }
}

function getCurrentTheme() {
  return document.documentElement.getAttribute('data-theme') || 'dark'
}

function createTerminal() {
  const theme = getCurrentTheme()
  return new Terminal({
    cursorBlink: true,
    cursorStyle: 'block',
    fontSize: 14,
    fontFamily: "'JetBrains Mono', 'Fira Code', 'Consolas', monospace",
    theme: terminalThemes[theme],
    scrollback: 10000,
    allowProposedApi: true,
  })
}

function updateAllTerminalThemes() {
  const theme = getCurrentTheme()
  Object.values(connectionShells.value).forEach(connShells => {
    connShells.shells.forEach(shell => {
      if (shell.terminal) {
        shell.terminal.options.theme = terminalThemes[theme]
      }
    })
  })
}

function ensureConnectionShells(connId) {
  if (!connectionShells.value[connId]) {
    connectionShells.value[connId] = {
      shells: [],
      activeShellId: null,
    }
  }
  return connectionShells.value[connId]
}

async function addShell() {
  if (!props.activeConnId) return

  const connId = props.activeConnId
  const connShells = ensureConnectionShells(connId)

  try {
    const result = await window.go?.main?.App?.StartShell(connId)
    if (result?.error) {
      console.error('启动终端失败:', result.error)
      return
    }

    const shellId = result.shellId
    const term = createTerminal()
    const fit = new FitAddon()
    term.loadAddon(fit)
    term.loadAddon(new WebLinksAddon())

    const eventCleanups = []
    if (window.runtime?.EventsOn) {
      const outputCleanup = window.runtime.EventsOn(`shell-output:${connId}:${shellId}`, (data) => {
        term.write(data)
        term.scrollToBottom()
      })
      eventCleanups.push(outputCleanup)

      const closeCleanup = window.runtime.EventsOn(`shell-closed:${connId}:${shellId}`, () => {
        term.writeln('\r\n\x1b[31m会话已断开\x1b[0m')
        term.scrollToBottom()
      })
      eventCleanups.push(closeCleanup)
    }

    term.onData(data => {
      window.go?.main?.App?.WriteShell(connId, shellId, data)
    })

    term.onResize(({ rows, cols }) => {
      window.go?.main?.App?.ResizeShell(connId, shellId, rows, cols)
    })

    const shellTab = {
      id: shellId,
      connId,
      terminal: term,
      fitAddon: fit,
      eventCleanups,
    }

    // 先卸载当前活跃的终端
    unmountCurrentTerminal()

    connShells.shells.push(shellTab)
    connShells.activeShellId = shellId

    await nextTick()
    mountTerminal(shellTab)
  } catch (e) {
    console.error('创建终端失败:', e)
  }
}

function unmountCurrentTerminal() {
  if (!props.activeConnId) return
  const connShells = connectionShells.value[props.activeConnId]
  if (!connShells) return
  
  const current = connShells.shells.find(s => s.id === connShells.activeShellId)
  if (current?.terminal?.element?.parentNode) {
    current.terminal.element.parentNode.removeChild(current.terminal.element)
  }
}

function mountTerminal(shell) {
  if (!terminalContainer.value || !shell) return

  if (!shell.terminal.element) {
    shell.terminal.open(terminalContainer.value)
  } else if (shell.terminal.element.parentNode !== terminalContainer.value) {
    terminalContainer.value.appendChild(shell.terminal.element)
  }

  shell.fitAddon.fit()

  const dims = shell.fitAddon.proposeDimensions()
  if (dims) {
    window.go?.main?.App?.ResizeShell(shell.connId, shell.id, dims.rows, dims.cols)
  }
}

function switchShell(shellId) {
  if (!props.activeConnId) return
  const connShells = connectionShells.value[props.activeConnId]
  if (!connShells || connShells.activeShellId === shellId) return

  unmountCurrentTerminal()
  connShells.activeShellId = shellId

  nextTick(() => {
    const shell = connShells.shells.find(s => s.id === shellId)
    if (shell) {
      mountTerminal(shell)
    }
  })
}

async function closeShell(shellId) {
  if (!props.activeConnId) return
  const connShells = connectionShells.value[props.activeConnId]
  if (!connShells) return

  const index = connShells.shells.findIndex(s => s.id === shellId)
  if (index === -1) return

  const shell = connShells.shells[index]
  shell.eventCleanups?.forEach(fn => fn?.())

  try {
    await window.go?.main?.App?.CloseShell(shell.connId, shellId)
  } catch (e) {
    console.error('关闭终端失败:', e)
  }

  shell.terminal?.dispose()
  connShells.shells.splice(index, 1)

  if (connShells.activeShellId === shellId) {
    if (connShells.shells.length > 0) {
      const newIndex = Math.min(index, connShells.shells.length - 1)
      connShells.activeShellId = connShells.shells[newIndex].id
      nextTick(() => {
        const newShell = connShells.shells[newIndex]
        if (newShell) {
          mountTerminal(newShell)
        }
      })
    } else {
      connShells.activeShellId = null
    }
  }
}

function clearTerminal() {
  if (!props.activeConnId) return
  const connShells = connectionShells.value[props.activeConnId]
  if (!connShells) return
  
  const shell = connShells.shells.find(s => s.id === connShells.activeShellId)
  shell?.terminal?.clear()
}

function handleResize() {
  if (!props.activeConnId) return
  const connShells = connectionShells.value[props.activeConnId]
  if (!connShells?.activeShellId) return
  
  const shell = connShells.shells.find(s => s.id === connShells.activeShellId)
  shell?.fitAddon?.fit()
}

// 切换连接时，卸载旧连接的终端，挂载新连接的终端
watch(() => props.activeConnId, async (newId, oldId) => {
  // 卸载旧连接的当前终端
  if (oldId && connectionShells.value[oldId]) {
    const oldShells = connectionShells.value[oldId]
    const oldShell = oldShells.shells.find(s => s.id === oldShells.activeShellId)
    if (oldShell?.terminal?.element?.parentNode) {
      oldShell.terminal.element.parentNode.removeChild(oldShell.terminal.element)
    }
  }

  // 挂载新连接的当前终端，或自动创建一个
  if (newId) {
    const connShells = connectionShells.value[newId]
    if (connShells && connShells.shells.length > 0) {
      await nextTick()
      const shell = connShells.shells.find(s => s.id === connShells.activeShellId)
      if (shell) {
        mountTerminal(shell)
      }
    } else {
      // 新连接，自动打开一个终端
      await nextTick()
      addShell()
    }
  }
})

// 连接断开时清理对应的终端
watch(() => props.connections, (newConns) => {
  if (!newConns) return
  const connIds = Object.keys(newConns)
  
  // 清理已断开连接的终端
  for (const connId of Object.keys(connectionShells.value)) {
    if (!connIds.includes(connId)) {
      const connShells = connectionShells.value[connId]
      for (const shell of connShells.shells) {
        shell.eventCleanups?.forEach(fn => fn?.())
        shell.terminal?.dispose()
      }
      delete connectionShells.value[connId]
    }
  }
}, { deep: true })

let themeObserver = null

onMounted(() => {
  resizeHandler = handleResize
  window.addEventListener('resize', resizeHandler)
  
  // 监听主题变化
  themeObserver = new MutationObserver((mutations) => {
    mutations.forEach((mutation) => {
      if (mutation.attributeName === 'data-theme') {
        updateAllTerminalThemes()
      }
    })
  })
  themeObserver.observe(document.documentElement, { attributes: true })
})

onUnmounted(() => {
  if (resizeHandler) {
    window.removeEventListener('resize', resizeHandler)
  }
  if (themeObserver) {
    themeObserver.disconnect()
  }
  
  // 清理所有终端
  for (const connId of Object.keys(connectionShells.value)) {
    const connShells = connectionShells.value[connId]
    for (const shell of connShells.shells) {
      shell.eventCleanups?.forEach(fn => fn?.())
      window.go?.main?.App?.CloseShell(shell.connId, shell.id)
      shell.terminal?.dispose()
    }
  }
  connectionShells.value = {}
})
</script>

<style lang="scss" scoped>
.terminal-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--terminal-bg);
  min-height: 0;
  overflow: hidden;
}

.connection-bar {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border);
}

.connection-tabs {
  display: flex;
  align-items: center;
  gap: 6px;
  overflow-x: auto;
  flex: 1;
  
  &::-webkit-scrollbar {
    height: 4px;
  }
  
  &::-webkit-scrollbar-thumb {
    background: var(--border);
    border-radius: 2px;
  }
}

.connection-tab {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  font-size: 13px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.15s ease;
  white-space: nowrap;

  &:hover {
    background: var(--bg-hover);
    color: var(--text-primary);
  }

  &.active {
    background: var(--accent-light);
    border-color: var(--accent);
    color: var(--accent);

    .el-icon:first-child {
      color: var(--success);
    }
  }

  .tab-close {
    font-size: 12px;
    padding: 2px;
    border-radius: var(--radius-sm);
    margin-left: 4px;
    opacity: 0.6;
    transition: all 0.15s ease;

    &:hover {
      opacity: 1;
      background: rgba(255, 255, 255, 0.1);
      color: var(--danger);
    }
  }
}

.terminal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border);
}

.terminal-tabs {
  display: flex;
  align-items: center;
  gap: 4px;
  overflow-x: auto;
  flex: 1;
  margin-right: 12px;
}

.terminal-tab {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  font-size: 12px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.15s ease;
  white-space: nowrap;

  &:hover {
    background: var(--bg-hover);
    color: var(--text-primary);
  }

  &.active {
    background: var(--bg-primary);
    border-color: var(--accent);
    color: var(--text-primary);

    .el-icon:first-child {
      color: var(--success);
    }
  }

  .tab-close {
    font-size: 12px;
    padding: 2px;
    border-radius: var(--radius-sm);
    margin-left: 4px;
    opacity: 0.6;
    transition: all 0.15s ease;

    &:hover {
      opacity: 1;
      background: rgba(255, 255, 255, 0.1);
      color: var(--danger);
    }
  }

  &.add-tab {
    padding: 6px 10px;
    background: transparent;
    border: 1px dashed var(--border);

    &:hover {
      border-color: var(--accent);
      color: var(--accent);
      background: var(--accent-light);
    }
  }
}

.terminal-actions {
  display: flex;
  gap: 8px;
}

.terminal-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  color: var(--text-muted);
  
  .el-icon {
    opacity: 0.4;
  }
  
  p {
    margin: 0;
    font-size: 14px;
  }
}

.terminal-container {
  flex: 1;
  padding: 8px;
  padding-bottom: 32px;
  overflow: hidden;
  min-height: 0;

  :deep(.xterm) {
    height: 100%;
  }

  :deep(.xterm-viewport) {
    overflow-y: auto !important;

    &::-webkit-scrollbar {
      width: 8px;
    }

    &::-webkit-scrollbar-track {
      background: transparent;
    }

    &::-webkit-scrollbar-thumb {
      background: var(--bg-hover);
      border-radius: 4px;

      &:hover {
        background: var(--border-light);
      }
    }
  }
}
</style>
