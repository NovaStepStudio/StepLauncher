<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, provide, reactive, ref } from 'vue';
import { IconListDetails, IconSearch } from '@tabler/icons-vue';
import FaqGroup from '@/Faq/Components/FaqGroup.vue';
import FaqItem from '@/Faq/Components/FaqItem.vue';
import { FAQ_KEY, metaDelGrupo, visiblesDelGrupo, type FaqContexto } from '@/Faq/faq';

// Las preguntas están abajo en la plantilla: cada <FaqItem> se registra
// solo y entra al buscador, al índice y a los conteos sin más trámite.
const consulta = ref('');
const ctx: FaqContexto = {
    consulta,
    grupos: reactive(new Map<string, string>()),
    items: reactive(new Map<symbol, { grupo: string; texto: string }>()),
};
provide(FAQ_KEY, ctx);

const listaGrupos = computed(() => [...ctx.grupos.entries()]);
const buscando = computed(() => consulta.value.trim() !== '');
const totalVisibles = computed(() => {
    const q = consulta.value.trim().toLowerCase();
    if (q === '') return ctx.items.size;
    let n = 0;
    for (const item of ctx.items.values()) {
        if (item.texto.includes(q)) n++;
    }
    return n;
});

function conteo(id: string): number {
    return visiblesDelGrupo(ctx, id);
}

function limpiar() {
    consulta.value = '';
}

// Scrollspy: resalta en el índice el tema que el usuario está mirando.
const activo = ref('');
let observador: IntersectionObserver | null = null;

onMounted(() => {
    nextTick(() => {
        observador = new IntersectionObserver(
            (entradas) => {
                for (const e of entradas) {
                    if (e.isIntersecting) activo.value = e.target.id.replace(/^grupo-/, '');
                }
            },
            { rootMargin: '-20% 0px -70% 0px' },
        );
        for (const id of ctx.grupos.keys()) {
            const el = document.getElementById(`grupo-${id}`);
            if (el) observador.observe(el);
        }
    });
});

onUnmounted(() => observador?.disconnect());
</script>

<template>
    <div class="Faq" id="faq-top">
        <div class="Hero sl-enter">
            <div class="Tag">
                <IconListDetails stroke="2" />
                <span>Ayuda</span>
            </div>
            <h1>Preguntas frecuentes</h1>
            <p>Respuestas directas sobre el launcher, las descargas oficiales, los mods y las cuentas con Yggdrasil. Todo a la vista, en español.</p>
            <div class="Badges">
                <span class="BadgeMain">{{ ctx.items.size }} {{ ctx.items.size === 1 ? 'respuesta' : 'respuestas' }}</span>
                <span class="Badge">{{ ctx.grupos.size }} temas</span>
            </div>
        </div>
        <nav class="Temas sl-stagger" aria-label="Temas">
            <a v-for="[id, titulo] in listaGrupos" :key="id" class="TemaCard" :href="`#grupo-${id}`">
                <span class="CardIcon"><component :is="metaDelGrupo(id).icono" stroke="2" /></span>
                <span class="CardTxt">
                    <b>{{ titulo }}</b>
                    <small>{{ metaDelGrupo(id).descripcion }}</small>
                </span>
                <i>{{ conteo(id) }}</i>
            </a>
        </nav>
        <div class="Layout">
            <aside v-if="!buscando" class="Side" aria-label="Temas">
                <b>Temas</b>
                <a
                    v-for="[id, titulo] in listaGrupos"
                    :key="id"
                    :href="`#grupo-${id}`"
                    :class="{ active: activo === id }"
                >
                    {{ titulo }}
                    <i>{{ conteo(id) }}</i>
                </a>
            </aside>
            <div class="Main">
                <label class="Search">
                    <IconSearch stroke="2" />
                    <input v-model="consulta" type="search" placeholder="Buscar en el FAQ..." aria-label="Buscar en el FAQ">
                </label>
                <p v-if="buscando" class="ResultCount">
                    {{ totalVisibles }} {{ totalVisibles === 1 ? 'resultado' : 'resultados' }} para "{{ consulta }}"
                </p>
                <template v-if="!buscando || totalVisibles > 0">
                    <FaqGroup id="general" titulo="General">
                        <FaqItem pregunta="¿Qué es StepLauncher?">
                            <p>StepLauncher es un launcher gratuito y de código abierto para Minecraft: Java Edition, desarrollado por NovaStepStudio. Reúne en una sola app de escritorio todo lo que necesitás para jugar: descarga y ejecución del juego desde fuentes oficiales, instancias con su versión y su Java, modloaders (Forge, Fabric, Quilt y NeoForge), mods de Modrinth, cuentas con Yggdrasil público, música y personalización total.</p>
                            <p>Lo acompañan esta web (descargas, historial y tu panel de cuenta) y una API de cuentas y comunidad. Todo el código es abierto bajo licencia GPL-3.0 y el proyecto está en beta, en desarrollo activo. Es independiente: no está afiliado a Mojang Studios ni a Microsoft. Conocé más en <RouterLink to="/about">Acerca de</RouterLink>.</p>
                        </FaqItem>
                        <FaqItem pregunta="¿StepLauncher tiene algún costo económico?">
                            <p>No: es gratis para siempre, sin compras, suscripciones ni funciones bloqueadas. Lo único que necesitás por tu cuenta es el derecho a jugar Minecraft según las políticas de Mojang y Microsoft: StepLauncher no vende el juego ni otorga licencias.</p>
                        </FaqItem>
                        <FaqItem pregunta="¿StepLauncher es seguro? ¿Está libre de virus?">
                            <p>Sí. El código es abierto y auditable por cualquiera, y cada versión se publica en el <a href="https://github.com/NovaStepStudio/StepLauncher/releases" target="_blank" rel="noopener">GitHub oficial</a>. No incluye malware, mineros, adware ni programas extra: lo que descargás es solo el launcher.</p>
                            <p>Dos consejos para estar tranquilo: descargalo siempre desde esta web o desde el GitHub oficial, nunca desde enlaces de terceros; y recordá que está en beta, así que puede tener errores. Tus mundos y configuraciones viven en tu computadora: hacé copias de seguridad.</p>
                        </FaqItem>
                        <FaqItem pregunta="¿En qué sistemas funciona?">
                            <p>En Windows, Linux y macOS: la misma app, con descargas separadas en <RouterLink to="/download">Descarga</RouterLink>. Es una app de escritorio: no hay versión web jugable ni app de celular.</p>
                        </FaqItem>
                        <FaqItem pregunta="¿Por qué elegir StepLauncher sobre otros launchers?">
                            <ul>
                                <li><b>Liviano y rápido:</b> backend en Go con app de escritorio nativa, sin el peso de Electron.</li>
                                <li><b>Todo integrado:</b> versiones, modloaders, instancias, mods de Modrinth, música y personalización en una sola app.</li>
                                <li><b>Cuentas con Yggdrasil público:</b> tu cuenta es tu identidad en el juego y cualquier servidor puede aceptarla, además de poder sumar servidores externos.</li>
                                <li><b>En tu idioma:</b> interfaz y errores explicados en español, sin códigos crípticos.</li>
                                <li><b>Código abierto:</b> auditable por cualquiera bajo licencia GPL-3.0.</li>
                            </ul>
                        </FaqItem>
                        <FaqItem pregunta="¿Qué otras características hacen que StepLauncher sea único?">
                            <p>La combinación de todo en un solo lugar: reproductor de música con playlists sin salir del launcher, personalización total con vista previa en vivo, servicio propio de autenticación Yggdrasil abierto a cualquier servidor, panel web con skins y capas en 3D, amigos y notificaciones, historial de errores explicado en español, e instancias que se pueden clonar, verificar y respaldar.</p>
                        </FaqItem>
                    </FaqGroup>
                    <FaqGroup id="juego" titulo="Juego y descargas">
                        <FaqItem pregunta="¿Qué versiones de Minecraft están disponibles?">
                            <p>Las de Minecraft: Java Edition, organizadas por instancias: cada instancia guarda su versión, el Java que le corresponde y su configuración propia. Podés tener perfiles vanilla, snapshots y perfiles con mods sin que se pisen entre sí, clonarlos para probar combinaciones y verificar o respaldar sus archivos cuando quieras.</p>
                        </FaqItem>
                        <FaqItem pregunta="¿Desde dónde se descargan el juego y los modloaders?">
                            <p>Todo se descarga desde los proyectos oficiales y se ejecuta en tu computadora:</p>
                            <ul>
                                <li><b>El juego,</b> desde la infraestructura oficial de Mojang (<a href="https://www.minecraft.net/" target="_blank" rel="noopener">minecraft.net</a>).</li>
                                <li><b>Los modloaders,</b> cada uno desde su sitio oficial: Forge (<a href="https://files.minecraftforge.net/" target="_blank" rel="noopener">files.minecraftforge.net</a>), Fabric (<a href="https://fabricmc.net/" target="_blank" rel="noopener">fabricmc.net</a>), Quilt (<a href="https://quiltmc.org/" target="_blank" rel="noopener">quiltmc.org</a>) y NeoForge (<a href="https://neoforged.net/" target="_blank" rel="noopener">neoforged.net</a>).</li>
                                <li><b>Los mods, modpacks, shaders y paquetes de recursos,</b> desde <a href="https://modrinth.com/" target="_blank" rel="noopener">Modrinth</a>, integrado al launcher.</li>
                            </ul>
                            <p>StepLauncher respeta las políticas de Mojang y Microsoft (EULA y términos de uso): no vende el juego ni otorga licencias.</p>
                        </FaqItem>
                        <FaqItem pregunta="¿Qué Java necesito?">
                            <p>Depende de la versión: cada versión de Minecraft pide un Java específico, y cada instancia guarda el suyo con su configuración. Elegí el Java que corresponda a la versión que vas a jugar y queda configurado por instancia, sin afectar a las demás.</p>
                        </FaqItem>
                        <FaqItem pregunta="¿Es una cuenta premium? ¿Necesito tener el juego comprado?">
                            <p>No: la cuenta de StepLauncher no es premium ni genera cuentas del juego. Su Yggdrasil es una identidad alternativa e independiente que no desbloquea ni reemplaza la autenticación oficial: los servidores que exigen cuenta de Mojang o Microsoft la siguen exigiendo.</p>
                            <p>Para jugar necesitás tener derecho a Minecraft según las políticas de Mojang y Microsoft. Tenés el detalle legal en los <RouterLink to="/terms">Términos y condiciones</RouterLink>.</p>
                        </FaqItem>
                        <FaqItem pregunta="¿Cómo gestiono mods y modpacks en StepLauncher?">
                            <p>Con Modrinth dentro del launcher: buscá e instalá mods, modpacks, shaders y paquetes de recursos sin salir de la app. Cada instancia lleva su modloader (Forge, Fabric, Quilt o NeoForge) y su configuración propia. Si querés experimentar, cloná la instancia y probá combinaciones sin miedo a romper tu perfil principal.</p>
                        </FaqItem>
                    </FaqGroup>
                    <FaqGroup id="cuentas" titulo="Cuentas y Yggdrasil">
                        <FaqItem pregunta="¿Qué es el Yggdrasil de StepLauncher?">
                            <p>Es un servicio público de autenticación compatible con el protocolo abierto Yggdrasil, que verifica tu identidad dentro del juego: nombre, UUID y texturas (skin y capa). Cualquier servidor puede adoptarlo y cualquier jugador con cuenta puede usarlo.</p>
                            <p>Además, el launcher permite agregar servidores Yggdrasil externos de terceros, cada uno con sus propias cuentas y reglas.</p>
                        </FaqItem>
                        <FaqItem pregunta="¿Cómo funcionan el modo offline y el modo online?">
                            <p><b>En modo offline</b> jugás sin verificar tu identidad con nadie: tu cuenta no se comparte con ningún servicio.</p>
                            <p><b>En modo online</b> iniciás sesión contra el servicio Yggdrasil que vos elijas (StepLauncher u otro externo). El launcher recupera los datos de tu cuenta de ese servicio (nombre, UUID y texturas) y, al lanzar el juego con authlib-injector, es el propio Minecraft el que comparte tu identidad con el servidor asociado para verificarte y dejarte entrar.</p>
                        </FaqItem>
                        <FaqItem pregunta="¿Puedo usar mi skin y mi capa?">
                            <p>Sí. Las subís desde tu panel y se muestran en 3D en la web. En modo online viajan con tu identidad Yggdrasil y se ven en los servidores compatibles que acepten StepLauncher.</p>
                        </FaqItem>
                        <FaqItem pregunta="¿Mis datos se comparten con Mojang o Microsoft?">
                            <p>No. StepLauncher no envía tu cuenta ni tus datos personales a Mojang o Microsoft, y nunca verifica tu identidad con ellos. En modo online tu identidad solo llega al servicio Yggdrasil que elegiste y al servidor donde jugás.</p>
                            <p>Las descargas desde fuentes oficiales son como cualquier otra descarga: tu dispositivo se conecta a esa infraestructura para bajar los archivos, nada más. Tenés el detalle en la <RouterLink to="/privacy">Política de privacidad</RouterLink>.</p>
                        </FaqItem>
                        <FaqItem pregunta="¿Puedo agregar servidores Yggdrasil externos?">
                            <p>Sí, agregando su dirección en el launcher. Esos servicios son ajenos: no los operamos, no controlamos su seguridad o disponibilidad y no respondemos por lo que hagan. Usalos bajo tu responsabilidad y leé sus reglas antes de conectarte.</p>
                            <p><b>Nunca ingreses tu contraseña de StepLauncher en un servicio externo:</b> cada uno usa sus propios datos, que quedan guardados solo en tu dispositivo.</p>
                        </FaqItem>
                    </FaqGroup>
                    <FaqGroup id="datos" titulo="Datos y cuenta">
                        <FaqItem pregunta="¿Dónde están mis datos y cómo los borro?">
                            <p>Se guarda lo mínimo necesario: correo, usuario, perfil, texturas y sesiones de juego. Viven en Supabase con reglas que hacen que cada usuario solo pueda tocar su propia información, y la sesión queda guardada únicamente en tu navegador o dispositivo.</p>
                            <p>Desde tu panel ves y editás todo cuando quieras y cerrás tus sesiones, incluyendo las de juego. Si querés borrar tu cuenta completa con todos sus datos asociados, pedilo desde GitHub y lo procesamos. El detalle está en la <RouterLink to="/privacy">Política de privacidad</RouterLink> y las reglas de uso en los <RouterLink to="/terms">Términos y condiciones</RouterLink>.</p>
                        </FaqItem>
                    </FaqGroup>
                </template>
                <div v-if="buscando && totalVisibles === 0" class="StateCard">
                    <b>Nada coincide con esa búsqueda.</b>
                    <p>Probá con otras palabras, como "mods", "offline" o "premium".</p>
                    <button class="Btn" type="button" @click="limpiar">Limpiar búsqueda</button>
                </div>
                <a v-else class="Top" href="#faq-top">↑ Volver arriba</a>
            </div>
        </div>
    </div>
</template>

<style scoped lang="scss">
.Faq{
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
        .Badges{
            display: flex;
            justify-content: center;
            align-items: center;
            gap: .5rem;
            .BadgeMain,
            .Badge{
                font-size: .7rem;
                padding: .25rem .65rem;
                border-radius: 99rem;
                border: 1px solid #ffffff25;
                background: #ffffff0d;
            }
            .BadgeMain{
                background: #fff;
                color: #000;
                font-weight: 700;
                border-color: #fff;
            }
        }
    }
    .Temas{
        display: grid;
        grid-template-columns: repeat(4, 1fr);
        gap: .8rem;
        width: 100%;
        max-width: 64rem;
        margin-bottom: 2rem;
        .TemaCard{
            display: flex;
            align-items: flex-start;
            gap: .7rem;
            padding: .9rem 1rem;
            border-radius: .7rem;
            border: 1px solid #ffffff14;
            background: #ffffff08;
            color: #fff;
            text-decoration: none;
            transition: border-color 150ms, background 150ms, transform 150ms;
            &:hover{
                border-color: #ffffff2e;
                background: #ffffff0d;
                transform: translateY(-2px);
            }
            .CardIcon{
                display: flex;
                justify-content: center;
                align-items: center;
                width: 2.4rem;
                height: 2.4rem;
                border-radius: .55rem;
                background: #ffffff10;
                border: 1px solid #ffffff14;
                flex-shrink: 0;
                svg{
                    width: 1.2rem;
                    height: 1.2rem;
                }
            }
            .CardTxt{
                display: flex;
                flex-direction: column;
                align-items: flex-start;
                text-align: left;
                gap: .2rem;
                flex: 1;
                min-width: 0;
                b{
                    font-size: .82rem;
                    font-family: 'Lexend';
                }
                small{
                    font-size: .7rem;
                    opacity: .55;
                    line-height: 1.5;
                }
            }
            i{
                display: flex;
                justify-content: center;
                align-items: center;
                min-width: 1.4rem;
                height: 1.4rem;
                padding: 0 .3rem;
                border-radius: 99rem;
                background: #fff;
                color: #000;
                font-style: normal;
                font-size: .65rem;
                font-weight: 800;
                flex-shrink: 0;
            }
        }
    }
    .Layout{
        display: flex;
        align-items: flex-start;
        gap: 2rem;
        width: 100%;
        max-width: 64rem;
        .Side{
            position: sticky;
            top: 5.5rem;
            align-self: flex-start;
            display: flex;
            flex-direction: column;
            align-items: stretch;
            gap: .3rem;
            width: 13rem;
            flex-shrink: 0;
            max-height: calc(100dvh - 7rem);
            overflow-y: auto;
            b{
                font-size: .7rem;
                text-transform: uppercase;
                letter-spacing: .12em;
                opacity: .4;
                font-family: 'Lexend';
                margin-bottom: .3rem;
                padding-left: .8rem;
            }
            a{
                display: flex;
                justify-content: space-between;
                align-items: center;
                gap: .5rem;
                padding: .55rem .8rem;
                border-radius: .55rem;
                font-size: .82rem;
                color: #ffffffa6;
                text-decoration: none;
                transition: background 150ms, color 150ms;
                &:hover{
                    background: #ffffff10;
                    color: #fff;
                }
                &.active{
                    background: #ffffff10;
                    color: #fff;
                    box-shadow: inset 2px 0 0 #fff;
                }
                i{
                    display: flex;
                    justify-content: center;
                    align-items: center;
                    min-width: 1.4rem;
                    height: 1.4rem;
                    padding: 0 .3rem;
                    border-radius: 99rem;
                    background: #ffffff10;
                    border: 1px solid #ffffff18;
                    font-style: normal;
                    font-size: .65rem;
                    font-weight: 700;
                }
            }
        }
        .Main{
            display: flex;
            flex-direction: column;
            gap: 1.5rem;
            flex: 1;
            min-width: 0;
            .Search{
                display: flex;
                align-items: center;
                gap: .45rem;
                padding: .55rem 1rem;
                border-radius: .7rem;
                border: 1px solid #ffffff25;
                background: #ffffff08;
                transition: border-color 150ms;
                &:focus-within{
                    border-color: #ffffff50;
                }
                svg{
                    width: 1.1rem;
                    height: 1.1rem;
                    flex-shrink: 0;
                    opacity: .55;
                }
                input{
                    width: 100%;
                    border: 0;
                    outline: 0;
                    background: transparent;
                    color: #fff;
                    font-size: .85rem;
                    font-family: inherit;
                    &::placeholder{
                        color: #ffffff55;
                    }
                    &::-webkit-search-cancel-button{
                        cursor: pointer;
                    }
                }
            }
            .ResultCount{
                margin: 0;
                font-size: .78rem;
                opacity: .5;
            }
            .Top{
                align-self: center;
                font-size: .78rem;
                color: #ffffffa6;
                text-decoration: none;
                padding: .5rem 1rem;
                border-radius: 99rem;
                border: 1px solid transparent;
                transition: background 150ms, color 150ms, border-color 150ms;
                &:hover{
                    background: #ffffff10;
                    color: #fff;
                    border-color: #ffffff18;
                }
            }
            .StateCard{
                display: flex;
                flex-direction: column;
                align-items: center;
                gap: 1rem;
                padding: 1.5rem 2rem;
                border-radius: .9rem;
                border: 1px solid #ffffff18;
                b{
                    font-size: .85rem;
                    font-family: 'Lexend';
                }
                p{
                    margin: 0;
                    font-size: .8rem;
                    opacity: .55;
                    text-align: center;
                }
                .Btn{
                    padding: .6rem 1.3rem;
                    border-radius: .5rem;
                    font-size: .85rem;
                    font-weight: 600;
                    font-family: inherit;
                    background: #ffffff10;
                    color: #fff;
                    border: 1px solid #ffffff25;
                    cursor: pointer;
                    &:hover{
                        background: #ffffff1c;
                    }
                }
            }
        }
    }
}
@media (max-width: 1000px){
    .Faq{
        .Temas{
            grid-template-columns: 1fr 1fr;
        }
        .Layout{
            flex-direction: column;
            .Side{
                position: static;
                width: 100%;
                max-height: none;
                flex-direction: row;
                align-items: center;
                gap: .5rem;
                overflow-x: auto;
                padding-bottom: .3rem;
                b{
                    display: none;
                }
                a{
                    flex-shrink: 0;
                    border: 1px solid #ffffff25;
                    background: #ffffff08;
                    border-radius: 99rem;
                    i{
                        border: 0;
                        background: transparent;
                        padding: 0;
                    }
                }
            }
        }
    }
}
@media (max-width: 600px){
    .Faq{
        .Temas{
            grid-template-columns: 1fr;
        }
    }
}
</style>
