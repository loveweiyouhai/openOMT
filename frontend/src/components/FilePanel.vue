<template>
  <div class="file-panel" @contextmenu.prevent="showContextMenu">
    <el-table
      :data="fileList"
      style="width: 100%"
      :row-class-name="rowClassName"
      @row-click="handleRowClick"
      @selection-change="handleSelectionChange"
      ref="tableRef"
      empty-text="此目录为空"
    >
      <el-table-column type="selection" width="50" />
      <el-table-column label="名称" min-width="200">
        <template #default="{ row }">
          <div class="file-name">
            <el-icon v-if="row.isDir" color="#f0c674"><Folder /></el-icon>
            <el-icon v-else color="#8b949e"><Document /></el-icon>
            <span>{{ row.name }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="大小" width="120">
        <template #default="{ row }">
          {{ row.isDir ? '-' : formatSize(row.size) }}
        </template>
      </el-table-column>
      <el-table-column prop="modTime" label="修改时间" width="180" />
      <el-table-column label="操作" width="150">
        <template #default="{ row }">
          <el-button
            v-if="!row.isDir"
            size="small"
            type="primary"
            link
            @click.stop="$emit('download', row.path)"
          >
            下载
          </el-button>
          <el-button
            size="small"
            type="danger"
            link
            @click.stop="$emit('delete', { path: row.path, isDir: row.isDir })"
          >
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <div v-if="!connected" class="empty-state">
      <el-icon :size="48"><FolderOpened /></el-icon>
      <p>请先连接服务器并选择路径</p>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { Folder, Document, FolderOpened } from '@element-plus/icons-vue'

const props = defineProps({
  fileList: Array,
  currentPath: String,
  connected: Boolean,
  selected: Array,
})

const emit = defineEmits(['navigate', 'download', 'delete', 'create-folder', 'create-file', 'upload', 'refresh', 'update:selected'])

const tableRef = ref(null)

function formatSize(bytes) {
  if (!bytes) return '-'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  let n = bytes
  while (n >= 1024 && i < units.length - 1) {
    n /= 1024
    i++
  }
  return n.toFixed(i === 0 ? 0 : 1) + ' ' + units[i]
}

function rowClassName({ row }) {
  return row.isDir ? 'row-dir' : 'row-file'
}

function handleRowClick(row, column, event) {
  if (column.type === 'selection') return
  if (row.isDir) {
    emit('navigate', row.path)
  }
}

function handleSelectionChange(selection) {
  emit('update:selected', selection.map(r => ({ path: r.path, isDir: r.isDir })))
}

function showContextMenu(e) {
  const items = [
    { label: '刷新', onClick: () => emit('refresh') },
    { label: '上传文件', onClick: () => emit('upload') },
    { label: '新建文件夹', onClick: () => {
      const name = prompt('请输入文件夹名称：')
      if (name) emit('create-folder', name.trim())
    }},
    { label: '新建文件', onClick: () => {
      const name = prompt('请输入文件名：')
      if (name) emit('create-file', name.trim())
    }},
  ]
  
  const menu = document.createElement('div')
  menu.style.cssText = `position:fixed;left:${e.clientX}px;top:${e.clientY}px;z-index:9999;background:var(--bg-secondary);border:1px solid var(--border);border-radius:6px;padding:4px 0;min-width:120px;`
  items.forEach(item => {
    const btn = document.createElement('button')
    btn.textContent = item.label
    btn.style.cssText = 'display:block;width:100%;padding:8px 16px;text-align:left;background:none;border:none;color:var(--text-primary);cursor:pointer;font-size:14px;'
    btn.onmouseover = () => btn.style.background = 'var(--bg-tertiary)'
    btn.onmouseout = () => btn.style.background = 'none'
    btn.onclick = () => {
      document.body.removeChild(menu)
      item.onClick()
    }
    menu.appendChild(btn)
  })
  document.body.appendChild(menu)
  const closeMenu = () => {
    if (document.body.contains(menu)) document.body.removeChild(menu)
    document.removeEventListener('click', closeMenu)
  }
  setTimeout(() => document.addEventListener('click', closeMenu), 0)
}
</script>

<style lang="scss" scoped>
.file-panel {
  flex: 1;
  overflow: auto;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  position: relative;
}

.file-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.empty-state {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
  color: var(--text-secondary);

  p {
    margin-top: 12px;
  }
}

:deep(.row-dir) {
  cursor: pointer;
}

:deep(.el-table__empty-block) {
  min-height: 200px;
}
</style>
