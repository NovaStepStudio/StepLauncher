<script setup lang="ts">
import Hero from './Components/Hero.vue';
import Project from './Components/Project.vue';
import HowItWorks from './Components/HowItWorks.vue';
import Yggdrasil from './Components/Yggdrasil.vue';
import Stack from './Components/Stack.vue';
import Ecosystem from './Components/Ecosystem.vue';
import Studio from './Components/Studio.vue';
</script>

<template>
    <div class="About">
        <Hero />
        <Project />
        <HowItWorks />
        <Yggdrasil />
        <Stack />
        <Ecosystem />
        <Studio />
    </div>
</template>

<style scoped lang="scss">
.About{
    display:flex;
    flex-direction:column;
    background:#000;
}
</style>
