import { defineStore } from 'pinia'
import { api, type DefectNode, type Finding, type Task } from '../api'

export interface StreamLine {
  ts: string
  level: string
  message: string
  flash?: boolean
}

export const useScanStore = defineStore('scan', {
  state: () => ({
    task: null as Task | null,
    tree: [] as DefectNode[],
    findings: [] as Finding[],
    lines: [] as StreamLine[],
    toast: '' as string,
  }),
  actions: {
    setTask(t: Task) {
      this.task = t
    },
    pushLine(line: StreamLine) {
      this.lines.push({ ...line, flash: true })
      if (this.lines.length > 400) this.lines.splice(0, this.lines.length - 400)
    },
    pushFinding(f: Finding) {
      if (this.findings.some((x) => x.id === f.id)) return
      this.findings.push(f)
    },
    async hydrate(id: string) {
      this.task = await api.get(id)
      const pack = await api.findings(id)
      this.findings = pack.findings
      this.tree = pack.tree
    },
    showToast(msg: string) {
      this.toast = msg
      window.setTimeout(() => {
        if (this.toast === msg) this.toast = ''
      }, 5000)
    },
  },
})
