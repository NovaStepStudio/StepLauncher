<script setup lang="ts">
import { IconScale } from '@tabler/icons-vue';
</script>

<template>
    <div class="Terms">
        <div class="Hero sl-enter">
            <div class="Tag">
                <IconScale stroke="2" />
                <span>Legal</span>
            </div>
            <h1>Términos y condiciones</h1>
            <p>Las reglas para usar StepLauncher y tu cuenta. Cortas y en español.</p>
            <small>Última actualización: septiembre de 2026.</small>
        </div>
        <div class="Body" v-reveal>
            <section>
                <h2>1. Qué es StepLauncher</h2>
                <p>StepLauncher es un launcher gratuito y de código abierto para Minecraft: Java Edition, hecho por NovaStepStudio. Incluye una app de escritorio, esta web y una API de cuentas y comunidad. Se entrega "tal cual", en versión beta y en desarrollo activo.</p>
            </section>
            <section>
                <h2>2. Licencia del código</h2>
                <p>El código es open source bajo licencia GPL-3.0: podés usarlo, estudiarlo y mejorarlo respetando esa licencia. Las marcas, logos y diseños del proyecto no están incluidos en esa libertad: no los uses para hacer pasar otro proyecto por StepLauncher.</p>
            </section>
            <section>
                <h2>3. Minecraft no incluido</h2>
                <p>StepLauncher no incluye ni vende Minecraft: necesitás tu propia copia del juego. No está afiliado a Mojang Studios ni a Microsoft. Tampoco es una cuenta premium ni genera cuentas del juego.</p>
            </section>
            <section>
                <h2>4. Tu cuenta</h2>
                <ul>
                    <li>Sos responsable de mantener tu contraseña en secreto y de lo que se haga con tu cuenta.</li>
                    <li>No compartas tu sesión ni intentes entrar a cuentas ajenas.</li>
                    <li>Podemos suspender cuentas que hagan spam, suban contenido ilegal, acosen a otros jugadores o intenten romper la seguridad del servicio.</li>
                </ul>
            </section>
            <section>
                <h2>5. Uso aceptable</h2>
                <ul>
                    <li>No subir contenido ilegal, ofensivo o que viole derechos de terceros como avatar, banner, skin o biografía.</li>
                    <li>No abusar de los límites (subidas, solicitudes, intentos de acceso) ni automatizar el servicio para degradarlo.</li>
                    <li>No hacer ingeniería inversa del backend con fines maliciosos ni intentar acceder a datos de otros usuarios.</li>
                </ul>
            </section>
            <section>
                <h2>6. Disponibilidad y cambios</h2>
                <p>Al estar en beta, puede haber errores, caídas o cambios sin aviso. Podemos modificar o descontinuar funciones, manteniendo el historial de cambios a la vista. Los errores visibles de la app y la web son mensajes genéricos: el detalle técnico nunca incluye tus datos.</p>
            </section>
            <section>
                <h2>7. Responsabilidad</h2>
                <p>En la medida máxima permitida por la ley, NovaStepStudio no responde por daños derivados del uso del servicio, pérdida de datos locales de tu computadora o interrupciones. Tus mundos y configuraciones locales son tu responsabilidad: hacé copias de seguridad.</p>
            </section>
            <section>
                <h2>8. Privacidad</h2>
                <p>El tratamiento de tus datos se explica en la <RouterLink to="/privacy">Política de privacidad</RouterLink>, que forma parte de estos términos.</p>
            </section>
            <section>
                <h2>9. Contacto</h2>
                <p>Para consultas sobre estos términos: <a href="https://github.com/NovaStepStudio/StepLauncher" target="_blank" rel="noopener">NovaStepStudio en GitHub</a>.</p>
            </section>
        </div>
    </div>
</template>

<style scoped lang="scss">
.Terms{
    display:flex;
    flex-direction:column;
    align-items:center;
    background:#000;
    padding: 7rem 1.5rem 4rem 1.5rem;
    .Hero{
        display:flex;
        flex-direction:column;
        align-items:center;
        text-align:center;
        gap: .8rem;
        max-width: 42rem;
        margin-bottom: 2rem;
        .Tag{
            display:flex;
            align-items:center;
            gap: .4rem;
            font-size: .7rem;
            text-transform: uppercase;
            letter-spacing: .12em;
            padding: .25rem .65rem;
            border-radius: 99rem;
            border: 1px solid #ffffff25;
            background: #ffffff0d;
            opacity: .85;
            svg{
                width: .95rem;
                height: .95rem;
            }
        }
        h1{
            margin: 0;
            font-family: 'Lexend';
            font-size: 2.2rem;
            font-weight: 600;
        }
        p{
            margin: 0;
            font-size: .95rem;
            line-height: 1.6;
            opacity: .65;
        }
        small{
            font-size: .72rem;
            opacity: .4;
        }
    }
    .Body{
        width: 100%;
        max-width: 42rem;
        display:flex;
        flex-direction:column;
        gap: 1.8rem;
        section{
            h2{
                margin: 0 0 .5rem 0;
                font-family: 'Lexend';
                font-size: 1.05rem;
                font-weight: 600;
            }
            p{
                margin: 0;
                font-size: .85rem;
                line-height: 1.7;
                opacity: .7;
            }
            ul{
                margin: .5rem 0 0 0;
                padding-left: 1.2rem;
                display:flex;
                flex-direction:column;
                gap: .4rem;
                li{
                    font-size: .85rem;
                    line-height: 1.7;
                    opacity: .7;
                }
            }
            a{
                color: #fff;
            }
        }
    }
}
</style>
