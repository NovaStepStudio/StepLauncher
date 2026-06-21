export interface ThemeMetadata {
  name: string
  author: string
  homePage: string
  version: string
  thumbnail: string
  gallery: string[]
  config: Record<string, unknown>
}

export interface ThemeInfo {
  name: string
  author: string
  homePage: string
  version: string
  thumbnail: string
  path: string
}
