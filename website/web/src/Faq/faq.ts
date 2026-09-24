// Infraestructura del FAQ: contexto compartido, sin preguntas.
// Las preguntas viven en la plantilla (Index.vue) como bloques
// <FaqItem>: para agregar una basta copiar un bloque, el buscador,
// el índice y los conteos salen solos del texto.
import type { Component, InjectionKey, Ref } from 'vue';
import { IconBox, IconKey, IconListDetails, IconShieldCheck } from '@tabler/icons-vue';

export interface FaqItemDatos {
    grupo: string;
    texto: string;
}

export interface FaqContexto {
    consulta: Ref<string>;
    grupos: Map<string, string>;
    items: Map<symbol, FaqItemDatos>;
}

export const FAQ_KEY: InjectionKey<FaqContexto> = Symbol('faq');

// Id del grupo al que pertenece cada pregunta (lo provee FaqGroup).
export const GRUPO_KEY: InjectionKey<string> = Symbol('faq-grupo');

// Cantidad de respuestas visibles de un grupo según la búsqueda actual.
export function visiblesDelGrupo(ctx: FaqContexto, grupo: string): number {
    const q = ctx.consulta.value.trim().toLowerCase();
    let n = 0;
    for (const item of ctx.items.values()) {
        if (item.grupo === grupo && (q === '' || item.texto.includes(q))) n++;
    }
    return n;
}

// Ficha visual de cada tema (tarjetas del índice). Si se agrega un grupo
// nuevo sin ficha, usa el icono y la descripción genéricos.
export interface FaqMeta {
    icono: Component;
    descripcion: string;
}

const META: Record<string, FaqMeta> = {
    general: { icono: IconListDetails, descripcion: 'Costo, seguridad, sistemas y por qué elegirnos.' },
    juego: { icono: IconBox, descripcion: 'Versiones, descargas oficiales, Java y mods.' },
    cuentas: { icono: IconKey, descripcion: 'Yggdrasil, offline y online, skins y externos.' },
    datos: { icono: IconShieldCheck, descripcion: 'Qué guardamos, dónde vive y cómo borrarlo.' },
};

const META_GENERICA: FaqMeta = { icono: IconListDetails, descripcion: '' };

export function metaDelGrupo(id: string): FaqMeta {
    return META[id] ?? META_GENERICA;
}
