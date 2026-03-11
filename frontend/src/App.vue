<template>
  <div class="app">
    <AppHeader 
      :connected="hasActiveConnection"
      :connection-name="activeConnectionName"
      @disconnect="handleDisconnectActive"
    />
    <div class="app-body">
      <!-- 左侧面板：服务器列表 + 文件管理 -->
      <div class="left-panel">
        <!-- 服务器列表 -->
        <div class="server-section">
          <div class="section-header">
            <span class="section-title">
              <el-icon><Monitor /></el-icon>
              服务器
            </span>
            <el-button type="primary" size="small" @click="connectDialog.visible = true">
              <el-icon><Plus /></el-icon>
              添加
            </el-button>
          </div>
          <div class="server-list">
            <div
              v-for="conn in state.savedConnections"
              :key="conn.id"
              class="server-item"
              :title="`${conn.host}:${conn.port}`"
              @click="handleConnectSaved(conn.id)"
              @contextmenu.prevent="showServerContextMenu($event, conn)"
            >
              <div class="server-icon">
                <el-icon><Monitor /></el-icon>
              </div>
              <div class="server-info">
                <span class="server-name">{{ conn.name || conn.host }}</span>
                <span class="server-meta">{{ conn.host }}:{{ conn.port }}</span>
              </div>
            </div>
            <div v-if="state.savedConnections.length === 0" class="server-empty">
              <p>暂无服务器</p>
            </div>
          </div>
        </div>

        <!-- 文件管理 -->
        <div class="file-section">
          <div class="section-header">
            <span class="section-title">
              <el-icon><Folder /></el-icon>
              文件管理
            </span>
            <span class="active-conn-name" v-if="hasActiveConnection">
              {{ activeConnectionName }}
            </span>
          </div>

          <div v-if="!hasActiveConnection" class="file-empty">
            <el-icon :size="48"><FolderOpened /></el-icon>
            <p>请先连接服务器</p>
          </div>
          
          <template v-else>
            <div class="file-toolbar">
              <el-button 
                v-if="currentPath !== '/'" 
                class="toolbar-btn"
                @click="handleGoUp"
                title="返回上级"
              >
                <el-icon :size="18"><ArrowLeft /></el-icon>
              </el-button>
              <el-button class="toolbar-btn" @click="handleGoHome" title="根目录">
                <el-icon :size="18"><HomeFilled /></el-icon>
              </el-button>
              <div class="path-display">{{ currentPath || '/' }}</div>
              <el-button class="toolbar-btn" @click="handleRefresh" title="刷新">
                <el-icon :size="18"><Refresh /></el-icon>
              </el-button>
              <div class="toolbar-divider"></div>
              <el-button class="toolbar-btn" @click="handleNewFolder" title="新建文件夹">
                <el-icon :size="18"><FolderAdd /></el-icon>
              </el-button>
              <el-button class="toolbar-btn" @click="handleNewFile" title="新建文件">
                <el-icon :size="18"><DocumentAdd /></el-icon>
              </el-button>
              <el-button class="toolbar-btn" @click="triggerUpload" title="上传">
                <el-icon :size="18"><Upload /></el-icon>
              </el-button>
            </div>

            <div class="file-list" @contextmenu.prevent="showFileContextMenu($event, null)">
              <div
                v-for="file in currentFiles"
                :key="file.name"
                class="file-item"
                @click="handleFileClick(file)"
                @dblclick="handleFileDbClick(file)"
                @contextmenu.prevent.stop="showFileContextMenu($event, file)"
              >
                <div class="file-icon" :class="file.isDir ? 'is-folder' : getFileType(file.name)">
                  <el-icon v-if="file.isDir"><FolderOpened /></el-icon>
                  <el-icon v-else><Document /></el-icon>
                </div>
                <div class="file-info">
                  <span class="file-name">{{ file.name }}</span>
                  <span class="file-meta">
                    {{ file.isDir ? '文件夹' : formatSize(file.size) }}
                    <template v-if="file.modTime"> · {{ formatTime(file.modTime) }}</template>
                  </span>
                </div>
              </div>
              <div v-if="currentFiles.length === 0" class="file-empty-list">
                <p>此目录为空</p>
              </div>
            </div>
          </template>
        </div>
      </div>

      <!-- 右侧终端 -->
      <div class="right-panel">
        <TerminalPanel
          :connections="state.connections"
          :active-conn-id="state.activeConnId"
          @switch-connection="handleSwitchConnection"
          @close-connection="handleCloseConnection"
        />
      </div>
    </div>

    <!-- 底部版权信息 -->
    <footer class="app-footer">
      <span>openOMT v1.0</span>
      <span class="separator">|</span>
      <span>© 2026 All Rights Reserved</span>
    </footer>

    <!-- 文件右键菜单 -->
    <div 
      v-show="fileContextMenu.visible" 
      class="context-menu"
      :style="{ left: fileContextMenu.x + 'px', top: fileContextMenu.y + 'px' }"
    >
      <div class="menu-item" @click="handleContextMenuAction('upload')">
        <el-icon><Upload /></el-icon>
        <span>上传文件</span>
      </div>
      <div class="menu-item" @click="handleContextMenuAction('newFolder')">
        <el-icon><FolderAdd /></el-icon>
        <span>新建文件夹</span>
      </div>
      <div class="menu-item" @click="handleContextMenuAction('newFile')">
        <el-icon><DocumentAdd /></el-icon>
        <span>新建文件</span>
      </div>
      <template v-if="fileContextMenu.target">
        <div class="menu-divider"></div>
        <div class="menu-item" v-if="!fileContextMenu.target.isDir" @click="handleContextMenuAction('download')">
          <el-icon><Download /></el-icon>
          <span>下载</span>
        </div>
        <div class="menu-item danger" @click="handleContextMenuAction('delete')">
          <el-icon><Delete /></el-icon>
          <span>删除</span>
        </div>
      </template>
    </div>

    <!-- 服务器右键菜单 -->
    <div 
      v-show="serverContextMenu.visible" 
      class="context-menu"
      :style="{ left: serverContextMenu.x + 'px', top: serverContextMenu.y + 'px' }"
    >
      <div class="menu-item" @click="handleServerContextMenuAction('rename')">
        <el-icon><Edit /></el-icon>
        <span>重命名</span>
      </div>
      <div class="menu-item danger" @click="handleServerContextMenuAction('delete')">
        <el-icon><Delete /></el-icon>
        <span>删除</span>
      </div>
    </div>

    <!-- 上传文件输入 -->
    <input 
      ref="uploadInput"
      type="file" 
      multiple 
      style="display: none" 
      @change="handleUploadChange"
    />
  </div>

  <!-- 重命名对话框 -->
  <el-dialog v-model="renameDialog.visible" title="重命名服务器" width="420px" :close-on-click-modal="false">
    <el-form label-position="top">
      <el-form-item label="新名称">
        <el-input v-model="renameDialog.name" placeholder="例如：生产服务器" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="renameDialog.visible = false">取消</el-button>
      <el-button type="primary" @click="confirmRename">确定</el-button>
    </template>
  </el-dialog>

  <!-- 确认对话框 -->
  <el-dialog 
    v-model="confirmDialog.visible" 
    :title="confirmDialog.title"
    width="420px"
    :close-on-click-modal="false"
    class="confirm-dialog"
  >
    <div class="confirm-content">
      <el-icon class="confirm-icon" :class="confirmDialog.type">
        <WarningFilled v-if="confirmDialog.type === 'danger'" />
        <QuestionFilled v-else />
      </el-icon>
      <p class="confirm-message">{{ confirmDialog.message }}</p>
    </div>
    <template #footer>
      <el-button @click="confirmDialog.visible = false">取消</el-button>
      <el-button 
        :type="confirmDialog.type === 'danger' ? 'danger' : 'primary'" 
        @click="confirmDialog.onConfirm"
      >
        确定
      </el-button>
    </template>
  </el-dialog>

  <!-- 输入对话框 -->
  <el-dialog
    v-model="inputDialog.visible"
    :title="inputDialog.title"
    width="420px"
    :close-on-click-modal="false"
  >
    <el-form-item :label="inputDialog.label">
      <el-input v-model="inputDialog.value" :placeholder="inputDialog.placeholder" @keydown.enter.prevent.stop="inputDialog.onConfirm" />
    </el-form-item>
    <template #footer>
      <el-button @click="inputDialog.visible = false">取消</el-button>
      <el-button type="primary" @click="inputDialog.onConfirm">确定</el-button>
    </template>
  </el-dialog>

  <!-- 连接服务器对话框 -->
  <el-dialog
    v-model="connectDialog.visible"
    title="连接服务器"
    width="480px"
    :close-on-click-modal="false"
    class="connect-dialog"
  >
    <el-form :model="connectDialog.form" label-position="top" size="default">
      <el-form-item label="连接协议">
        <el-select v-model="connectDialog.form.protocol" @change="updateConnectPort" placeholder="选择协议" style="width: 100%">
          <el-option label="SFTP (SSH)" value="sftp">
            <div class="select-option">
              <el-icon><Key /></el-icon>
              <span>SFTP (SSH)</span>
            </div>
          </el-option>
          <el-option label="FTP" value="ftp">
            <div class="select-option">
              <el-icon><Connection /></el-icon>
              <span>FTP</span>
            </div>
          </el-option>
        </el-select>
      </el-form-item>
      <el-row :gutter="16">
        <el-col :span="16">
          <el-form-item label="服务器地址">
            <el-input v-model="connectDialog.form.host" placeholder="192.168.1.1 或 example.com" />
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="端口">
            <el-input v-model.number="connectDialog.form.port" type="number" :placeholder="connectDialog.form.protocol === 'sftp' ? '22' : '21'" />
          </el-form-item>
        </el-col>
      </el-row>
      <el-form-item label="用户名">
        <el-input v-model="connectDialog.form.username" placeholder="登录用户名" />
      </el-form-item>
      <el-form-item label="密码">
        <el-input v-model="connectDialog.form.password" type="password" placeholder="登录密码" show-password />
      </el-form-item>
      <el-form-item label="服务器名称（可选，便于识别）">
        <el-input v-model="connectDialog.form.name" placeholder="例如：生产服务器" />
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="connect-dialog-footer">
        <el-button @click="handleSaveOnly" :disabled="!connectDialog.form.host">
          <el-icon><Star /></el-icon>
          仅保存
        </el-button>
        <div class="footer-right">
          <el-button @click="connectDialog.visible = false">取消</el-button>
          <el-button type="primary" :loading="connectDialog.loading" @click="handleConnectFromDialog">
            <el-icon><Connection /></el-icon>
            连接
          </el-button>
        </div>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { reactive, computed, ref, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { 
  WarningFilled, QuestionFilled, Key, Connection, Star, Monitor, Plus, 
  Folder, FolderOpened, FolderAdd, DocumentAdd, Document, Upload, Download,
  Delete, Close, ArrowLeft, HomeFilled, Refresh, Edit
} from '@element-plus/icons-vue'
import AppHeader from './components/AppHeader.vue'
import TerminalPanel from './components/TerminalPanel.vue'

const isDesktop = typeof window !== 'undefined' && window.go

const state = reactive({
  connections: {},
  activeConnId: null,
  savedConnections: [],
  activeTab: 'files',
})

const confirmDialog = reactive({
  visible: false,
  title: '',
  message: '',
  type: 'default',
  onConfirm: () => {},
})

const inputDialog = reactive({
  visible: false,
  title: '',
  label: '',
  value: '',
  placeholder: '',
  onConfirm: () => {},
})

const connectDialog = reactive({
  visible: false,
  loading: false,
  form: {
    protocol: 'sftp',
    host: '',
    port: 22,
    username: '',
    password: '',
    name: '',
  },
})

const fileContextMenu = reactive({
  visible: false,
  x: 0,
  y: 0,
  target: null,
})

const serverContextMenu = reactive({
  visible: false,
  x: 0,
  y: 0,
  target: null,
})

const uploadInput = ref()
const renameDialog = reactive({
  visible: false,
  id: '',
  name: '',
})

const hasActiveConnection = computed(() => !!state.activeConnId && !!state.connections[state.activeConnId])

const activeConnectionName = computed(() => {
  if (!state.activeConnId) return ''
  const conn = state.connections[state.activeConnId]
  return conn?.name || conn?.host || ''
})

const currentPath = computed(() => {
  if (!state.activeConnId) return '/'
  return state.connections[state.activeConnId]?.currentPath || '/'
})

const currentFiles = computed(() => {
  if (!state.activeConnId) return []
  return state.connections[state.activeConnId]?.files || []
})


async function callBackend(method, ...args) {
  if (!isDesktop) {
    ElMessage.warning('仅桌面端可用')
    return null
  }
  try {
    const fn = window.go?.main?.App?.[method]
    if (fn) return await fn(...args)
    console.warn(`Backend method not found: ${method}`)
    return null
  } catch (e) {
    console.error(`Backend error [${method}]:`, e)
    throw e
  }
}

async function loadSavedConnections() {
  try {
    const list = await callBackend('GetSavedConnections')
    state.savedConnections = Array.isArray(list) ? list : []
  } catch (e) {
    console.error('Load connections failed:', e)
    state.savedConnections = []
  }
}

async function loadFiles(connId) {
  if (!connId || !state.connections[connId]) return
  try {
    const conn = state.connections[connId]
    const result = await callBackend('List', connId, conn.currentPath)
    if (result?.error) {
      ElMessage.error('加载文件列表失败: ' + result.error)
      conn.files = []
    } else {
      conn.files = Array.isArray(result?.list) ? result.list : []
    }
  } catch (e) {
    ElMessage.error('加载文件列表失败: ' + e)
    if (state.connections[connId]) {
      state.connections[connId].files = []
    }
  }
}

async function loadActiveConnections() {
  try {
    const list = await callBackend('ListActiveConnections')
    if (Array.isArray(list) && list.length > 0) {
      for (const info of list) {
        if (!state.connections[info.id]) {
          state.connections[info.id] = {
            id: info.id,
            name: info.name,
            host: info.host,
            port: info.port,
            protocol: info.protocol,
            currentPath: '/',
            files: [],
            terminalLines: [],
          }
        }
      }
      if (!state.activeConnId && list.length > 0) {
        state.activeConnId = list[0].id
        await loadFiles(state.activeConnId)
      }
    }
  } catch (e) {
    console.error('Load active connections failed:', e)
  }
}

function formatSize(bytes) {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let i = 0
  while (bytes >= 1024 && i < units.length - 1) {
    bytes /= 1024
    i++
  }
  return bytes.toFixed(i > 0 ? 1 : 0) + ' ' + units[i]
}

function formatTime(isoTime) {
  if (!isoTime) return ''
  const d = new Date(isoTime)
  const pad = n => String(n).padStart(2, '0')
  return `${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function getFileType(name) {
  const ext = name.split('.').pop()?.toLowerCase()
  const types = {
    image: ['jpg', 'jpeg', 'png', 'gif', 'svg', 'webp'],
    code: ['js', 'ts', 'vue', 'jsx', 'tsx', 'go', 'py', 'java', 'php', 'html', 'css', 'scss', 'json'],
    doc: ['txt', 'md', 'pdf', 'doc', 'docx', 'xls', 'xlsx'],
    archive: ['zip', 'tar', 'gz', 'rar', '7z'],
    video: ['mp4', 'avi', 'mkv', 'mov'],
  }
  for (const [type, exts] of Object.entries(types)) {
    if (exts.includes(ext)) return type
  }
  return 'default'
}

function handleFileClick(file) {
  if (file.isDir) {
    handleNavigate(file.name)
  }
}

function handleFileDbClick(file) {
  if (file.isDir) {
    handleNavigate(file.name)
  } else {
    handleDownload(file)
  }
}

function showFileContextMenu(e, file) {
  const menuHeight = file ? 200 : 120
  const menuWidth = 160
  const viewportHeight = window.innerHeight
  const viewportWidth = window.innerWidth
  
  let x = e.clientX
  let y = e.clientY
  
  if (y + menuHeight > viewportHeight) {
    y = viewportHeight - menuHeight - 10
  }
  if (x + menuWidth > viewportWidth) {
    x = viewportWidth - menuWidth - 10
  }
  
  fileContextMenu.x = x
  fileContextMenu.y = y
  fileContextMenu.target = file
  fileContextMenu.visible = true
  serverContextMenu.visible = false
}

function showServerContextMenu(e, conn) {
  const menuHeight = 100
  const menuWidth = 160
  const viewportHeight = window.innerHeight
  const viewportWidth = window.innerWidth
  
  let x = e.clientX
  let y = e.clientY
  
  if (y + menuHeight > viewportHeight) {
    y = viewportHeight - menuHeight - 10
  }
  if (x + menuWidth > viewportWidth) {
    x = viewportWidth - menuWidth - 10
  }
  
  serverContextMenu.x = x
  serverContextMenu.y = y
  serverContextMenu.target = conn
  serverContextMenu.visible = true
  fileContextMenu.visible = false
}

function handleContextMenuAction(action) {
  fileContextMenu.visible = false
  switch (action) {
    case 'upload':
      triggerUpload()
      break
    case 'newFolder':
      handleNewFolder()
      break
    case 'newFile':
      handleNewFile()
      break
    case 'download':
      if (fileContextMenu.target) handleDownload(fileContextMenu.target)
      break
    case 'delete':
      if (fileContextMenu.target) handleDelete(fileContextMenu.target)
      break
  }
}

function handleServerContextMenuAction(action) {
  serverContextMenu.visible = false
  const conn = serverContextMenu.target
  if (!conn) return
  
  switch (action) {
    case 'rename':
      renameDialog.id = conn.id
      renameDialog.name = conn.name || ''
      renameDialog.visible = true
      break
    case 'delete':
      handleDeleteConnection(conn.id)
      break
  }
}

function triggerUpload() {
  uploadInput.value?.click()
}

function handleUploadChange(e) {
  const files = Array.from(e.target.files || [])
  if (files.length > 0) {
    handleUpload(files)
  }
  e.target.value = ''
}

function closeContextMenus() {
  fileContextMenu.visible = false
  serverContextMenu.visible = false
}

function handleTabChange(tab) {
  state.activeTab = tab
}

function handleSwitchConnection(connId) {
  if (state.connections[connId]) {
    state.activeConnId = connId
  }
}

async function handleCloseConnection(connId) {
  try {
    await callBackend('Disconnect', connId)
    delete state.connections[connId]
    if (state.activeConnId === connId) {
      const connIds = Object.keys(state.connections)
      state.activeConnId = connIds.length > 0 ? connIds[0] : null
    }
  } catch (e) {
    console.error('Disconnect error:', e)
  }
}

function updateConnectPort() {
  connectDialog.form.port = connectDialog.form.protocol === 'sftp' ? 22 : 21
}

function resetConnectForm() {
  connectDialog.form.protocol = 'sftp'
  connectDialog.form.host = ''
  connectDialog.form.port = 22
  connectDialog.form.username = ''
  connectDialog.form.password = ''
  connectDialog.form.name = ''
}

async function handleConnectFromDialog() {
  const form = connectDialog.form
  if (!form.host) {
    ElMessage.warning('请填写服务器地址')
    return
  }
  
  connectDialog.loading = true
  try {
    const result = await callBackend('Connect', {
      protocol: form.protocol,
      host: form.host.trim(),
      port: form.port || (form.protocol === 'sftp' ? 22 : 21),
      username: form.username.trim(),
      password: form.password,
    })
    if (result?.error) {
      ElMessage.error('连接失败: ' + result.error)
      return
    }
    const connId = result.connId
    state.connections[connId] = {
      id: connId,
      name: form.name || result.name || form.host,
      host: result.host || form.host,
      port: result.port || form.port,
      protocol: form.protocol,
      currentPath: '/',
      files: [],
      terminalLines: [],
    }
    state.activeConnId = connId
    
    // 如果填写了名称，自动保存
    if (form.name) {
      await handleSaveConnection({
        name: form.name,
        protocol: form.protocol,
        host: form.host.trim(),
        port: form.port,
        username: form.username.trim(),
        password: form.password,
      })
    }
    
    connectDialog.visible = false
    resetConnectForm()
    await loadFiles(connId)
  } catch (e) {
    ElMessage.error('连接失败: ' + e)
  } finally {
    connectDialog.loading = false
  }
}

async function handleSaveOnly() {
  const form = connectDialog.form
  if (!form.host) {
    ElMessage.warning('请填写服务器地址')
    return
  }
  await handleSaveConnection({
    name: form.name || form.host,
    protocol: form.protocol,
    host: form.host.trim(),
    port: form.port || (form.protocol === 'sftp' ? 22 : 21),
    username: form.username.trim(),
    password: form.password,
  })
  connectDialog.visible = false
  resetConnectForm()
}

async function handleConnect(info) {
  try {
    const result = await callBackend('Connect', {
      protocol: info.protocol,
      host: info.host,
      port: info.port,
      username: info.username,
      password: info.password,
    })
    if (result?.error) {
      ElMessage.error('连接失败: ' + result.error)
      return
    }
    const connId = result.connId
    state.connections[connId] = {
      id: connId,
      name: result.name || info.host,
      host: result.host || info.host,
      port: result.port || info.port,
      protocol: info.protocol,
      currentPath: '/',
      files: [],
      terminalLines: [],
    }
    state.activeConnId = connId
    await loadFiles(connId)
  } catch (e) {
    ElMessage.error('连接失败: ' + e)
  }
}

async function handleConnectSaved(savedId) {
  try {
    const result = await callBackend('ConnectSaved', savedId)
    if (result?.error) {
      ElMessage.error('连接失败: ' + result.error)
      return
    }
    const connId = result.connId
    state.connections[connId] = {
      id: connId,
      name: result.name || result.host,
      host: result.host,
      port: result.port,
      protocol: 'sftp',
      currentPath: '/',
      files: [],
      terminalLines: [],
    }
    state.activeConnId = connId
    await loadFiles(connId)
  } catch (e) {
    ElMessage.error('连接失败: ' + e)
  }
}

async function handleDisconnectActive() {
  if (state.activeConnId) {
    await handleCloseConnection(state.activeConnId)
  }
}

async function handleSaveConnection(info) {
  try {
    await callBackend('SaveConnection', '', info.name, info.host, info.port, info.username, info.password, info.protocol)
    ElMessage.success('保存成功')
    await loadSavedConnections()
  } catch (e) {
    ElMessage.error('保存失败: ' + e)
  }
}

async function handleDeleteConnection(id) {
  confirmDialog.title = '删除服务器'
  confirmDialog.message = '确定要删除这个保存的服务器吗？此操作不可恢复。'
  confirmDialog.type = 'danger'
  confirmDialog.visible = true
  confirmDialog.onConfirm = async () => {
    confirmDialog.visible = false
    try {
      await callBackend('DeleteConnection', id)
      ElMessage.success('已删除')
      await loadSavedConnections()
    } catch (e) {
      ElMessage.error('删除失败: ' + e)
    }
  }
}

async function handleRenameConnection({ id, name }) {
  try {
    await callBackend('RenameConnection', id, name)
    ElMessage.success('重命名成功')
    await loadSavedConnections()
  } catch (e) {
    ElMessage.error('重命名失败: ' + e)
  }
}

async function confirmRename() {
  if (!renameDialog.name.trim()) {
    ElMessage.warning('请输入名称')
    return
  }
  await handleRenameConnection({ id: renameDialog.id, name: renameDialog.name.trim() })
  renameDialog.visible = false
}

function handleGoUp() {
  if (!state.activeConnId) return
  const conn = state.connections[state.activeConnId]
  if (!conn || conn.currentPath === '/') return
  const parts = conn.currentPath.split('/').filter(Boolean)
  parts.pop()
  conn.currentPath = '/' + parts.join('/')
  loadFiles(state.activeConnId)
}

function handleGoHome() {
  if (!state.activeConnId) return
  const conn = state.connections[state.activeConnId]
  if (conn) {
    conn.currentPath = '/'
    loadFiles(state.activeConnId)
  }
}

function handleRefresh() {
  if (state.activeConnId) {
    loadFiles(state.activeConnId)
  }
}

async function handleNavigate(name) {
  if (!state.activeConnId) return
  const conn = state.connections[state.activeConnId]
  if (!conn) return
  const newPath = conn.currentPath === '/' ? '/' + name : conn.currentPath + '/' + name
  conn.currentPath = newPath
  await loadFiles(state.activeConnId)
}

async function handleNewFolder() {
  if (!state.activeConnId) return
  inputDialog.title = '新建文件夹'
  inputDialog.label = '文件夹名称'
  inputDialog.value = ''
  inputDialog.placeholder = '请输入文件夹名称'
  inputDialog.visible = true
  inputDialog.onConfirm = async () => {
    if (!inputDialog.value.trim()) {
      ElMessage.warning('请输入名称')
      return
    }
    inputDialog.visible = false
    try {
      const conn = state.connections[state.activeConnId]
      const path = conn.currentPath === '/' ? '/' + inputDialog.value : conn.currentPath + '/' + inputDialog.value
      await callBackend('CreateDir', state.activeConnId, path)
      ElMessage.success('创建成功')
      await loadFiles(state.activeConnId)
    } catch (e) {
      ElMessage.error('创建失败: ' + e)
    }
  }
}

async function handleNewFile() {
  if (!state.activeConnId) return
  inputDialog.title = '新建文件'
  inputDialog.label = '文件名称'
  inputDialog.value = ''
  inputDialog.placeholder = '请输入文件名称'
  inputDialog.visible = true
  inputDialog.onConfirm = async () => {
    if (!inputDialog.value.trim()) {
      ElMessage.warning('请输入名称')
      return
    }
    inputDialog.visible = false
    try {
      const conn = state.connections[state.activeConnId]
      const path = conn.currentPath === '/' ? '/' + inputDialog.value : conn.currentPath + '/' + inputDialog.value
      await callBackend('CreateFile', state.activeConnId, path)
      ElMessage.success('创建成功')
      await loadFiles(state.activeConnId)
    } catch (e) {
      ElMessage.error('创建失败: ' + e)
    }
  }
}

async function handleUpload(files) {
  if (!state.activeConnId) return
  const conn = state.connections[state.activeConnId]
  for (const file of files) {
    try {
      const remotePath = conn.currentPath === '/' ? '/' + file.name : conn.currentPath + '/' + file.name
      const base64 = await fileToBase64(file)
      await callBackend('UploadBase64', state.activeConnId, remotePath, base64)
      ElMessage.success(`上传成功: ${file.name}`)
    } catch (e) {
      ElMessage.error(`上传失败: ${file.name} - ${e}`)
    }
  }
  await loadFiles(state.activeConnId)
}

function fileToBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      const base64 = reader.result.split(',')[1]
      resolve(base64)
    }
    reader.onerror = reject
    reader.readAsDataURL(file)
  })
}

async function handleDownload(file) {
  if (!state.activeConnId) return
  try {
    const conn = state.connections[state.activeConnId]
    const remotePath = conn.currentPath === '/' ? '/' + file.name : conn.currentPath + '/' + file.name
    const localPath = await callBackend('SaveFileDialog', file.name)
    if (localPath) {
      await callBackend('DownloadToPath', state.activeConnId, remotePath, localPath)
      ElMessage.success('下载成功')
    }
  } catch (e) {
    ElMessage.error('下载失败: ' + e)
  }
}

async function handleDelete(file) {
  if (!state.activeConnId) return
  confirmDialog.title = '删除确认'
  confirmDialog.message = `确定要删除 "${file.name}" 吗？此操作不可恢复。`
  confirmDialog.type = 'danger'
  confirmDialog.visible = true
  confirmDialog.onConfirm = async () => {
    confirmDialog.visible = false
    try {
      const conn = state.connections[state.activeConnId]
      const remotePath = conn.currentPath === '/' ? '/' + file.name : conn.currentPath + '/' + file.name
      if (file.isDir) {
        await callBackend('DeleteDirRecursive', state.activeConnId, remotePath)
      } else {
        await callBackend('DeleteFile', state.activeConnId, remotePath)
      }
      ElMessage.success('删除成功')
      await loadFiles(state.activeConnId)
    } catch (e) {
      ElMessage.error('删除失败: ' + e)
    }
  }
}

async function handleDeleteSelected(items) {
  if (!state.activeConnId || items.length === 0) return
  confirmDialog.title = '批量删除'
  confirmDialog.message = `确定要删除选中的 ${items.length} 个项目吗？此操作不可恢复。`
  confirmDialog.type = 'danger'
  confirmDialog.visible = true
  confirmDialog.onConfirm = async () => {
    confirmDialog.visible = false
    const conn = state.connections[state.activeConnId]
    for (const item of items) {
      try {
        const remotePath = conn.currentPath === '/' ? '/' + item.name : conn.currentPath + '/' + item.name
        if (item.isDir) {
          await callBackend('DeleteDirRecursive', state.activeConnId, remotePath)
        } else {
          await callBackend('DeleteFile', state.activeConnId, remotePath)
        }
      } catch (e) {
        ElMessage.error(`删除失败: ${item.name} - ${e}`)
      }
    }
    ElMessage.success('批量删除完成')
    await loadFiles(state.activeConnId)
  }
}

onMounted(() => {
  loadSavedConnections()
  loadActiveConnections()
  document.addEventListener('click', closeContextMenus)
})

onUnmounted(() => {
  document.removeEventListener('click', closeContextMenus)
})
</script>

<style lang="scss">
@use './styles/main.scss' as *;

.app {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}

.app-body {
  flex: 1;
  display: flex;
  overflow: hidden;
  min-height: 0;
}

.app-footer {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 6px 16px;
  background: var(--bg-secondary);
  border-top: 1px solid var(--border);
  font-size: 11px;
  color: var(--text-muted);
  
  .separator {
    opacity: 0.5;
  }
}

.left-panel {
  width: 380px;
  min-width: 320px;
  max-width: 480px;
  display: flex;
  flex-direction: column;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border);
  overflow: hidden;
  min-height: 0;
}

.server-section {
  height: 30%;
  min-height: 120px;
  display: flex;
  flex-direction: column;
  border-bottom: 1px solid var(--border);
  overflow: hidden;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: var(--bg-tertiary);
  border-bottom: 1px solid var(--border);
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  
  .el-icon {
    color: var(--accent);
  }
}

.active-conn-name {
  font-size: 12px;
  color: var(--accent);
  background: var(--accent-light);
  padding: 4px 10px;
  border-radius: var(--radius-sm);
}

.server-list {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 8px;
  
  &::-webkit-scrollbar {
    width: 6px;
  }
  
  &::-webkit-scrollbar-track {
    background: transparent;
  }
  
  &::-webkit-scrollbar-thumb {
    background: var(--border);
    border-radius: 3px;
    
    &:hover {
      background: var(--text-muted);
    }
  }
}

.server-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border-radius: var(--radius);
  cursor: pointer;
  transition: all 0.15s;
  
  &:hover {
    background: var(--bg-hover);
    
    .server-icon {
      background: var(--accent);
      color: white;
    }
  }
}

.server-icon {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-tertiary);
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  transition: all 0.15s;
  flex-shrink: 0;
  font-size: 14px;
}

.server-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.server-name {
  font-size: 13px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.server-meta {
  font-size: 11px;
  color: var(--text-muted);
}

.server-empty {
  padding: 20px;
  text-align: center;
  color: var(--text-muted);
  font-size: 13px;
  
  p { margin: 0; }
}

.file-section {
  height: 70%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
}

.file-toolbar {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 8px 12px;
  border-bottom: 1px solid var(--border);
  background: var(--bg-tertiary);
}

.toolbar-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  padding: 0;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s ease;
  
  &:hover {
    background: var(--bg-hover);
    border-color: var(--border-light);
    color: var(--text-primary);
  }
  
  &:active {
    transform: scale(0.95);
  }
}

.path-display {
  flex: 1;
  padding: 0 12px;
  font-size: 13px;
  font-family: var(--font-mono);
  color: var(--text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.toolbar-divider {
  width: 1px;
  height: 24px;
  background: var(--border);
  margin: 0 8px;
}

.file-list {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 8px;
  
  &::-webkit-scrollbar {
    width: 8px;
  }
  
  &::-webkit-scrollbar-track {
    background: transparent;
  }
  
  &::-webkit-scrollbar-thumb {
    background: var(--border);
    border-radius: 4px;
    
    &:hover {
      background: var(--text-muted);
    }
  }
}

.file-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border-radius: var(--radius);
  cursor: pointer;
  transition: all 0.15s;
  
  &:hover {
    background: var(--bg-hover);
  }
}

.file-icon {
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  font-size: 15px;
  flex-shrink: 0;
  
  &.is-folder {
    background: rgba(251, 191, 36, 0.15);
    color: #fbbf24;
  }
  
  &.image {
    background: rgba(168, 85, 247, 0.15);
    color: #a855f7;
  }
  
  &.code {
    background: rgba(63, 185, 80, 0.15);
    color: #3fb950;
  }
  
  &.doc {
    background: rgba(88, 166, 255, 0.15);
    color: #58a6ff;
  }
  
  &.archive {
    background: rgba(234, 88, 12, 0.15);
    color: #ea580c;
  }
  
  &.video {
    background: rgba(236, 72, 153, 0.15);
    color: #ec4899;
  }
  
  &.default {
    background: var(--bg-tertiary);
    color: var(--text-secondary);
  }
}

.file-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.file-name {
  font-size: 13px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.file-meta {
  font-size: 11px;
  color: var(--text-muted);
}

.file-empty, .file-empty-list {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 40px 20px;
  color: var(--text-muted);
  flex: 1;
  
  .el-icon {
    opacity: 0.4;
  }
  
  p {
    margin: 0;
    font-size: 14px;
  }
}

.right-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: var(--bg-primary);
  overflow: hidden;
  min-height: 0;
}

.context-menu {
  position: fixed;
  z-index: 9999;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 6px 0;
  min-width: 160px;
  box-shadow: var(--shadow-lg);
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  cursor: pointer;
  color: var(--text-primary);
  font-size: 13px;
  transition: background 0.15s;
  
  &:hover {
    background: var(--bg-hover);
  }
  
  &.danger {
    color: var(--danger);
    
    &:hover {
      background: var(--danger-light);
    }
  }
}

.menu-divider {
  height: 1px;
  background: var(--border);
  margin: 6px 0;
}

.confirm-dialog {
  .confirm-content {
    display: flex;
    align-items: flex-start;
    gap: 16px;
    padding: 8px 0;
  }
  
  .confirm-icon {
    font-size: 24px;
    
    &.danger {
      color: var(--danger);
    }
  }
  
  .confirm-message {
    margin: 0;
    line-height: 1.6;
    color: var(--text-secondary);
    padding-top: 2px;
  }
}

.connect-dialog {
  .select-option {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  
  .connect-dialog-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  
  .footer-right {
    display: flex;
    gap: 12px;
  }
}
</style>
