import { ipcMain } from "electron"
import { readFileSync, writeFileSync, mkdirSync, existsSync, readdirSync, rmSync } from "fs"
import { join } from "path"
import { homedir, platform } from "os"

function getInstancesDir(): string {
    const base = platform() === "win32"
        ? join(process.env.APPDATA || homedir(), ".StepLauncher")
        : join(homedir(), ".StepLauncher")
    const dir = join(base, "instances")
    mkdirSync(dir, { recursive: true })
    mkdirSync(join(dir, "shared"), { recursive: true })
    return dir
}

function getSharedDir(): string {
    return join(getInstancesDir(), "shared")
}

export function RegisterInstancesIpc() {
    ipcMain.handle("instances:GetDir", () => getInstancesDir())
    ipcMain.handle("instances:GetSharedDir", () => getSharedDir())

    ipcMain.handle("instances:List", async () => {
        const dir = getInstancesDir()
        if (!existsSync(dir)) return []
        const entries = readdirSync(dir, { withFileTypes: true })
        const instances: any[] = []
        for (const entry of entries) {
            if (!entry.isDirectory() || entry.name === "shared") continue
            const instDir = join(dir, entry.name)
            const configPath = join(instDir, "instance.config.json")
            const metaPath = join(instDir, "instance.metadata.json")
            let config = null, metadata = null
            if (existsSync(configPath)) {
                try { config = JSON.parse(readFileSync(configPath, "utf-8")) } catch {}
            }
            if (existsSync(metaPath)) {
                try { metadata = JSON.parse(readFileSync(metaPath, "utf-8")) } catch {}
            }
            instances.push({
                id: entry.name,
                dir: instDir,
                config,
                metadata,
            })
        }
        return instances
    })

    ipcMain.handle("instances:Get", async (_e, instanceId: string) => {
        const dir = join(getInstancesDir(), instanceId)
        if (!existsSync(dir)) return null
        const configPath = join(dir, "instance.config.json")
        const metaPath = join(dir, "instance.metadata.json")
        let config = null, metadata = null
        if (existsSync(configPath)) {
            try { config = JSON.parse(readFileSync(configPath, "utf-8")) } catch {}
        }
        if (existsSync(metaPath)) {
            try { metadata = JSON.parse(readFileSync(metaPath, "utf-8")) } catch {}
        }
        return { id: instanceId, dir, config, metadata }
    })

    ipcMain.handle("instances:UpdateConfig", async (_e, instanceId: string, updates: Record<string, any>) => {
        const dir = join(getInstancesDir(), instanceId)
        mkdirSync(dir, { recursive: true })
        const configPath = join(dir, "instance.config.json")
        let config: any = {}
        if (existsSync(configPath)) {
            try { config = JSON.parse(readFileSync(configPath, "utf-8")) } catch {}
        }
        const merged = JSON.parse(JSON.stringify(config))
        if (updates.name !== undefined) {
            if (!merged.instanceMetadata) merged.instanceMetadata = {}
            if (!merged.instanceMetadata.frontend) merged.instanceMetadata.frontend = {}
            merged.instanceMetadata.frontend.name = updates.name
        }
        if (updates.description !== undefined) {
            if (!merged.instanceMetadata) merged.instanceMetadata = {}
            if (!merged.instanceMetadata.frontend) merged.instanceMetadata.frontend = {}
            merged.instanceMetadata.frontend.description = updates.description
        }
        if (updates.icon !== undefined) {
            if (!merged.instanceMetadata) merged.instanceMetadata = {}
            if (!merged.instanceMetadata.frontend) merged.instanceMetadata.frontend = {}
            merged.instanceMetadata.frontend.icon = updates.icon
        }
        if (updates.hero !== undefined) {
            if (!merged.instanceMetadata) merged.instanceMetadata = {}
            if (!merged.instanceMetadata.frontend) merged.instanceMetadata.frontend = {}
            merged.instanceMetadata.frontend.hero = updates.hero
        }
        if (updates.configInstance !== undefined) {
            merged.configInstance = { ...(merged.configInstance || {}), ...updates.configInstance }
        }
        if (updates.instanceMetadata !== undefined) {
            if (!merged.instanceMetadata) merged.instanceMetadata = {}
            merged.instanceMetadata = { ...merged.instanceMetadata, ...updates.instanceMetadata }
        }
        merged.instanceMetadata.updatedAt = new Date().toISOString()
        writeFileSync(configPath, JSON.stringify(merged, null, 2), "utf-8")
        return { success: true }
    })

    ipcMain.handle("instances:Delete", async (_e, instanceId: string) => {
        const dir = join(getInstancesDir(), instanceId)
        if (!existsSync(dir)) return { success: false, error: "Not found" }
        rmSync(dir, { recursive: true, force: true })
        return { success: true }
    })
}
