<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { IconChevronDown, IconTag } from '@tabler/icons-vue';
import { channelOf, useReleases, type GithubRelease } from '@/Common/Composables/useReleases';
import ReleaseCard from './ReleaseCard.vue';

const { stable, betas, alphas, loading, load } = useReleases();
const showPrevias = ref(false);

// Solo las 2 previas más nuevas (sin repetir el tag estable de arriba).
// El resto vive en el historial (/changelog).
const previas = computed(() => {
    const todas = [...betas.value, ...alphas.value].filter((r) => r.tag !== stable.value?.tag);
    todas.sort((a, b) => Date.parse(b.publishedAt || '') - Date.parse(a.publishedAt || ''));
    return todas.slice(0, 2);
});

function badgeDe(r: GithubRelease): string {
    return channelOf(r.tag, r.prerelease) === 'alpha' ? 'Alpha' : 'Beta';
}

onMounted(() => load());
</script>

<template>
    <div class="Prerelease" v-reveal>
        <button class="Toggle" @click="showPrevias = !showPrevias">
            <IconTag stroke="2" />
            También quiero ver las versiones previas
            <IconChevronDown stroke="2" :class="{ open: showPrevias }" />
        </button>
        <div class="Collapse" :class="{ open: showPrevias }">
            <div class="CollapseInner">
                <div class="BetaBody">
                    <p v-if="loading" class="State">Buscando previas en GitHub...</p>
                    <template v-else-if="previas.length > 0">
                        <ReleaseCard v-for="p in previas" :key="p.tag" :release="p" :badge="badgeDe(p)" />
                        <RouterLink class="AllHistory" to="/changelog">Ver todas en el historial</RouterLink>
                    </template>
                    <p v-else class="State">No hay previas por ahora. Cuando haya, aparecen acá.</p>
                    <small class="Warn">Las betas y alphas pueden tener errores: son para probar lo nuevo antes que nadie.</small>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped lang="scss">
.Prerelease{
    width: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 0 1.5rem 4rem 1.5rem;
    gap: 1.2rem;
    .Collapse{
        width: 100%;
    }
    .Toggle{
        display:flex;
        justify-content:center;
        align-items:center;
        gap: .5rem;
        padding: .6rem 1.2rem;
        border-radius: .5rem;
        border: 1px solid #ffffff25;
        background: #ffffff08;
        color: #fff;
        font-size: .82rem;
        font-weight: 600;
        cursor: pointer;
        transition: background 150ms;
        svg{
            width: 1rem;
            height: 1rem;
            &:last-child{
                transition: rotate 200ms;
                &.open{
                    rotate: 180deg;
                }
            }
        }
        &:hover{
            background: #ffffff14;
        }
    }
    .BetaBody{
        display:flex;
        flex-direction:column;
        align-items:center;
        gap: .8rem;
        width: 100%;
        .State{
            font-size: .85rem;
            opacity: .55;
        }
        .AllHistory{
            font-size: .78rem;
            color: #ffffffa6;
            text-decoration: none;
            &:hover{
                color: #fff;
            }
        }
        .Warn{
            font-size: .72rem;
            opacity: .45;
        }
    }
}
</style>
