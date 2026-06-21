import { ref } from 'vue'

const SUPPORTED_LOCALES = ['es-AR', 'en-US', 'es-ES', 'es-MX', 'fr-FR', 'pt-BR', 'ru-RU', 'de-DE'] as const
export type Locale = typeof SUPPORTED_LOCALES[number]
export { SUPPORTED_LOCALES }

export interface LangMeta {
  official: boolean
  level: number
}

export const LANG_META: Record<string, LangMeta> = {
  'es-AR': { official: true, level: 75 },
  'en-US': { official: true, level: 100 },
  'es-ES': { official: true, level: 75 },
  'es-MX': { official: true, level: 75 },
  'fr-FR': { official: true, level: 77 },
  'pt-BR': { official: true, level: 75 },
  'ru-RU': { official: true, level: 83 },
  'de-DE': { official: true, level: 74 },
}

export function getLangMeta(loc: string): LangMeta {
  const cached = messageCache[loc]?._metadata
  if (cached) {
    return {
      official: cached.es_oficial ?? LANG_META[loc]?.official ?? false,
      level: cached.nivel_de_traduccion ?? LANG_META[loc]?.level ?? 0,
    }
  }
  return LANG_META[loc] ?? { official: false, level: 0 }
}

export const locale = ref<string>('es-AR')

const messageCache: Record<string, Record<string, any>> = {}

async function loadFromElectron(loc: Locale): Promise<Record<string, any> | null> {
  if (typeof window !== 'undefined' && (window as any).ElectronAPI?.ReadLocaleFile) {
    return (window as any).ElectronAPI.ReadLocaleFile(loc)
  }
  console.warn('[LangManager] ElectronAPI.ReadLocaleFile not available — run inside Electron')
  return null
}

export async function loadLocale(loc: Locale): Promise<boolean> {
  if (messageCache[loc]) return true
  const msgs = await loadFromElectron(loc)
  if (msgs) {
    messageCache[loc] = msgs
    return true
  }
  console.warn(`[LangManager] No se pudo cargar el locale: ${loc}`)
  return false
}

export async function setLocale(loc: Locale) {
  await loadLocale(loc)
  locale.value = loc
  document.documentElement.lang = loc
  try {
    await (window as any).ConfigManager.UpdateLauncher({ locale: loc })
  } catch {}
}

export async function getInitialLocale(): Promise<Locale> {
  try {
    const config = await (window as any).ConfigManager.Get()
    const stored = config?.launcher?.locale as Locale | undefined
    if (stored && SUPPORTED_LOCALES.includes(stored)) return stored
  } catch {}

  const navLang = (navigator.language || '').toLowerCase()
  const exact = SUPPORTED_LOCALES.find(l => l.toLowerCase() === navLang)
  if (exact) return exact

  const langPrefix = navLang.split('-')[0]
  const prefixMatch = SUPPORTED_LOCALES.find(l => l.toLowerCase().startsWith(langPrefix + '-') || l.toLowerCase() === langPrefix)
  if (prefixMatch) return prefixMatch

  return 'es-AR'
}

export async function initLocale() {
  const initial = await getInitialLocale()
  await loadLocale(initial)
  locale.value = initial
  document.documentElement.lang = initial
}

export function getMessageCache(): Record<string, Record<string, any>> {
  return messageCache
}

export function t(key: string, params?: Record<string, any>): string {
  const loc = locale.value
  const msgs = messageCache[loc]
  if (!msgs) return key

  const parts = key.split('.')
  let val: any = msgs
  for (const part of parts) {
    if (val && typeof val === 'object') val = val[part]
    else return key
  }

  if (typeof val !== 'string') return key
  if (!params) return val

  return val.replace(/\{(\w+)\}/g, (_, k: string) =>
    params[k] !== undefined ? String(params[k]) : `{${k}}`
  )
}


