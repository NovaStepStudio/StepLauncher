<script setup lang="ts">
import Header from './Common/Header.vue';
import Footer from './Common/Footer.vue';
</script>

<template>
  <Header />
  <RouterView v-slot="{ Component }">
    <Transition name="sl-page" mode="out-in">
      <component :is="Component" />
    </Transition>
  </RouterView>
  <Footer />
</template>

<style scoped lang="scss"></style>
