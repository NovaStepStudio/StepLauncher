import './Styles/Global.scss'
import { createApp } from 'vue'
import App from './App.vue'
import i18n, { initI18n, t } from './i18n'

document.addEventListener('DOMContentLoaded', () => {
  document.getElementById('btn-minimize')?.addEventListener('click', () => window.ElectronAPI.Minimize())
  document.getElementById('btn-maximize')?.addEventListener('click', () => window.ElectronAPI.Maximize())
  document.getElementById('btn-close')?.addEventListener('click', () => window.ElectronAPI.Close())
})

const app = createApp(App)
app.use(i18n)

app.config.globalProperties.$t = t as any

initI18n().then(() => {
  app.mount('#app')
})
