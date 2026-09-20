<script setup lang="ts">
import { computed, ref } from 'vue';
import { marked } from 'marked';
import { IconChevronDown, IconCalendar, IconTag } from '@tabler/icons-vue';
import { channelOf, formatDate, type GithubRelease } from '@/Common/Composables/useReleases';

const props = defineProps<{
    release: GithubRelease;
    open?: boolean;
}>();

const expanded = ref(!!props.open);

const channel = computed(() => channelOf(props.release.tag, props.release.prerelease));
const channelLabel = computed(() =>
    channel.value === 'stable' ? 'Estable' : channel.value === 'beta' ? 'Beta' : 'Alpha',
);
const html = computed(() => marked.parse(props.release.body || '') as string);
</script>

<template>
    <div class="Note" :class="{ open: expanded }">
        <button class="NoteHead" type="button" @click="expanded = !expanded" :aria-expanded="expanded">
            <span class="BadgeMain">{{ channelLabel }}</span>
            <span class="NoteTitle">
                <b>{{ release.tag }}</b>
                <small>
                    <i><IconTag stroke="2" /> {{ release.name }}</i>
                    <i><IconCalendar stroke="2" /> {{ formatDate(release.publishedAt) }}</i>
                </small>
            </span>
            <IconChevronDown stroke="2" :class="{ open: expanded }" />
        </button>
        <div class="Collapse" :class="{ open: expanded }">
            <div class="CollapseInner">
                <div v-if="html" class="NoteBody md-body" v-html="html"></div>
                <p v-else class="NoBody">Este release no trae notas escritas.</p>
            </div>
        </div>
    </div>
</template>

<style scoped lang="scss">
.Note{
    display: flex;
    flex-direction: column;
    border-radius: .7rem;
    border: 1px solid #ffffff14;
    background: #0b0b0d;
    overflow: hidden;
    transition: border-color 200ms;
    &.open{
        border-color: #ffffff2e;
    }
    .NoteHead{
        display: flex;
        align-items: center;
        gap: .8rem;
        width: 100%;
        padding: .9rem 1rem;
        border: 0;
        background: transparent;
        color: #fff;
        font-family: inherit;
        text-align: left;
        cursor: pointer;
        transition: background 150ms;
        &:hover{
            background: #ffffff08;
        }
        .BadgeMain{
            flex-shrink: 0;
            font-size: .62rem;
            font-weight: 700;
            text-transform: uppercase;
            letter-spacing: .08em;
            padding: .25rem .6rem;
            border-radius: 99rem;
            background: #fff;
            color: #000;
            border: 1px solid #fff;
        }
        .NoteTitle{
            display: flex;
            flex-direction: column;
            align-items: flex-start;
            gap: .15rem;
            min-width: 0;
            flex: 1;
            b{
                font-family: 'Lexend';
                font-size: .95rem;
            }
            small{
                display: flex;
                align-items: center;
                flex-wrap: wrap;
                gap: .35rem .8rem;
                font-size: .72rem;
                opacity: .55;
                i{
                    display: flex;
                    align-items: center;
                    gap: .3rem;
                    font-style: normal;
                    svg{
                        width: .9rem;
                        height: .9rem;
                    }
                }
            }
        }
        svg{
            width: 1.1rem;
            height: 1.1rem;
            flex-shrink: 0;
            opacity: .7;
            transition: rotate 200ms;
            &.open{
                rotate: 180deg;
            }
        }
    }
    .NoteBody{
        padding: 0 1.1rem 1.1rem 1.1rem;
    }
    .NoBody{
        margin: 0;
        padding: 0 1.1rem 1.1rem 1.1rem;
        font-size: .78rem;
        opacity: .45;
    }
}
</style>
