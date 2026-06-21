import { ipcMain } from "electron"
import { getNovaCoreManager } from "../Services/NovaCoreManager.js"

export function RegisterNovaCoreIpc() {
    const mgr = getNovaCoreManager()

    ipcMain.handle("novacore:Start", async () => {
        try {
            await mgr.start()
            return { success: true }
        } catch (e: any) {
            return { success: false, error: e?.message ?? String(e) }
        }
    })

    ipcMain.handle("novacore:Stop", async () => {
        try {
            await mgr.stop()
            return { success: true }
        } catch (e: any) {
            return { success: false, error: e?.message ?? String(e) }
        }
    })

    ipcMain.handle("novacore:Status", () => {
        return mgr.getStatus()
    })

    ipcMain.handle("novacore:Health", async () => {
        return mgr.health()
    })

    ipcMain.handle("novacore:Versions", async (_event, limit?: number) => {
        return mgr.getVersions(limit)
    })

    ipcMain.handle("novacore:Install", async (_event, req: any) => {
        try {
            const sessionId = await mgr.install(req)
            return { success: true, sessionId }
        } catch (e: any) {
            return { success: false, error: e?.message ?? String(e) }
        }
    })

    ipcMain.handle("novacore:Launch", async (_event, req: any) => {
        try {
            const launchId = await mgr.launch(req)
            return { success: true, launchId }
        } catch (e: any) {
            return { success: false, error: e?.message ?? String(e) }
        }
    })

    ipcMain.handle("novacore:KillInstance", async (_event, id: string) => {
        try {
            await mgr.killInstance(id)
            return { success: true }
        } catch (e: any) {
            return { success: false, error: e?.message ?? String(e) }
        }
    })

    ipcMain.handle("novacore:EngineInfo", async () => {
        return mgr.getEngineInfo()
    })

    ipcMain.handle("novacore:ModLoaders", async () => {
        return mgr.getModLoaders()
    })

    ipcMain.handle("novacore:ModLoaderVersions", async (_event, loader: string, mcVersion: string) => {
        return mgr.getModLoaderVersions(loader, mcVersion)
    })

    ipcMain.handle("novacore:Progress", async (_event, sessionId: string) => {
        return mgr.getProgress(sessionId)
    })

    ipcMain.handle("novacore:GlobalProgress", async () => {
        return mgr.getGlobalProgress()
    })

    ipcMain.handle("novacore:PauseInstall", async (_event, sessionId: string) => {
        await mgr.pauseInstall(sessionId)
    })

    ipcMain.handle("novacore:ResumeInstall", async (_event, sessionId: string) => {
        await mgr.resumeInstall(sessionId)
    })

    ipcMain.handle("novacore:CancelInstall", async (_event, sessionId: string) => {
        await mgr.cancelInstall(sessionId)
    })

    ipcMain.handle("novacore:RunningInstances", async () => {
        return mgr.getRunningInstances()
    })

    ipcMain.handle("novacore:LatestCrash", async () => {
        return mgr.getLatestCrash()
    })

    ipcMain.handle("novacore:Sessions", async () => {
        return mgr.getSessions()
    })

    ipcMain.handle("novacore:RecoverySessions", async () => {
        return mgr.getRecoverySessions()
    })

    ipcMain.handle("novacore:InstallModLoader", async (_event, req: any) => {
        try {
            await mgr.installModLoader(req)
            return { success: true }
        } catch (e: any) {
            return { success: false, error: e?.message ?? String(e) }
        }
    })

    ipcMain.handle("novacore:DownloadRuntime", async (_event, version: string, instancePath: string, sharedPath?: string) => {
        await mgr.downloadRuntime(version, instancePath, sharedPath)
    })

    ipcMain.handle("novacore:Worlds", async () => {
        return mgr.getWorlds()
    })

    ipcMain.handle("novacore:GetInfo", async () => {
        return {
            version: mgr.getVersion(),
            jarPath: mgr.getJarPath(),
            authlibPath: mgr.getAuthlibPath(),
            instancesDir: mgr.getInstancesDir(),
            baseDir: mgr.getBaseDir(),
            versionsDir: mgr.getVersionsDir(),
            javaPath: mgr.getJavaPath(),
        }
    })

    ipcMain.handle("novacore:GetDownloadedVersions", async () => {
        return mgr.getDownloadedVersions()
    })

    ipcMain.handle("novacore:GetLogFiles", async () => {
        return mgr.getLogFiles()
    })

    ipcMain.handle("novacore:ReadLogFile", async (_event, fileName: string, maxLines?: number) => {
        return mgr.readLogFile(fileName, maxLines)
    })
}
