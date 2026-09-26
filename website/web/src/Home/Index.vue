<script setup lang="ts">
import Hero from './Components/Hero.vue';
import Instances from './Components/Instances.vue';
import Mods from './Components/Mods.vue';
import Music from './Components/Music.vue';
import Performance from './Components/Performance.vue';
import Customize from './Components/Customize.vue';
import OpenSource from './Components/OpenSource.vue';
import DownloadCTA from './Components/DownloadCTA.vue';
import BadApple from './Components/BadApple.vue';
import { useBadApple } from './Composables/useBadApple';

// Easter egg Bad Apple: ?apple, Konami o 5 clics en el logo del Hero.
const { abierto, cerrar, clicsLogo } = useBadApple();
</script>

<template>
    <div class="Home">
        <Hero @clic-logo="clicsLogo" />
        <Instances />
        <Mods />
        <Music />
        <Performance />
        <Customize />
        <OpenSource />
        <DownloadCTA />
        <BadApple v-if="abierto" @cerrar="cerrar" />
    </div>
</template>

<style scoped lang="scss">
.Home{
    display:flex;
    flex-direction:column;
    background:#000;
}
</style>
