// Caché local de URLs firmadas (1 h): evita pedirlas de nuevo.

import { urlArchivo } from '@/Auth/Api';

export type ConAuth = <T>(llamada: (token: string) => Promise<T>) => Promise<T>;

interface Entrada {
    url: string;
    vence: number;
}

const cache = new Map<string, Entrada>();

export async function urlFirmada(conAuth: ConAuth, id: string): Promise<string> {
    const previa = cache.get(id);
    if (previa && previa.vence > Date.now()) return previa.url;
    const res = await conAuth((token) => urlArchivo(token, id));
    // Margen de 5 minutos antes del vencimiento real.
    cache.set(id, { url: res.url, vence: Date.now() + Math.max(0, res.expiresIn - 300) * 1000 });
    return res.url;
}

export function olvidarUrl(id: string): void {
    cache.delete(id);
}

export function olvidarTodas(): void {
    cache.clear();
}
