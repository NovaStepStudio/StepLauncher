import { ref, type Ref } from 'vue';
// API centralizada: la versión vive solo en Go `engineconfig.AppVersion`.
// El frontend la obtiene vía `GetAppVersion()` / `GetLauncherVersionInfo()` para no hardcodear por release.
import { GetAppVersion } from '@wailsjs/StepLauncher/internal/Services/System/systemservice';

function wailsFlagsVersion(): string | undefined {
    const flags = (window as unknown as { _wails?: { flags?: Record<string, unknown> } })._wails?.flags;
    const v = flags?.version;
    return typeof v === 'string' && v.length > 0 ? v : undefined;
}

export function useAppVersion(): { appVersion: Ref<string> } {
    // Fuente única: backend `engineconfig.AppVersion`. Fallback vacío hasta fetch.
    const appVersion = ref<string>(wailsFlagsVersion() ?? '');

    const readFlags = () => {
        const v = wailsFlagsVersion();
        if (v) appVersion.value = v;
    };

    const fetchBackend = async () => {
        if (appVersion.value) return;
        try {
            const v = await GetAppVersion();
            if (typeof v === 'string' && v.length > 0) appVersion.value = v;
        } catch (_e) {}
    };

    readFlags();
    fetchBackend();
    window.addEventListener('wails:runtime-config-ready', () => {
        readFlags();
        fetchBackend();
    });

    return { appVersion };
}
