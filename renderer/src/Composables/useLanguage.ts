import { computed } from 'vue'
import { locale, t, setLocale, getInitialLocale, SUPPORTED_LOCALES, loadLocale, type Locale } from '../i18n'

export function useLanguage() {
  const locales = SUPPORTED_LOCALES

  async function changeLanguage(loc: Locale) {
    await loadLocale(loc)
    await setLocale(loc)
  }

  function getFlagIcon(loc: Locale): string {
    return `./assets/locales/icons/${loc}.svg`
  }

  async function initLanguage() {
    const initial = await getInitialLocale()
    await loadLocale(initial)
    await setLocale(initial)
  }

  return {
    locale,
    locales,
    t,
    changeLanguage,
    getFlagIcon,
    initLanguage,
  }
}
