const BASE = '/api'

export type UploadResponse = { path: string }
export type SubmitResponse = { problem_text: string }
export type ExplainResponse = { task_id: string }
export type StepResponse = { title: string; content: string; image_url?: string }
export type ResultResponse = { steps: StepResponse[] }

export async function uploadImage(file: File): Promise<UploadResponse> {
  const form = new FormData()
  form.append('file', file)
  const r = await fetch(`${BASE}/upload`, { method: 'POST', body: form })
  if (!r.ok) throw new Error(await r.text() || '上传失败')
  return r.json()
}

export async function submitProblem(problemText: string): Promise<SubmitResponse> {
  const r = await fetch(`${BASE}/submit`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ text: problemText }),
  })
  if (!r.ok) throw new Error(await r.text() || '提交失败')
  return r.json()
}

export async function submitImage(imagePath: string): Promise<SubmitResponse> {
  const r = await fetch(`${BASE}/submit`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ image_path: imagePath }),
  })
  if (!r.ok) throw new Error(await r.text() || '识图失败')
  return r.json()
}

/** 解析接口可能较慢（多模态/长文本），给足时间避免前端先超时 */
const EXPLAIN_TIMEOUT_MS = 4 * 60 * 1000

async function explainFetch(body: object): Promise<Response> {
  const ac = new AbortController()
  const t = setTimeout(() => ac.abort(), EXPLAIN_TIMEOUT_MS)
  let r: Response
  try {
    r = await fetch(`${BASE}/explain`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
      signal: ac.signal,
    })
  } catch (e) {
    clearTimeout(t)
    if (e instanceof Error && e.name === 'AbortError') {
      throw new Error('请求超时，请稍后重试或缩短题目内容')
    }
    throw e
  }
  clearTimeout(t)
  if (!r.ok) {
    const text = await r.text()
    const msg = text?.trim() || '解析失败'
    if (r.status === 504) throw new Error(`请求超时（504），请稍后重试。${msg ? ` ${msg}` : ''}`)
    if (r.status === 500) throw new Error(msg)
    throw new Error(`${r.status}: ${msg}`)
  }
  return r
}

export async function startExplain(problemText: string): Promise<ExplainResponse> {
  const r = await explainFetch({ problem_text: problemText })
  return r.json()
}

/** 直接根据已上传的题目图片让模型解析（不经过 OCR 识图） */
export async function startExplainFromImage(imagePath: string): Promise<ExplainResponse> {
  const r = await explainFetch({ image_path: imagePath })
  return r.json()
}

export async function getResult(taskId: string): Promise<ResultResponse> {
  const r = await fetch(`${BASE}/result/${taskId}`)
  if (!r.ok) throw new Error(await r.text() || '获取结果失败')
  return r.json()
}

// 解析历史（存后端）
export type HistoryStep = { title: string; content: string; image_url?: string }
export type HistoryResult = { steps: HistoryStep[] }
export type HistoryItem = {
  id: string
  type: 'upload' | 'text'
  path?: string
  text?: string
  at: number
  result?: HistoryResult | null
  task_id?: string
}

export async function listHistory(): Promise<HistoryItem[]> {
  const r = await fetch(`${BASE}/history`, { cache: 'no-store' })
  if (!r.ok) throw new Error(await r.text() || '获取历史失败')
  const data = await r.json()
  const list = data.items ?? data.Items ?? []
  return Array.isArray(list) ? list : []
}

export async function createHistoryItem(item: {
  type: 'upload' | 'text'
  path?: string
  text?: string
  at: number
}): Promise<{ id: string }> {
  const r = await fetch(`${BASE}/history`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(item),
  })
  if (!r.ok) throw new Error(await r.text() || '添加历史失败')
  return r.json()
}

export async function updateHistoryResult(
  id: string,
  result: ResultResponse,
  taskId: string
): Promise<void> {
  const r = await fetch(`${BASE}/history/${id}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ result, task_id: taskId }),
  })
  if (!r.ok) throw new Error(await r.text() || '更新历史失败')
}

export async function deleteHistoryItem(id: string): Promise<void> {
  const r = await fetch(`${BASE}/history/${id}`, { method: 'DELETE' })
  if (!r.ok) throw new Error(await r.text() || '删除失败')
}

export async function findLatestUploadHistoryId(path: string): Promise<string | null> {
  const r = await fetch(`${BASE}/history/find-upload?path=${encodeURIComponent(path)}`)
  if (!r.ok) return null
  const data = await r.json()
  return data.id ?? null
}
