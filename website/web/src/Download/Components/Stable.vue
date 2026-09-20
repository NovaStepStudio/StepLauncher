<script setup lang="ts">
import { computed, onMounted } from 'vue';
import { IconDownload } from '@tabler/icons-vue';
import { useReleases } from '@/Common/Composables/useReleases';
import ReleaseCard from './ReleaseCard.vue';

const { stable, prerelease, loading, error, load } = useReleases();
const shown = computed(() => stable.value ?? prerelease.value);
const isBeta = computed(() => !stable.value && !!prerelease.value);

onMounted(() => load());
</script>

<template>
    <div class="Stable" id="stable" v-reveal>
        <div class="Head">
            <div class="Tag">
                <IconDownload stroke="2" />
                <span>Último release</span>
            </div>
            <div class="Description">
                <h2>La versión recomendada.</h2>
                <p>Archivos publicados en GitHub, listos para tu sistema operativo.</p>
            </div>
        </div>
        <p v-if="loading" class="State">Buscando la última versión en GitHub...</p>
        <div v-else-if="error" class="StateCard">
            <b>{{ error }}</b>
            <a class="Btn" href="https://github.com/NovaStepStudio/StepLauncher/releases" target="_blank" rel="noopener">Ir a los releases</a>
        </div>
        <ReleaseCard v-else-if="shown" :release="shown" :badge="isBeta ? 'Beta' : 'Estable'" />
        <div v-else class="StateCard">
            <b>Todavía no hay releases publicados.</b>
            <a class="Btn" href="https://github.com/NovaStepStudio/StepLauncher" target="_blank" rel="noopener">Ver el repo</a>
        </div>
    </div>
</template>

<style scoped lang="scss">
.Stable{
    width: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 3rem 1.5rem 2rem;
    gap: 1.5rem;
    div{
        display:flex;
        justify-content:center;
        align-items:center;
    }
    .Head{
        flex-direction:column;
        gap: 1rem;
        text-align:center;
        .Tag{
            gap: .4rem;
            font-size: .7rem;
            text-transform: uppercase;
            letter-spacing: .12em;
            padding: .25rem .65rem;
            border-radius: 99rem;
            border: 1px solid #ffffff25;
            background: #ffffff0d;
            opacity: .85;
            svg{
                width: .95rem;
                height: .95rem;
            }
        }
        .Description{
            flex-direction:column;
            gap: .5rem;
            h2{
                margin: 0;
                font-family: 'Lexend';
                font-size: 1.6rem;
                font-weight: 600;
            }
            p{
                margin: 0;
                font-size: .9rem;
                line-height: 1.6;
                opacity: .65;
            }
        }
    }
    .State{
        font-size: .85rem;
        opacity: .55;
    }
    .StateCard{
        flex-direction:column;
        gap: 1rem;
        padding: 1.5rem 2rem;
        border-radius: .9rem;
        border: 1px solid #ffffff18;
        background: transparent;
        b{
            font-size: .85rem;
            font-family: 'Lexend';
        }
        .Btn{
            padding: .6rem 1.3rem;
            border-radius: .5rem;
            font-size: .85rem;
            font-weight: 600;
            text-decoration: none;
            background: #ffffff10;
            color: #fff;
            border: 1px solid #ffffff25;
            &:hover{
                background: #ffffff1c;
            }
        }
    }
}
</style>
