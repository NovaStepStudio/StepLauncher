/// <reference types="vite/client" />

// Declaraciones de tipos para los imports de assets (imagenes, fuentes, etc.)
// que Vite procesa y hashea: sin esto vue-tsc fallaria con imports de .png,
// .gif, .woff2, etc.
declare module '*.png' {
    const src: string;
    export default src;
}
declare module '*.jpg' {
    const src: string;
    export default src;
}
declare module '*.jpeg' {
    const src: string;
    export default src;
}
declare module '*.webp' {
    const src: string;
    export default src;
}
declare module '*.gif' {
    const src: string;
    export default src;
}
declare module '*.svg' {
    const src: string;
    export default src;
}
declare module '*.woff' {
    const src: string;
    export default src;
}
declare module '*.woff2' {
    const src: string;
    export default src;
}
