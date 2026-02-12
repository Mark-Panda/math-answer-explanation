<script setup lang="ts">
import { computed } from 'vue'
import katex from 'katex'
import 'katex/dist/katex.min.css'

const props = withDefaults(
  defineProps<{
    content: string
    displayMode?: boolean
  }>(),
  { displayMode: false }
)

// 将 content 中 $$...$$ 与 $...$ 转为 KaTeX 渲染后的 HTML，其余转义显示
const rendered = computed(() => {
  const s = props.content
  if (!s || !s.trim()) return ''
  const parts: string[] = []
  let i = 0
  const len = s.length
  while (i < len) {
    if (s.slice(i, i + 2) === '$$') {
      const end = s.indexOf('$$', i + 2)
      if (end === -1) {
        parts.push(escapeHtml(s.slice(i)))
        break
      }
      const math = s.slice(i + 2, end).trim()
      try {
        parts.push(katex.renderToString(math, { displayMode: true, throwOnError: false }))
      } catch {
        parts.push(escapeHtml(s.slice(i, end + 2)))
      }
      i = end + 2
      continue
    }
    if (s[i] === '$' && (i === 0 || s[i - 1] !== '\\')) {
      const end = s.indexOf('$', i + 1)
      if (end !== -1 && end > i + 1) {
        const math = s.slice(i + 1, end).trim()
        try {
          parts.push(katex.renderToString(math, { displayMode: false, throwOnError: false }))
        } catch {
          parts.push(escapeHtml(s[i] + math + '$'))
        }
        i = end + 1
        continue
      }
    }
    const nextDollar = s.indexOf('$', i)
    const nextDD = s.indexOf('$$', i)
    let next = len
    if (nextDollar !== -1) next = Math.min(next, nextDollar)
    if (nextDD !== -1) next = Math.min(next, nextDD)
    parts.push(escapeHtml(s.slice(i, next)))
    i = next
  }
  return parts.join('')
})

function escapeHtml(text: string): string {
  const div = document.createElement('div')
  div.textContent = text
  return div.innerHTML
}
</script>

<template>
  <div class="katex-render" :class="{ 'katex-display': displayMode }" v-html="rendered"></div>
</template>

<style scoped>
.katex-render {
  line-height: 1.6;
  word-break: break-word;
}
.katex-render :deep(.katex-display) {
  margin: 0.5em 0;
  overflow-x: auto;
}
.katex-render :deep(.katex) {
  font-size: 1.1em;
}
</style>
