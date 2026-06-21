import type { ApiError } from "../Types/Auth.js"

const API_BASE_URL = "https://api.stepnicka012.workers.dev"

class ApiClient {
  private baseUrl: string
  private token: string | null = null

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl.replace(/\/+$/, "")
  }

  GetBaseUrl(): string {
    return this.baseUrl
  }

  SetToken(token: string | null) {
    this.token = token
  }

  GetToken(): string | null {
    return this.token
  }

  private BuildHeaders(): Record<string, string> {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    }
    if (this.token) {
      headers["Authorization"] = `Bearer ${this.token}`
    }
    return headers
  }

  private async Request<T>(
    method: string,
    path: string,
    body?: unknown
  ): Promise<T> {
    const url = `${this.baseUrl}${path}`
    const options: RequestInit = {
      method,
      headers: this.BuildHeaders(),
    }

    if (body !== undefined) {
      options.body = JSON.stringify(body)
    }

    const response = await fetch(url, options)
    const data = await response.json()

    if (!response.ok) {
      const err = data as ApiError
      throw new Error(err.detail ?? err.error ?? `HTTP ${response.status}`)
    }

    return data as T
  }

  async Get<T>(path: string): Promise<T> {
    return this.Request<T>("GET", path)
  }

  async Post<T>(path: string, body?: unknown): Promise<T> {
    return this.Request<T>("POST", path, body)
  }

  async Patch<T>(path: string, body?: unknown): Promise<T> {
    return this.Request<T>("PATCH", path, body)
  }

  async Put<T>(path: string, body?: unknown): Promise<T> {
    return this.Request<T>("PUT", path, body)
  }

  async Delete<T>(path: string, body?: unknown): Promise<T> {
    return this.Request<T>("DELETE", path, body)
  }
}

export const api = new ApiClient()
