// Utilidades criptográficas y de formato del protocolo Yggdrasil.
// Sin secretos hardcodeados: la clave privada llega por entorno (ver env.ts).

/** Quita los guiones de un UUID (`853c80ef-…` → `853c80ef…`). */
export function stripUuid(uuid: string): string {
  return uuid.replace(/-/g, "").toLowerCase();
}

/** Pone guiones a un UUID sin guiones. Si no tiene 32 hex, lo devuelve igual. */
export function hyphenateUuid(raw: string): string {
  const s = raw.replace(/-/g, "").toLowerCase();
  if (!/^[0-9a-f]{32}$/.test(s)) return raw;
  return `${s.slice(0, 8)}-${s.slice(8, 12)}-${s.slice(12, 16)}-${s.slice(16, 20)}-${s.slice(20)}`;
}

/** ¿Es un UUID (con o sin guiones)? */
export function isUuid(value: string): boolean {
  return /^[0-9a-f]{8}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{12}$/i.test(value.trim());
}

/** Token opaco de 32 hex (128 bits aleatorios) para `accessToken`. */
export function newAccessToken(): string {
  const b = new Uint8Array(16);
  crypto.getRandomValues(b);
  return [...b].map((x) => x.toString(16).padStart(2, "0")).join("");
}

/** `clientToken` por defecto: UUID v4 sin guiones. */
export function newClientToken(): string {
  const b = new Uint8Array(16);
  crypto.getRandomValues(b);
  const sixth = b[6] ?? 0;
  const eighth = b[8] ?? 0;
  b[6] = (sixth & 0x0f) | 0x40;
  b[8] = (eighth & 0x3f) | 0x80;
  const hex = [...b].map((x) => x.toString(16).padStart(2, "0")).join("");
  return hex;
}

/** Expiración de tokens Yggdrasil: 15 días desde ahora (formato ISO). */
export function tokenExpiryIso(): string {
  return new Date(Date.now() + 15 * 24 * 60 * 60 * 1000).toISOString();
}

/** Expiración de un `join`: 30 segundos desde ahora (formato ISO). */
export function joinExpiryIso(): string {
  return new Date(Date.now() + 30 * 1000).toISOString();
}

/**
 * Firma un valor de atributo (`textures` en base64) con SHA1withRSA, el
 * algoritmo que el juego verifica contra `signaturePublickey` (sin esta
 * firma el cliente moderno no muestra la skin).
 * Intenta WebCrypto estándar primero y `node:crypto` como respaldo; si todo
 * falla lanza con el motivo para que el llamador lo registre (`wrangler tail`).
 */
export async function signRsaSha1(pemPrivateKey: string, data: string): Promise<string> {
  // Normaliza `\n` literales que a veces llegan desde `.dev.vars`/secrets.
  const pem = pemPrivateKey.replace(/\\n/g, "\n");
  const payload = new TextEncoder().encode(data);

  try {
    return await signWebCrypto(pem, payload);
  } catch (webErr) {
    try {
      return await signNodeCrypto(pem, data);
    } catch (nodeErr) {
      throw new Error(
        `firma RSA-SHA1 imposible (webcrypto: ${mensaje(webErr)}; node:crypto: ${mensaje(nodeErr)})`,
      );
    }
  }
}

function mensaje(err: unknown): string {
  return err instanceof Error ? err.message : String(err);
}

/** Firma vía WebCrypto (disponible en todo runtime Workers, sin flags). */
async function signWebCrypto(pem: string, payload: Uint8Array): Promise<string> {
  const b64 = pem
    .replace(/-----(BEGIN|END) PRIVATE KEY-----/g, "")
    .replace(/\s+/g, "");
  const bin = atob(b64);
  const der = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; i++) der[i] = bin.charCodeAt(i);
  const key = await crypto.subtle.importKey(
    "pkcs8",
    der.buffer as ArrayBuffer,
    { name: "RSASSA-PKCS1-v1_5", hash: "SHA-1" },
    false,
    ["sign"],
  );
  const firma = await crypto.subtle.sign({ name: "RSASSA-PKCS1-v1_5" }, key, payload as BufferSource);
  return base64De(new Uint8Array(firma));
}

/** Firma vía `node:crypto` (requiere `nodejs_compat` en el Worker). */
async function signNodeCrypto(pem: string, data: string): Promise<string> {
  // Importación diferida para no romper el arranque si el runtime cambia.
  const { createSign } = await import("node:crypto");
  const firmador = createSign("RSA-SHA1");
  firmador.update(data, "utf8");
  firmador.end();
  return firmador.sign(pem, "base64");
}

/**
 * Deriva la pública PEM (SPKI) desde la privada PKCS8.
 * Vacío si la clave es inválida (el llamador decide cómo degradar).
 */
export async function derivePublicPem(privatePem: string): Promise<string> {
  try {
    const { createPublicKey } = await import("node:crypto");
    const pub = createPublicKey(privatePem.replace(/\\n/g, "\n"));
    const out = pub.export({ type: "spki", format: "pem" }) as string;
    return out.trim() + "\n";
  } catch {
    return "";
  }
}

/** Base64 estándar de unos bytes (en trozos: los PNG llegan al MB). */
function base64De(bytes: Uint8Array): string {
  let bin = "";
  const paso = 0x8000;
  for (let i = 0; i < bytes.length; i += paso) {
    bin += String.fromCharCode(...bytes.subarray(i, i + paso));
  }
  return btoa(bin);
}

/** Codifica un objeto a base64 estándar (para el valor de `textures`). */
export function toBase64Json(value: unknown): string {
  const json = JSON.stringify(value);
  const bytes = new TextEncoder().encode(json);
  let bin = "";
  for (const b of bytes) bin += String.fromCharCode(b);
  return btoa(bin);
}
