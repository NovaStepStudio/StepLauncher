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
                <p>StepLauncher es un launcher gratuito y de código abierto para Minecraft: Java Edition, hecho por NovaStepStudio. Incluye una app de escritorio, esta web, una API de cuentas y comunidad, y un servicio de autenticación compatible con el protocolo Yggdrasil: tu cuenta StepLauncher puede usarse como identidad dentro del juego en servidores que lo acepten, y el launcher además permite agregar servidores Yggdrasil externos de terceros. Se entrega "tal cual", en versión beta y en desarrollo activo.</p>
            </section>
            <section>
                <h2>2. Licencia del código</h2>
                <p>El código es open source bajo licencia GPL-3.0: podés usarlo, estudiarlo y mejorarlo respetando esa licencia. Las marcas, logos y diseños del proyecto no están incluidos en esa libertad: no los uses para hacer pasar otro proyecto por StepLauncher.</p>
            </section>
            <section>
                <h2>3. Minecraft, modloaders y Mojang</h2>
                <p>StepLauncher <b>descarga y ejecuta</b> Minecraft: Java Edition en tu computadora desde fuentes oficiales, siguiendo las políticas de Mojang Studios y Microsoft (EULA y términos de uso del juego). También puede descargar modloaders como Forge, Fabric, Quilt o NeoForge desde sus fuentes oficiales para que los uses en tus instancias.</p>
                <ul>
                    <li><b>No vende el juego ni otorga licencias:</b> necesitás tener derecho a jugar Minecraft según las políticas de Mojang y Microsoft.</li>
                    <li><b>Proyecto independiente:</b> no está afiliado a Mojang Studios ni a Microsoft. El servicio Yggdrasil de StepLauncher es una identidad alternativa e independiente: no desbloquea ni reemplaza la autenticación oficial. Los servidores que exigen cuenta oficial van a seguir exigiéndola.</li>
                    <li><b>Modo offline y online:</b> en modo offline jugás sin verificar tu identidad con nadie. En modo online iniciás sesión contra el servicio Yggdrasil que vos elijas (StepLauncher u otro externo).</li>
                    <li><b>Cómo viaja tu identidad en modo online:</b> al iniciar sesión, StepLauncher recupera los datos de tu cuenta de ese servicio Yggdrasil (nombre, UUID y texturas) y, al lanzar el juego con authlib-injector, es el propio Minecraft el que comparte esa identidad con el servidor asociado para verificarte. Tus datos de cuenta solo llegan al servicio Yggdrasil que elegiste y al servidor donde jugás: StepLauncher no envía tu cuenta ni tus datos personales a Mojang o Microsoft.</li>
                </ul>
            </section>
            <section>
                <h2>4. Tu cuenta</h2>
                <ul>
                    <li>Sos responsable de mantener tu contraseña en secreto y de lo que se haga con tu cuenta, incluyendo las sesiones de juego Yggdrasil que se generen al entrar a un servidor.</li>
                    <li>No compartas tu sesión ni tus tokens de acceso, ni intentes entrar a cuentas ajenas o hacerte pasar por otro jugador.</li>
                    <li>Si usás un dispositivo compartido, cerrá sesión al terminar para no dejar sesiones Yggdrasil activas.</li>
                    <li>Podemos suspender cuentas que hagan spam, suban contenido ilegal, acosen a otros jugadores o intenten romper la seguridad del servicio.</li>
                </ul>
            </section>
            <section>
                <h2>5. Servicio Yggdrasil público</h2>
                <p>Yggdrasil es el protocolo abierto que usan launchers y servidores para verificar tu identidad en el juego (nombre, UUID y texturas como skin y capa). El de StepLauncher es un servicio público: cualquier servidor puede aceptarlo y cualquier jugador con cuenta puede usarlo.</p>
                <ul>
                    <li><b>Cuenta propia:</b> si un servidor de Minecraft acepta StepLauncher, podés entrar con tu cuenta en modo online. Al lanzar el juego con authlib-injector, tu nombre, tu UUID y tus texturas se comparten con ese servidor para verificarte y dejarte jugar.</li>
                    <li><b>Servidores externos:</b> podés agregar en el launcher direcciones Yggdrasil de terceros. Esos servicios son ajenos: no los operamos, no los controlamos ni garantizamos su seguridad o disponibilidad. Usalos bajo tu responsabilidad y revisá sus propias reglas.</li>
                    <li><b>Nunca ingreses tu contraseña de StepLauncher en servicios externos.</b> El acceso con externos se configura con sus propios datos y queda guardado solo en tu dispositivo.</li>
                    <li>No uses Yggdrasil para evadir los controles de un servidor que exige autenticación oficial, ni para suplantar identidades o texturas de otros jugadores.</li>
                </ul>
            </section>
            <section>
                <h2>6. Uso aceptable</h2>
                <ul>
                    <li>No subir contenido ilegal, ofensivo o que viole derechos de terceros como avatar, banner, skin o biografía.</li>
                    <li>No abusar de los límites (subidas, solicitudes, intentos de acceso e inicios de sesión Yggdrasil) ni automatizar el servicio para degradarlo.</li>
                    <li>No hacer ingeniería inversa del backend con fines maliciosos ni intentar acceder a datos de otros usuarios.</li>
                    <li>Respetar las reglas de cada servidor de Minecraft al que entres: StepLauncher no responde por sanciones que te aplique un servidor por incumplirlas.</li>
                </ul>
            </section>
            <section>
                <h2>7. Disponibilidad y cambios</h2>
                <p>Al estar en beta, puede haber errores, caídas o cambios sin aviso. Podemos modificar o descontinuar funciones, incluyendo el servicio Yggdrasil, manteniendo el historial de cambios a la vista. Los errores visibles de la app y la web son mensajes genéricos: el detalle técnico nunca incluye tus datos.</p>
            </section>
            <section>
                <h2>8. Responsabilidad</h2>
                <p>En la medida máxima permitida por la ley, NovaStepStudio no responde por daños derivados del uso del servicio, pérdida de datos locales de tu computadora o interrupciones. Tus mundos y configuraciones locales son tu responsabilidad: hacé copias de seguridad. Tampoco responde por lo que hagan los servidores Yggdrasil externos o los servidores de juego de terceros a los que te conectes.</p>
            </section>
            <section>
                <h2>9. Privacidad</h2>
                <p>El tratamiento de tus datos, incluyendo tu identidad Yggdrasil y las sesiones de juego, se explica en la <RouterLink to="/privacy">Política de privacidad</RouterLink>, que forma parte de estos términos.</p>
            </section>
            <section>
                <h2>10. Contacto</h2>
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
