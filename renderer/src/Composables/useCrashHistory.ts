import { ref } from 'vue'

export interface CrashRecord {
  id: number
  reason: string
  context: string[]
  exitCode: number
  timestamp: number
  _expanded?: boolean
}

const crashHistory = ref<CrashRecord[]>([])
let nextCrashId = 0

export function useCrashHistory() {
  function addCrash(reason: string, context: string[], exitCode: number): CrashRecord {
    const record: CrashRecord = {
      id: nextCrashId++,
      reason,
      context,
      exitCode,
      timestamp: Date.now(),
    }
    crashHistory.value.unshift(record)
    if (crashHistory.value.length > 50) {
      crashHistory.value = crashHistory.value.slice(0, 50)
    }
    return record
  }

  function clearHistory() {
    crashHistory.value = []
  }

  function removeCrash(id: number) {
    const idx = crashHistory.value.findIndex(c => c.id === id)
    if (idx >= 0) crashHistory.value.splice(idx, 1)
  }

  return { crashHistory, addCrash, clearHistory, removeCrash }
}
