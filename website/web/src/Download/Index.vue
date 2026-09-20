<script setup lang="ts">
import Hero from './Components/Hero.vue';
import Stable from './Components/Stable.vue';
import Prerelease from './Components/Prerelease.vue';
import Steps from './Components/Steps.vue';
</script>

<template>
    <div class="Download">
        <Hero />
        <Stable />
        <Prerelease />
        <Steps />
    </div>
</template>

<style scoped lang="scss">
.Download{
    display:flex;
    flex-direction:column;
    background:#000;
}
</style>
