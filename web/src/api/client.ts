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

export async function startExplain(problemText: string): Promise<ExplainResponse> {
  const r = await fetch(`${BASE}/explain`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ problem_text: problemText }),
  })
  if (!r.ok) throw new Error(await r.text() || '解析失败')
  return r.json()
}

/** 直接根据已上传的题目图片让模型解析（不经过 OCR 识图） */
export async function startExplainFromImage(imagePath: string): Promise<ExplainResponse> {
  const r = await fetch(`${BASE}/explain`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ image_path: imagePath }),
  })
  if (!r.ok) throw new Error(await r.text() || '解析失败')
  return r.json()
}

export async function getResult(taskId: string): Promise<ResultResponse> {
  const r = await fetch(`${BASE}/result/${taskId}`)
  if (!r.ok) throw new Error(await r.text() || '获取结果失败')
  return r.json()
}
