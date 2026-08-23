export type Severity = 'critical' | 'high' | 'medium' | 'low' | 'info'

export interface Task {
  id: string
  base_url: string
  status: string
  concurrency: number
  timeout_ms: number
  authorized: boolean
  swagger_name: string
  total: number
  sent: number
  hits: number
  error: string
  created_at: string
  updated_at: string
  critical: number
  high: number
  medium: number
  low: number
  info: number
}

export interface Finding {
  id: string
  task_id: string
  endpoint: string
  method: string
  class: string
  severity: Severity
  title: string
  evidence: string
  payload: string
  param_name: string
  status_code: number
  latency_ms: number
  advice: string
  created_at: string
}

export interface DefectNode {
  key: string
  label: string
  method?: string
  path?: string
  severity?: Severity
  count: number
  finding?: Finding
  children: DefectNode[]
}

export interface Envelope<T> {
  ok: boolean
  error?: string
  data: T
}

export interface Meta {
  scan_mode: string
  lab_public_url: string
  default_base_url: string
  timezone: string
}

export interface ReportPreview {
  task: Task
  findings: Finding[]
  tree: DefectNode[]
  stats: { critical: number; high: number; medium: number; low: number; info: number }
  advice: string[]
  generated_at: string
}

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, init)
  const body = await res.json() as Envelope<T>
  if (!res.ok || !body.ok) {
    throw new Error(body.error || `HTTP ${res.status}`)
  }
  return body.data
}

export const api = {
  meta: () => req<Meta>('/api/v1/meta'),
  health: () => req<{ status: string }>('/api/v1/health'),
  list: () => req<Task[]>('/api/v1/scans'),
  get: (id: string) => req<Task>(`/api/v1/scans/${id}`),
  findings: (id: string) => req<{ findings: Finding[]; tree: DefectNode[]; stats: ReportPreview['stats'] }>(`/api/v1/scans/${id}/findings`),
  report: (id: string) => req<ReportPreview>(`/api/v1/scans/${id}/report`),
  cancel: (id: string) => req<{ status: string }>(`/api/v1/scans/${id}/cancel`, { method: 'POST' }),
  createJSON: (payload: { base_url: string; concurrency: number; timeout_ms: number; authorized: boolean }) =>
    req<Task>('/api/v1/scans', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    }),
  createForm: (fd: FormData) =>
    req<Task>('/api/v1/scans', { method: 'POST', body: fd }),
}

export function pdfURL(id: string) {
  return `/api/v1/scans/${id}/report.pdf`
}

export function wsURL(id: string) {
  const proto = location.protocol === 'https:' ? 'wss' : 'ws'
  return `${proto}://${location.host}/api/v1/ws?task_id=${encodeURIComponent(id)}`
}
