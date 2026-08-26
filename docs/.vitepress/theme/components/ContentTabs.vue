<script setup lang="ts">
import { nextTick, onMounted, ref, useId, watch } from 'vue'

interface Tab {
  id: string
  label: string
}

const props = withDefaults(
  defineProps<{
    tabs: Tab[]
    ariaLabel?: string
    detectOs?: boolean
  }>(),
  {
    ariaLabel: 'Content tabs',
    detectOs: false
  }
)

const instanceId = useId().replaceAll(':', '')
const activeTab = ref(props.tabs[0]?.id ?? '')
const tabButtons = ref<HTMLButtonElement[]>([])

watch(
  () => props.tabs,
  (tabs) => {
    if (!tabs.some((tab) => tab.id === activeTab.value)) {
      activeTab.value = tabs[0]?.id ?? ''
    }
  }
)

function tabId(id: string) {
  return `${instanceId}-tab-${id}`
}

function panelId(id: string) {
  return `${instanceId}-panel-${id}`
}

function detectOperatingSystem() {
  const browserNavigator = navigator as Navigator & {
    userAgentData?: { platform?: string }
  }
  const platform = browserNavigator.userAgentData?.platform ?? navigator.platform
  const signature = `${platform} ${navigator.userAgent}`.toLowerCase()

  if (
    /android|iphone|ipad|ipod/.test(signature) ||
    (signature.includes('mac') && navigator.maxTouchPoints > 1)
  ) {
    return undefined
  }

  if (signature.includes('win')) return 'windows'
  if (signature.includes('mac')) return 'macos'
  if (/linux|x11/.test(signature)) return 'linux'

  return undefined
}

function activate(index: number, moveFocus = false) {
  const tab = props.tabs[index]

  if (!tab) return

  activeTab.value = tab.id

  if (moveFocus) {
    nextTick(() => tabButtons.value[index]?.focus())
  }
}

function handleKeydown(event: KeyboardEvent, index: number) {
  let nextIndex: number | undefined

  switch (event.key) {
    case 'ArrowRight':
      nextIndex = (index + 1) % props.tabs.length
      break
    case 'ArrowLeft':
      nextIndex = (index - 1 + props.tabs.length) % props.tabs.length
      break
    case 'Home':
      nextIndex = 0
      break
    case 'End':
      nextIndex = props.tabs.length - 1
      break
  }

  if (nextIndex === undefined) return

  event.preventDefault()
  activate(nextIndex, true)
}

onMounted(() => {
  if (!props.detectOs) return

  const operatingSystem = detectOperatingSystem()
  const index = props.tabs.findIndex((tab) => tab.id === operatingSystem)

  if (index >= 0) activate(index)
})
</script>

<template>
  <div v-if="tabs.length" class="content-tabs">
    <div class="content-tabs__list" role="tablist" :aria-label="ariaLabel">
      <button
        v-for="(tab, index) in tabs"
        :id="tabId(tab.id)"
        ref="tabButtons"
        :key="tab.id"
        class="content-tabs__tab"
        :class="{ 'is-active': activeTab === tab.id }"
        type="button"
        role="tab"
        :aria-controls="panelId(tab.id)"
        :aria-selected="activeTab === tab.id"
        :tabindex="activeTab === tab.id ? 0 : -1"
        @click="activate(index)"
        @keydown="handleKeydown($event, index)"
      >
        {{ tab.label }}
      </button>
    </div>

    <div
      v-for="tab in tabs"
      v-show="activeTab === tab.id"
      :id="panelId(tab.id)"
      :key="tab.id"
      class="content-tabs__panel"
      role="tabpanel"
      :aria-labelledby="tabId(tab.id)"
      tabindex="0"
    >
      <slot :name="tab.id" />
    </div>
  </div>
</template>

<style scoped>
.content-tabs {
  margin: 16px 0;
}

.content-tabs__list {
  display: flex;
  overflow-x: auto;
  background: var(--vp-c-bg-alt);
}

.content-tabs__tab {
  flex: 0 0 auto;
  border: 0;
  border-bottom: 2px solid transparent;
  padding: 10px 16px 8px;
  color: var(--vp-c-text-2);
  background: transparent;
  font: inherit;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: color 0.2s, border-color 0.2s, background-color 0.2s;
}

.content-tabs__tab:hover {
  color: var(--vp-c-text-1);
  background: var(--vp-c-bg-soft);
}

.content-tabs__tab:focus-visible {
  outline: 2px solid var(--vp-c-brand-1);
  outline-offset: -2px;
}

.content-tabs__tab.is-active {
  border-bottom-color: var(--vp-c-brand-1);
  color: var(--vp-c-brand-1);
}

.content-tabs__panel {
  padding: 20px 0 0;
  outline: none;
}

.content-tabs__panel:focus-visible {
  outline: 2px solid var(--vp-c-brand-1);
  outline-offset: 4px;
}

.content-tabs__panel :deep(> :first-child) {
  margin-top: 0;
}

.content-tabs__panel :deep(> :last-child) {
  margin-bottom: 0;
}

@media (max-width: 639px) {
  .content-tabs__tab {
    padding-right: 13px;
    padding-left: 13px;
  }

  .content-tabs__panel {
    padding-top: 16px;
  }
}
</style>
