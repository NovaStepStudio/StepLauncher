<script setup lang="ts">
import Hero from './Components/Hero.vue';
import Instances from './Components/Instances.vue';
import Mods from './Components/Mods.vue';
import Music from './Components/Music.vue';
import Performance from './Components/Performance.vue';
import Customize from './Components/Customize.vue';
import OpenSource from './Components/OpenSource.vue';
import DownloadCTA from './Components/DownloadCTA.vue';
</script>

<template>
    <div class="Home">
        <Hero />
        <Instances />
        <Mods />
        <Music />
        <Performance />
        <Customize />
        <OpenSource />
        <DownloadCTA />
    </div>
</template>

<style scoped lang="scss">
.Home{
    display:flex;
    flex-direction:column;
    background:#000;
}
</style>
