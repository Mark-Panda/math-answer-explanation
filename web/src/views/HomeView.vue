<script setup lang="ts">
import { ref, computed } from 'vue'
import { uploadImage, submitImage, startExplain, startExplainFromImage, getResult } from '@/api/client'
import KaTeXRender from '@/components/KaTeXRender.vue'

const mode = ref<'upload' | 'text'>('text')
const problemText = ref('')
const uploadPath = ref('')
const uploadError = ref('')
const recognizeLoading = ref(false)
const explainLoading = ref(false)
const explainError = ref('')
const taskId = ref('')
const result = ref<{ steps: { title: string; content: string; image_url?: string }[] } | null>(null)
const showLightbox = ref(false)
const lightboxImage = ref('')

const canStartExplain = computed(() => problemText.value.trim().length > 0)
const canStartExplainFromImage = computed(() => uploadPath.value.length > 0)

async function onFileSelect(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  uploadError.value = ''
  recognizeLoading.value = true
  try {
    const { path } = await uploadImage(file)
    uploadPath.value = path
    const { problem_text } = await submitImage(path)
    problemText.value = problem_text
  } catch (err) {
    uploadError.value = err instanceof Error ? err.message : '上传或识图失败'
  } finally {
    recognizeLoading.value = false
    input.value = ''
  }
}

async function onStartExplain() {
  if (!canStartExplain.value) return
  explainError.value = ''
  explainLoading.value = true
  result.value = null
  taskId.value = ''
  try {
    const { task_id } = await startExplain(problemText.value)
    taskId.value = task_id
    await pollResult(task_id)
  } catch (err) {
    explainError.value = err instanceof Error ? err.message : '解析失败'
  } finally {
    explainLoading.value = false
  }
}

async function onStartExplainFromImage() {
  if (!canStartExplainFromImage.value) return
  explainError.value = ''
  explainLoading.value = true
  result.value = null
  taskId.value = ''
  try {
    const { task_id } = await startExplainFromImage(uploadPath.value)
    taskId.value = task_id
    await pollResult(task_id)
  } catch (err) {
    explainError.value = err instanceof Error ? err.message : '解析失败'
  } finally {
    explainLoading.value = false
  }
}

async function pollResult(id: string) {
  const data = await getResult(id)
  result.value = data
}

function openLightbox(url: string) {
  lightboxImage.value = url
  showLightbox.value = true
}
function closeLightbox() {
  showLightbox.value = false
}

function switchToText() {
  mode.value = 'text'
  uploadError.value = ''
  uploadPath.value = ''
}
</script>

<template>
  <div class="home">
    <section class="input-section">
      <h2>题目输入</h2>
      <div class="tabs">
        <button
          type="button"
          :class="{ active: mode === 'upload' }"
          @click="mode = 'upload'"
        >
          上传题目图片
        </button>
        <button
          type="button"
          :class="{ active: mode === 'text' }"
          @click="mode = 'text'; switchToText()"
        >
          输入题目文字
        </button>
      </div>

      <div v-if="mode === 'upload'" class="upload-area">
        <input
          type="file"
          accept="image/jpeg,image/png,image/webp"
          :disabled="recognizeLoading"
          @change="onFileSelect"
        />
        <p v-if="recognizeLoading" class="loading">正在识别题目…</p>
        <p v-else-if="uploadError" class="error">
          {{ uploadError }}
          <button type="button" class="link" @click="switchToText">改为输入文字</button>
        </p>
        <p v-else-if="uploadPath" class="success">
          已上传，识别结果如下可编辑；或
          <button
            type="button"
            class="link primary-inline"
            :disabled="explainLoading"
            @click="onStartExplainFromImage"
          >
            {{ explainLoading ? '解析中…' : '直接解析' }}
          </button>
          让模型看图解析（不经过 OCR）。
        </p>
      </div>

      <div class="text-area">
        <textarea
          v-model="problemText"
          placeholder="输入或粘贴题目内容（上传图片后此处为识别结果，可编辑）"
          rows="5"
        />
        <button
          type="button"
          class="primary"
          :disabled="!canStartExplain || explainLoading"
          @click="onStartExplain"
        >
          {{ explainLoading ? '解析中…' : '开始解析' }}
        </button>
      </div>
      <p v-if="explainError" class="error">{{ explainError }} 可修改题目后重试。</p>
    </section>

    <section v-if="explainLoading && !result" class="loading-section">
      <p>正在生成步骤解析与配图，请稍候…</p>
    </section>

    <section v-if="result?.steps?.length" class="result-section">
      <h2>解析结果</h2>
      <nav class="step-nav">
        <a
          v-for="(step, i) in result.steps"
          :key="i"
          :href="`#step-${i}`"
          class="step-link"
        >
          {{ step.title || `步骤 ${i + 1}` }}
        </a>
      </nav>
      <div
        v-for="(step, i) in result.steps"
        :key="i"
        :id="`step-${i}`"
        class="step-block"
      >
        <h3>{{ step.title }}</h3>
        <div class="step-content">
          <KaTeXRender :content="step.content" />
        </div>
        <div v-if="step.image_url" class="step-image">
          <img
            :src="step.image_url"
            :alt="step.title"
            loading="lazy"
            @click="openLightbox(step.image_url!)"
          />
        </div>
        <p v-else class="no-image">本步无配图</p>
      </div>
    </section>

    <div v-if="showLightbox" class="lightbox" @click.self="closeLightbox">
      <button type="button" class="close" aria-label="关闭" @click="closeLightbox">×</button>
      <img :src="lightboxImage" alt="放大" />
    </div>
  </div>
</template>

<style scoped>
.home {
  max-width: 800px;
}
.input-section h2,
.result-section h2 {
  margin-top: 0;
  font-size: 1.25rem;
}
.tabs {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 1rem;
}
.tabs button {
  padding: 0.5rem 1rem;
  border: 1px solid #ccc;
  background: #fff;
  cursor: pointer;
  border-radius: 4px;
}
.tabs button.active {
  background: #333;
  color: #fff;
  border-color: #333;
}
.upload-area {
  margin-bottom: 1rem;
}
.upload-area input[type='file'] {
  margin-bottom: 0.5rem;
}
.text-area {
  margin-bottom: 1rem;
}
.text-area textarea {
  width: 100%;
  box-sizing: border-box;
  padding: 0.5rem;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-family: inherit;
  margin-bottom: 0.5rem;
}
.primary {
  padding: 0.5rem 1rem;
  background: #333;
  color: #fff;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}
.primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.loading,
.success {
  color: #666;
  font-size: 0.9rem;
}
.error {
  color: #c00;
  font-size: 0.9rem;
}
.link {
  background: none;
  border: none;
  color: #06c;
  cursor: pointer;
  text-decoration: underline;
  margin-left: 0.25rem;
}
.link.primary-inline {
  color: #333;
  font-weight: 500;
  text-decoration: none;
  padding: 0.2rem 0.4rem;
  border-radius: 4px;
  background: #e8e8e8;
}
.link.primary-inline:hover:not(:disabled) {
  background: #ddd;
}
.link.primary-inline:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.loading-section {
  padding: 2rem;
  text-align: center;
  color: #666;
}
.result-section {
  margin-top: 2rem;
  padding-top: 1.5rem;
  border-top: 1px solid #eee;
}
.step-nav {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-bottom: 1rem;
}
.step-link {
  padding: 0.25rem 0.5rem;
  background: #f0f0f0;
  border-radius: 4px;
  color: #333;
  text-decoration: none;
  font-size: 0.9rem;
}
.step-link:hover {
  background: #e0e0e0;
}
.step-block {
  margin-bottom: 1.5rem;
  scroll-margin-top: 1rem;
}
.step-block h3 {
  margin: 0 0 0.5rem;
  font-size: 1.1rem;
}
.step-content {
  margin-bottom: 0.75rem;
  line-height: 1.6;
}
.step-image img {
  max-width: 100%;
  border-radius: 4px;
  cursor: pointer;
  border: 1px solid #eee;
}
.no-image {
  color: #999;
  font-size: 0.9rem;
  margin: 0;
}
.lightbox {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.85);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}
.lightbox img {
  max-width: 95%;
  max-height: 95%;
  object-fit: contain;
}
.lightbox .close {
  position: absolute;
  top: 1rem;
  right: 1rem;
  background: #fff;
  border: none;
  width: 2rem;
  height: 2rem;
  font-size: 1.5rem;
  line-height: 1;
  cursor: pointer;
  border-radius: 4px;
}
</style>
