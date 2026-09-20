<script setup lang="ts">
import Hero from './Components/Hero.vue';
import Project from './Components/Project.vue';
import Stack from './Components/Stack.vue';
import Studio from './Components/Studio.vue';
</script>

<template>
    <div class="About">
        <Hero />
        <Project />
        <Stack />
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
