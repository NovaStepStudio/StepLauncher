import fs from 'fs'
import { fileURLToPath } from 'url'
import { dirname, join } from 'path'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

export const SUPPORTED_LOCALES = ['es-AR', 'en-US', 'es-ES', 'es-MX', 'fr-FR', 'pt-BR', 'ru-RU', 'de-DE'] as const

export function readLocaleFile(locale: string): Record<string, any> | null {
  if (!SUPPORTED_LOCALES.includes(locale as any)) {
    return null
  }

  const paths = [
    join(__dirname, '..', 'Renderer', 'assets', 'locales', `${locale}.json`),
    join(__dirname, '..', 'renderer', 'assets', 'locales', `${locale}.json`),
    join(__dirname, 'Renderer', 'assets', 'locales', `${locale}.json`),
  ]

  for (const filePath of paths) {
    try {
      return JSON.parse(fs.readFileSync(filePath, 'utf-8'))
    } catch {}
  }

  return null
}
