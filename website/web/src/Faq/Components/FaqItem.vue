<script setup lang="ts">
import { computed, inject, onMounted, onUnmounted, ref } from 'vue';
import { FAQ_KEY, GRUPO_KEY } from '@/Faq/faq';

// Una pregunta con su respuesta siempre visible. Para agregar una basta
// copiar un bloque como este dentro del grupo que corresponda: el texto
// (pregunta + respuesta) entra solo al buscador, sin claves ni ids.
const props = defineProps<{
    pregunta: string;
}>();

const ctx = inject(FAQ_KEY, null);
const grupo = inject(GRUPO_KEY, '');
const bloque = ref<HTMLElement | null>(null);
const clave = Symbol('faq-item');
const texto = ref('');

onMounted(() => {
    texto.value = `${props.pregunta} ${(bloque.value?.textContent ?? '').replace(/\s+/g, ' ')}`.toLowerCase();
    ctx?.items.set(clave, { grupo, texto: texto.value });
});
onUnmounted(() => ctx?.items.delete(clave));

const visible = computed(() => {
    if (!ctx) return true;
    const q = ctx.consulta.value.trim().toLowerCase();
    return q === '' || texto.value.includes(q);
});
</script>

<template>
    <article v-show="visible" ref="bloque" class="Item">
        <h3>{{ pregunta }}</h3>
        <div class="ItemBody">
            <slot />
        </div>
    </article>
</template>

<style scoped lang="scss">
.Item{
    display: flex;
    flex-direction: column;
    gap: .5rem;
    padding: 1.1rem 1.2rem;
    border-radius: .7rem;
    border: 1px solid #ffffff14;
    background: #ffffff08;
    counter-increment: faq;
    transition: border-color 150ms, background 150ms;
    &:hover{
        border-color: #ffffff2e;
        background: #ffffff0d;
    }
    h3{
        display: flex;
        align-items: baseline;
        gap: .6rem;
        margin: 0;
        font-family: 'Lexend';
        font-size: .92rem;
        font-weight: 600;
        line-height: 1.5;
        &::before{
            content: counter(faq, decimal-leading-zero);
            flex-shrink: 0;
            font-size: .7rem;
            font-weight: 700;
            opacity: .35;
        }
    }
    .ItemBody{
        p{
            margin: 0 0 .6rem 0;
            font-size: .82rem;
            line-height: 1.7;
            opacity: .7;
            &:last-child{
                margin-bottom: 0;
            }
        }
        ul{
            margin: .5rem 0 .6rem 0;
            padding-left: 1.2rem;
            display: flex;
            flex-direction: column;
            gap: .4rem;
            &:last-child{
                margin-bottom: 0;
            }
            li{
                font-size: .82rem;
                line-height: 1.7;
                opacity: .7;
                b{
                    opacity: 1;
                }
            }
        }
        a{
            color: #fff;
        }
    }
}
</style>
