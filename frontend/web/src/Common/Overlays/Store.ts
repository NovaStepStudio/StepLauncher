import { ref } from 'vue';

export type HeavyPanel = 'instances' | 'shots' | null;

export const heavyPanel = ref<HeavyPanel>(null);

export function openHeavyPanel(panel: Exclude<HeavyPanel, null>): void {
    heavyPanel.value = panel;
}

export function closeHeavyPanel(panel: Exclude<HeavyPanel, null>): void {
    if (heavyPanel.value === panel) heavyPanel.value = null;
}

export const PERSONALIZATION_PREVIEW_EVENT = 'personalization-preview-open';

// --- Pila de diálogos (overlays pequeños sobre la vista) ---

export interface DialogSpec {
    id: number;
    name: string;
    props?: Record<string, unknown>;
    listeners?: Record<string, (...args: any[]) => void>;
}

let dialogSeq = 0;

export const dialogs = ref<DialogSpec[]>([]);

export function openDialog(
    name: string,
    props?: Record<string, unknown>,
    listeners?: Record<string, (...args: any[]) => void>
): void {
    dialogs.value.push({ id: ++dialogSeq, name, props, listeners });
}

export function closeDialog(id: number): void {
    const index = dialogs.value.findIndex((d) => d.id === id);
    if (index >= 0) dialogs.value.splice(index, 1);
}

export function closeAllDialogs(): void {
    dialogs.value = [];
}

// --- Confirmación global (Promise) ---

export interface AskOptions {
    title: string;
    message?: string;
    confirmLabel?: string;
    cancelLabel?: string;
    danger?: boolean;
    inputLabel?: string;
    inputPlaceholder?: string;
    inputValue?: string;
}

export interface AskResult {
    confirmed: boolean;
    value: string;
}

export const confirmState = ref<{ opts: AskOptions; resolve: (r: AskResult) => void } | null>(null);

export function ask(opts: AskOptions): Promise<AskResult> {
    return new Promise((resolve) => {
        confirmState.value = { opts, resolve };
    });
}

export function resolveConfirm(result: AskResult): void {
    const current = confirmState.value;
    if (!current) return;
    confirmState.value = null;
    current.resolve(result);
}

// --- Visibilidad de los modales de nivel superior (App) ---

export const settingsOpen = ref(false);
export const accountsOpen = ref(false);
export const loginOpen = ref(false);
export const installOpen = ref(false);
export const versionsOpen = ref(false);
export const crashOpen = ref(false);
export const newsOpen = ref(false);
export const welcomeOpen = ref(false);
export const previewOpen = ref(false);

// --- Panel de capturas (abierto desde el panel de instancias o desde el menú) ---

export const shotsInstance = ref<string | null>(null);
export const shotsReturn = ref(false);
