<script setup lang="ts">
import { computed, inject, onMounted, onUnmounted, provide } from 'vue';
import { FAQ_KEY, GRUPO_KEY, visiblesDelGrupo } from '@/Faq/faq';

// Un tema del FAQ: agrupa bloques <FaqItem>. Para agregar un tema basta
// copiar un bloque como este; el índice lateral lo toma solo.
const props = defineProps<{
    id: string;
    titulo: string;
}>();

const ctx = inject(FAQ_KEY, null);
provide(GRUPO_KEY, props.id);

onMounted(() => ctx?.grupos.set(props.id, props.titulo));
onUnmounted(() => ctx?.grupos.delete(props.id));

const visibles = computed(() => (ctx ? visiblesDelGrupo(ctx, props.id) : 0));
const mostrar = computed(() => !ctx || ctx.consulta.value.trim() === '' || visibles.value > 0);
</script>

<template>
    <section v-show="mostrar" :id="`grupo-${id}`" class="Group">
        <div class="GroupHead">
            <h2>{{ titulo }}</h2>
            <span class="Count">{{ visibles }} {{ visibles === 1 ? 'respuesta' : 'respuestas' }}</span>
        </div>
        <div class="Items">
            <slot />
        </div>
    </section>
</template>

<style scoped lang="scss">
.Group{
    display: flex;
    flex-direction: column;
    gap: .8rem;
    scroll-margin-top: 5rem;
    .GroupHead{
        display: flex;
        align-items: center;
        gap: .6rem;
        h2{
            margin: 0;
            font-family: 'Lexend';
            font-size: 1.15rem;
            font-weight: 600;
        }
        .Count{
            font-size: .65rem;
            font-weight: 700;
            padding: .2rem .6rem;
            border-radius: 99rem;
            border: 1px solid #ffffff25;
            background: #ffffff0d;
            opacity: .75;
            white-space: nowrap;
        }
    }
    .Items{
        display: flex;
        flex-direction: column;
        gap: .8rem;
        counter-reset: faq;
    }
}
</style>
