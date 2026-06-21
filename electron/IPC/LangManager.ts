import { ipcMain } from 'electron'
import { readLocaleFile } from '../LangManager.js'

export function RegisterLangManagerIpc() {
  ipcMain.handle('app:ReadLocaleFile', async (_event, locale: string) => {
    return readLocaleFile(locale)
  })
}
