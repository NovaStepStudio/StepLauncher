import { Client } from "discord-rpc"

const clientId = "781091ae140df08966e7184f60cef3690b5afb95a213ca6378c9c480a7eccf72"
let client: Client | null = null
let activityInterval: ReturnType<typeof setInterval> | null = null
let isConnected = false

export async function startRPC(): Promise<void> {
  if (isConnected) return
  try {
    client = new Client({ transport: "ipc" })
    client.on("ready", () => {
      isConnected = true
      setActivity("En el launcher", "Navegando")
      activityInterval = setInterval(() => {
        if (isConnected) setActivity("En el launcher", "Navegando")
      }, 15_000)
    })
    client.on("disconnected", () => {
      isConnected = false
      stopRPC()
    })
    await client.login({ clientId })
  } catch {
    isConnected = false
    client = null
  }
}

export function stopRPC(): void {
  if (activityInterval) {
    clearInterval(activityInterval)
    activityInterval = null
  }
  if (client) {
    try { client.destroy() } catch {}
    client = null
  }
  isConnected = false
}

export function setActivity(details: string, state: string, startTimestamp?: number): void {
  if (!client || !isConnected) return
  try {
    client.setActivity({
      details,
      state,
      startTimestamp: startTimestamp ?? Date.now(),
      largeImageKey: "steplauncher",
      largeImageText: "StepLauncher",
      instance: false,
    })
  } catch {}
}
