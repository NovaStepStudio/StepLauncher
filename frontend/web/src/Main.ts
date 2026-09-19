import { createApp } from 'vue'
import App from './App.vue'
import BackgroundMusicWidget from '@/Common/Widgets/BackgroundMusic.vue'
import '@/Common/Styles/base/_index.scss'
import { runBootstrap } from '@/Common/Bootstrap'
import { bootstrapState } from '@/Common/Bootstrap/state'

// Asegura que el splash sea visible desde el primer fotograma, incluso antes
// de que Vue termine de montar. El estado ya es idle visible, pero lo
// marcamos como running de inmediato para que el texto no sea estatico.
bootstrapState.value.status = 'running'
bootstrapState.value.startedAt = Date.now()
bootstrapState.value.currentLabel = 'Iniciando StepLauncher'
bootstrapState.value.progress = 2

const app = createApp(App)
app.mount('#app')

const bgmHost = document.createElement('div')
bgmHost.id = 'bgm-host'
document.body.appendChild(bgmHost)
createApp(BackgroundMusicWidget).mount(bgmHost)

function hideStaticSplash() {
    const el = document.getElementById('static-splash')
    if (el) {
        el.style.transition = 'opacity 420ms ease, filter 420ms ease'
        el.style.opacity = '0'
        el.style.filter = 'blur(6px)'
        setTimeout(() => el.remove(), 520)
    }
}

// Espera a que el runtime de Wails este listo antes de tocar bindings.
// En dev el primer paint puede ocurrir antes de que window._wails exista.
function waitForRuntime(): Promise<void> {
    return new Promise((resolve) => {
        let tries = 0
        const check = () => {
            tries++
            // Si ya existe el bridge de Wails o pasaron 2s, se avanza igual
            const wailsReady = (window as any)._wails || (window as any).wails
            if (wailsReady || tries > 20) resolve()
            else setTimeout(check, 100)
        }
        if (document.readyState === 'complete') check()
        else window.addEventListener('load', check, { once: true })
        // Fallback por si load no dispara en Wails
        setTimeout(() => resolve(), 1200)
    })
}

waitForRuntime().then(() => {
    // Oculta el splash estatico en cuanto Vue ya tiene su propio SplashScreen
    // pero lo deja 80ms para evitar flash blanco entre ambos.
    setTimeout(hideStaticSplash, 80)
    return runBootstrap({ minSplashMs: 950, maxSplashMs: 16500 })
}).catch((e) => {
    console.error('[Main] bootstrap fallo', e)
    hideStaticSplash()
})

// Por si algo bloquea waitForRuntime, asegura que el estatico no quede colgado
setTimeout(hideStaticSplash, 7000)
