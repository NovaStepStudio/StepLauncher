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
    // Al activar se valida, pero SIN apagar el interruptor: los campos de
    // dirección siempre están visibles y el usuario puede estar a mitad de
    // la configuración. Solo se persiste cuando los datos están completos.
    if (proxyEnabled.value) {
        const host = proxyHost.value.trim();
        const port = Number(proxyPort.value);
        if (!host) {
            proxyMsg.value = 'Falta la dirección del proxy (ej. 127.0.0.1 + puerto 7890): todavía no se guardó nada.';
            proxyMsgOk.value = false;
            return;
        }
        if (!Number.isFinite(port) || port < 1 || port > 65535) {
            proxyMsg.value = 'El puerto tiene que estar entre 1 y 65535: todavía no se guardó nada.';
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
                    <span class="SsLabel">Validar la cuenta al iniciar sesión</span>
                    <span class="SsDesc">Comprueba con el servidor que tu sesión siga válida antes de entrar. Apagalo para entrar más rápido, pero puede fallar si tu cuenta cambió.</span>
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
                    <span class="SsLabel">Proxy solo para Minecraft</span>
                    <span class="SsDesc">El juego sale por este proxy al lanzarse; el launcher siempre va directo. Soporta HTTP y SOCKS5.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg"><input type="checkbox" v-model="proxyEnabled" @change="saveProxy"><span class="SsTgS"></span></label>
                </div>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Servidor proxy</span>
                    <span class="SsDesc">Host y puerto. Para SOCKS anteponé socks5:// al host (ej. Clash: HTTP 7890, SOCKS 7891).</span>
                </div>
                <div class="SsGrid">
                    <input class="SsIn" v-model="proxyHost" placeholder="Ej: 127.0.0.1 o socks5://127.0.0.1" @change="saveProxy">
                    <input class="SsIn" type="number" v-model.number="proxyPort" placeholder="8080" @change="saveProxy">
                </div>
            </div>
            <template v-if="proxyEnabled">
                <div class="SsRow">
                    <div class="SsInfo">
                    <span class="SsLabel">Credenciales del proxy</span>
                    <span class="SsDesc">Solo si tu proxy exige usuario y contraseña. Si no, dejalo vacío.</span>
                    </div>
                    <div class="SsGrid">
                        <input class="SsIn" v-model="proxyUser" placeholder="Usuario" @change="saveProxy">
                        <input class="SsIn" type="password" v-model="proxyPass" placeholder="Contraseña" @change="saveProxy">
                    </div>
                </div>
            </template>
            <p v-if="proxyMsg" :class="['Ss_ProxyMsg', { error: !proxyMsgOk }]">{{ proxyMsg }}</p>
            <div v-if="!proxyEnabled" class="SsTip">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
                <span>Completá la dirección primero y después activá el interruptor: solo se guarda cuando los datos están completos.</span>
            </div>
            <div v-if="proxyEnabled" class="SsTip">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
                <span>¿"malformed HTTP status"? Casi siempre es el puerto: probá 7890 para HTTP o 7891 con socks5:// delante si tu proxy es SOCKS (Clash/V2Ray).</span>
            </div>
        </div>

    </div>
</template>

<style scoped lang="scss">
@use '../Styles/General.scss';
</style>
