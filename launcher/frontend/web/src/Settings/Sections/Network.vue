<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { GetConfig, SetAuthVerify, SetProxy } from '@wailsjs/StepLauncher/internal/Services/Config/configservice';

const authVerify = ref(true);
const proxyEnabled = ref(false);
const proxyHost = ref('');
const proxyPort = ref(8080);
const proxyUser = ref('');
const proxyPass = ref('');
const proxyMsg = ref('');
const proxyMsgOk = ref(true);

async function loadConfig() {
    try {
        const cfg = await GetConfig();
        if (cfg) {
            const mc = cfg.minecraftConfig ?? {};
            authVerify.value = mc.authVerify ?? true;
            proxyEnabled.value = mc.proxyEnabled ?? false;
            proxyHost.value = mc.proxyHost ?? '';
            proxyPort.value = mc.proxyPort ?? 8080;
            proxyUser.value = mc.proxyUser ?? '';
            proxyPass.value = mc.proxyPass ?? '';
        }
    } catch (_e) {}
}

onMounted(() => {
    loadConfig();
});

async function saveAuthVerify() {
    try {
        await SetAuthVerify(authVerify.value);
    } catch (_e) {}
}

async function saveProxy() {
    proxyMsg.value = '';
    // No permitir un proxy "fantasma": activado sin host/puerto válidos
    // rompía todas las peticiones (manifiesto, modloaders) con errores raros.
    if (proxyEnabled.value) {
        const host = proxyHost.value.trim();
        const port = Number(proxyPort.value);
        if (!host) {
            proxyEnabled.value = false;
            proxyMsg.value = 'Escribe la dirección del proxy antes de activarlo.';
            proxyMsgOk.value = false;
            return;
        }
        if (!Number.isFinite(port) || port < 1 || port > 65535) {
            proxyEnabled.value = false;
            proxyMsg.value = 'El puerto debe estar entre 1 y 65535.';
            proxyMsgOk.value = false;
            return;
        }
        proxyHost.value = host;
    }
    try {
        await SetProxy?.(
            proxyEnabled.value,
            proxyHost.value,
            proxyPort.value,
            proxyUser.value,
            proxyPass.value
        );
    } catch (_e) {}
}
</script>

<template>
    <div class="Ss">

        <div class="SsGroup">
            <div class="SsGroupHead">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="8.5" cy="7" r="4"/><path d="M20 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
                <span>Cuenta y conexión</span>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Comprobar cuenta al iniciar sesión</span>
                    <span class="SsDesc">Verifica que tu cuenta sea válida antes de entrar. Si lo apagas, entras más rápido pero puede fallar si tu cuenta cambió.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="authVerify" @change="saveAuthVerify"><span class="SsTgS"></span></label>
                </div>
            </div>
        </div>

        <div class="SsGroup">
            <div class="SsGroupHead">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
                <span>Proxy del juego</span>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Usar proxy para Minecraft y descargas</span>
                    <span class="SsDesc">Si tu red necesita proxy, el juego y el launcher lo usarán para Mojang. Soporta HTTP y SOCKS5.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="proxyEnabled" @change="saveProxy"><span class="SsTgS"></span></label>
                </div>
            </div>
            <template v-if="proxyEnabled">
                <div class="SsRow">
                    <div class="SsInfo">
                        <span class="SsLabel">Dirección del proxy</span>
                        <span class="SsDesc">Host y puerto. Para SOCKS5 escribe socks5://127.0.0.1 (Clash: HTTP 7890, SOCKS 7891).</span>
                    </div>
                    <div class="SsGrid">
                        <input class="SsIn" v-model="proxyHost" placeholder="Ej: 127.0.0.1 o socks5://127.0.0.1" @change="saveProxy">
                        <input class="SsIn" type="number" v-model.number="proxyPort" placeholder="8080" @change="saveProxy">
                    </div>
                </div>
                <div class="SsRow">
                    <div class="SsInfo">
                        <span class="SsLabel">Usuario y contraseña</span>
                        <span class="SsDesc">Solo si tu proxy te los pide. Puedes dejarlo vacío.</span>
                    </div>
                    <div class="SsGrid">
                        <input class="SsIn" v-model="proxyUser" placeholder="Usuario" @change="saveProxy">
                        <input class="SsIn" type="password" v-model="proxyPass" placeholder="Contraseña" @change="saveProxy">
                    </div>
                </div>
            </template>
            <p v-if="proxyMsg" :class="['Ss_ProxyMsg', { error: !proxyMsgOk }]">{{ proxyMsg }}</p>
            <div v-if="proxyEnabled" class="SsTip">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
                <span>Error "malformed HTTP status" = puerto equivocado. Si usas Clash/V2Ray, el HTTP suele ser 7890 y el SOCKS 7891. Prueba con socks5:// delante del host si tu proxy es SOCKS.</span>
            </div>
        </div>

    </div>
</template>

<style scoped lang="scss">
@use '../Styles/General.scss';
</style>
