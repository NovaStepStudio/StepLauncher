/**
 * Bootstrap — punto de entrada público.
 *
 * Main.ts importa { runBootstrap, bootstrapState } y orquesta el arranque.
 * App.vue solo consume el estado para el SplashScreen y no inicia nada.
 */

export * from './Types';
export * from './State';
export * from './Runner';
export { default as SplashScreen } from './SplashScreen.vue';
