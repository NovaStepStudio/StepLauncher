<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { autoRefresh, loadAccounts, setAutoRefresh } from '@/Accounts/Store';

const busy = ref(false);
const msg = ref('');
const msgOk = ref(true);

onMounted(() => {
    loadAccounts();
});

async function toggle(v: boolean) {
    if (busy.value) return;
    busy.value = true;
    msg.value = '';
    const err = await setAutoRefresh(v);
    busy.value = false;
    if (err) {
        msg.value = err;
        msgOk.value = false;
    }
}
</script>

<template>
    <div class="Ss">
        <div class="SsGroup">
            <div class="SsGroupHead">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
                <span>Sesiones</span>
            </div>
            <div class="SsRow">
                <div class="SsInfo">
                    <span class="SsLabel">Renovar sesiones al iniciar el launcher</span>
                    <span class="SsDesc">Al abrir StepLauncher, comprueba las sesiones guardadas y renueva automáticamente las que hayan caducado. Si lo desactivas y una sesión expira, tendrás que volver a iniciar sesión antes de jugar.</span>
                </div>
                <div class="SsCtrl">
                    <label class="SsTg">
                        <input type="checkbox" :checked="autoRefresh" :disabled="busy" @change="toggle(($event.target as HTMLInputElement).checked)" />
                        <span class="SsTgS"></span>
                    </label>
                </div>
            </div>
            <p v-if="msg" :class="['Ss_AcctMsg', { error: !msgOk }]">{{ msg }}</p>
        </div>
    </div>
</template>

<style scoped lang="scss">
@use '../Styles/Accounts.scss';
</style>
