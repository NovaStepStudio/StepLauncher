import { createApp } from 'vue'
import App from './App.vue'
import Router from './Router'
import { useAuth } from './Auth/Composables/useAuth'

const app = createApp(App)

app.use(Router)

// Directiva v-reveal: reveal suave al entrar en viewport.
// La clase .reveal se añade por JS (sin JS el contenido sigue visible)
// y .is-visible dispara la transición CSS. Liviano: un solo observer
// que se desconecta tras revelar.
app.directive('reveal', {
    mounted(el: HTMLElement) {
        el.classList.add('reveal');
        if (typeof IntersectionObserver === 'undefined') {
            el.classList.add('is-visible');
            return;
        }
        const io = new IntersectionObserver(
            (entries) => {
                for (const entry of entries) {
                    if (entry.isIntersecting) {
                        entry.target.classList.add('is-visible');
                        io.disconnect();
                    }
                }
            },
            { threshold: 0.12, rootMargin: '0px 0px -8% 0px' },
        );
        io.observe(el);
    },
})

// La sesión guardada se restaura antes de montar para no parpadear
// el header (entrar vs. panel) en la primera pintura.
useAuth().init().finally(() => {
    app.mount('#app');
});
