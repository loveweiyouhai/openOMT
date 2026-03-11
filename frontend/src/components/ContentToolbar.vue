<template>
  <div class="toolbar">
    <div class="toolbar-tabs">
      <el-button
        :type="activeTab === 'files' ? 'primary' : 'default'"
        @click="$emit('tab-change', 'files')"
      >
        文件
      </el-button>
      <el-button
        :type="activeTab === 'terminal' ? 'primary' : 'default'"
        @click="$emit('tab-change', 'terminal')"
      >
        终端
      </el-button>
    </div>
    <div class="toolbar-right">
      <el-button size="small" :disabled="atRoot || !connected" @click="$emit('go-root')">
        根目录
      </el-button>
      <el-button size="small" :disabled="atRoot || !connected" @click="$emit('go-up')">
        ↑ 上一级
      </el-button>
      <span class="path-display">当前目录：{{ currentPath || '/' }}</span>
      <el-button
        size="small"
        type="danger"
        plain
        :disabled="selectedCount === 0"
        @click="$emit('delete-selected')"
      >
        删除所选 ({{ selectedCount }})
      </el-button>
      <el-button size="small" :disabled="!connected" @click="$emit('refresh')">
        刷新
      </el-button>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  currentPath: String,
  connected: Boolean,
  activeTab: String,
  selectedCount: Number,
})

defineEmits(['tab-change', 'go-root', 'go-up', 'refresh', 'delete-selected'])

const atRoot = computed(() => {
  return !props.currentPath || props.currentPath === '/' || props.currentPath === ''
})
</script>

<style lang="scss" scoped>
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  flex-wrap: wrap;
  gap: 12px;
}

.toolbar-tabs {
  display: flex;
  gap: 8px;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.path-display {
  font-family: var(--font-mono);
  font-size: 0.85rem;
  color: var(--text-secondary);
  padding: 0 8px;
}
</style>
