/**
 * Bootstrap — punto de entrada público.
 *
 * Main.ts importa { runBootstrap, bootstrapState } y orquesta el arranque.
 * App.vue solo consume el estado para el SplashScreen y no inicia nada.
 */

export * from './types';
export * from './state';
export * from './runner';
export { default as SplashScreen } from './SplashScreen.vue';
