<script setup lang="ts">
import { IconShieldCheck } from '@tabler/icons-vue';
</script>

<template>
    <div class="Privacy">
        <div class="Hero sl-enter">
            <div class="Tag">
                <IconShieldCheck stroke="2" />
                <span>Legal</span>
            </div>
            <h1>Política de privacidad</h1>
            <p>Qué datos guarda StepLauncher, para qué los usa y cómo pedir que se borren. Sin letra chica.</p>
            <small>Última actualización: septiembre de 2026.</small>
        </div>
        <div class="Body" v-reveal>
            <section>
                <h2>1. Qué datos se guardan</h2>
                <p>Solo se guarda lo necesario para que tu cuenta funcione:</p>
                <ul>
                    <li><b>Cuenta:</b> correo electrónico, nombre de usuario y contraseña (nunca se guarda en texto plano).</li>
                    <li><b>Perfil:</b> nombre visible, biografía, avatar, banner, UUID de Minecraft y última versión jugada, si los configurás.</li>
                    <li><b>Archivos:</b> las skins y capas que subas, que son privadas y solo visibles para vos salvo que las uses en tu perfil público.</li>
                    <li><b>Comunidad:</b> solicitudes y lista de amigos, bloqueos y notificaciones de tu bandeja.</li>
                    <li><b>Presencia:</b> si estás en línea o no, para mostrarlo a tus amigos.</li>
                </ul>
            </section>
            <section>
                <h2>2. Qué datos NO se guardan</h2>
                <ul>
                    <li>Tus credenciales de Minecraft o de Microsoft: el launcher nunca las pide ni las almacena.</li>
                    <li>Contenido de tus mundos o partidas: quedan en tu computadora.</li>
                    <li>Datos de pago: no hay compras ni suscripciones.</li>
                </ul>
            </section>
            <section>
                <h2>3. Dónde viven tus datos</h2>
                <p>La autenticación y la base de datos están en Supabase, con políticas de acceso que hacen que cada usuario solo pueda leer y modificar su propia información. La sesión queda guardada únicamente en tu navegador o dispositivo para mantenerte conectado; nunca viaja en URLs ni se muestra en la interfaz.</p>
            </section>
            <section>
                <h2>4. Para qué se usan</h2>
                <ul>
                    <li>Iniciar sesión y mantener tu sesión activa.</li>
                    <li>Mostrar tu perfil a la comunidad según tu configuración de privacidad.</li>
                    <li>Enviar notificaciones dentro del panel (nunca por correo salvo avisos de seguridad de tu cuenta).</li>
                    <li>Prevenir abusos, como el spam de solicitudes o la subida masiva de archivos.</li>
                </ul>
            </section>
            <section>
                <h2>5. Tus derechos</h2>
                <p>Podés ver y editar tu perfil, tu privacidad, tu correo y tu contraseña desde el panel en cualquier momento. Si querés borrar tu cuenta y todos sus datos asociados, escribinos desde GitHub y lo procesamos. Al borrar la cuenta se eliminan perfil, archivos, amistades y notificaciones.</p>
            </section>
            <section>
                <h2>6. Menores de edad</h2>
                <p>StepLauncher no está dirigido a menores de 13 años. Si sos menor, usalo con supervisión de un adulto responsable.</p>
            </section>
            <section>
                <h2>7. Cambios a esta política</h2>
                <p>Si esta política cambia de forma importante, la fecha de arriba se actualiza y se avisa en el historial de cambios. El uso continuado después de un cambio implica aceptación.</p>
            </section>
            <section>
                <h2>8. Contacto</h2>
                <p>Para dudas o pedidos sobre tus datos: <a href="https://github.com/NovaStepStudio/StepLauncher" target="_blank" rel="noopener">NovaStepStudio en GitHub</a>.</p>
            </section>
        </div>
    </div>
</template>

<style scoped lang="scss">
.Privacy{
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
                    b{
                        opacity: 1;
                    }
                }
            }
            a{
                color: #fff;
            }
        }
    }
}
</style>
