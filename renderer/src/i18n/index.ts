import { createI18n } from 'vue-i18n'
import * as LangManager from './LangManager'

const i18n = createI18n({
  legacy: false,
  locale: 'es-AR',
  fallbackLocale: 'es-AR',
})

export async function initI18n() {
  await LangManager.initLocale()

  const loc = LangManager.locale.value
  ;(i18n.global.locale as any).value = loc
  if ((i18n.global.availableLocales as string[]).includes(loc)) return

  const msgs = LangManager.getMessageCache()?.[loc]
  if (msgs) {
    i18n.global.setLocaleMessage(loc, msgs as any)
  }
}

export { LangManager }
export * from './LangManager'
export default i18n
