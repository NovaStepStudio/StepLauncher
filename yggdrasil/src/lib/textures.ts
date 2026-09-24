// Texturas del jugador para el protocolo Yggdrasil (solo lectura).
// Fuente: buckets `skins`/`capes` + libro `file_uploads` de `api/` (MISMO
// proyecto Supabase, solo lectura). Aquí NO se sube nada: la subida vive en
// `api/` + la web (`POST /v1/files/skin|cape`).
// Se sirven por `GET /textures/:hash` (hash = SHA-256 del PNG, como pide el
// spec para la caché del cliente) y se firman con `YGG_SIGN_PRIVATE_KEY`.

import type { YggEnvSafe } from "../env";
import { adminClient } from "./store";
import { signRsaSha1, stripUuid, toBase64Json } from "./crypto";

export interface SignedTextures {
  /** Base64 del JSON `{ timestamp, profileId, profileName, textures }`. */
  value: string;
  /** Firma `SHA1withRSA` del valor (ausente solo si falló la firma). */
  signature?: string;
}

interface StoredFile {
  bucket: string;
  path: string;
}

/** Última skin y capa del usuario según el libro `file_uploads` de `api/`. */
export async function latestUploads(
  env: YggEnvSafe,
  userId: string,
): Promise<{ skin?: StoredFile; cape?: StoredFile }> {
  const admin = adminClient(env);
  const { data, error } = await admin
    .from("file_uploads")
    .select("kind, bucket, path")
    .eq("user_id", userId)
    .in("kind", ["skin", "cape"])
    .order("created_at", { ascending: false })
    .limit(10);
  if (error || !data || !Array.isArray(data)) return {};
  const out: { skin?: StoredFile; cape?: StoredFile } = {};
  for (const row of data as Array<{ kind?: unknown; bucket?: unknown; path?: unknown }>) {
    if (typeof row.bucket !== "string" || typeof row.path !== "string") continue;
    if (row.kind === "skin" && !out.skin) out.skin = { bucket: row.bucket, path: row.path };
    if (row.kind === "cape" && !out.cape) out.cape = { bucket: row.bucket, path: row.path };
    if (out.skin && out.cape) break;
  }
  return out;
}

/** Descarga un PNG del storage con `service_role` y verifica su firma mágica. */
export async function downloadPng(
  env: YggEnvSafe,
  bucket: string,
  path: string,
): Promise<Uint8Array | null> {
  // Defensa en profundidad: `api/` ya verificó el PNG al subir, pero el
  // contenido pudo cambiar por fuera. Límite 2 MB contra PNG-bomb.
  const admin = adminClient(env);
  const { data, error } = await admin.storage.from(bucket).download(path);
  if (error || !data) return null;
  const bytes = new Uint8Array(await data.arrayBuffer());
  if (bytes.length === 0 || bytes.length > 2 * 1024 * 1024) return null;
  const png = [0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a];
  for (let i = 0; i < png.length; i++) {
    if (bytes[i] !== png[i]) return null;
  }
  return bytes;
}

/** SHA-256 en hex de unos bytes (identificador de textura para el cliente). */
export async function sha256Hex(bytes: Uint8Array): Promise<string> {
  const digest = await crypto.subtle.digest("SHA-256", bytes as BufferSource);
  return [...new Uint8Array(digest)].map((b) => b.toString(16).padStart(2, "0")).join("");
}

/** Recuerda hash → ubicación para servir `GET /textures/:hash` sin adivinar. */
async function rememberTexture(
  env: YggEnvSafe,
  hash: string,
  file: StoredFile,
  kind: "skin" | "cape",
  userId: string,
): Promise<void> {
  const admin = adminClient(env);
  const { error } = await admin.from("ygg_textures").upsert(
    { hash, bucket: file.bucket, path: file.path, kind, user_id: userId },
    { onConflict: "hash", ignoreDuplicates: true },
  );
  // No rompe la respuesta, pero se registra alto y claro: sin esta tabla las
  // URLs de texturas devuelven 404 y el juego queda en Steve (ver `wrangler tail`).
  if (error) {
    console.error(`rememberTexture: no se pudo guardar ${hash} (code ${error.code ?? "?"}):`, error.message);
  }
}

/**
 * Ubicación de una textura por su hash (para `GET /textures/:hash`).
 * Lanza si la tabla no existe (falta aplicar el SQL): es error de
 * configuración, no un 404, y debe verse en los logs.
 */
export async function lookupTexture(env: YggEnvSafe, hash: string): Promise<StoredFile | null> {
  const admin = adminClient(env);
  const { data, error } = await admin
    .from("ygg_textures")
    .select("bucket, path")
    .eq("hash", hash.toLowerCase())
    .maybeSingle();
  if (error) {
    if ((error as { code?: string }).code === "42P01") {
      throw new Error("tabla ygg_textures ausente: falta aplicar db/01_ygg.sql en Supabase");
    }
    console.error("lookupTexture:", error.message);
    return null;
  }
  if (!data) return null;
  const row = data as { bucket?: unknown; path?: unknown };
  if (typeof row.bucket !== "string" || typeof row.path !== "string") return null;
  return { bucket: row.bucket, path: row.path };
}

/**
 * Construye y firma la propiedad `textures` de un perfil.
 * Devuelve `undefined` si el jugador no tiene skin ni capa (el juego y el
 * launcher usan entonces la skin por defecto, sin romper nada).
 * Si la firma falla, devuelve el valor sin firma y lo registra: el juego lo
 * acepta salvo que el servidor exija perfil seguro.
 */
export async function signedTexturesFor(
  env: YggEnvSafe,
  opts: { userId: string; profileUuid: string; profileName: string; origin: string },
): Promise<SignedTextures | undefined> {
  const uploads = await latestUploads(env, opts.userId);
  if (!uploads.skin && !uploads.cape) return undefined;

  const textures: Record<string, { url: string }> = {};
  const jobs: Array<{ kind: "skin" | "cape"; file: StoredFile }> = [];
  if (uploads.skin) jobs.push({ kind: "skin", file: uploads.skin });
  if (uploads.cape) jobs.push({ kind: "cape", file: uploads.cape });

  for (const job of jobs) {
    const bytes = await downloadPng(env, job.file.bucket, job.file.path);
    if (!bytes) continue;
    const hash = await sha256Hex(bytes);
    await rememberTexture(env, hash, job.file, job.kind, opts.userId);
    // Sin `metadata.model`: `api/` no guarda slim/classic (clásico por defecto).
    textures[job.kind.toUpperCase()] = { url: `${opts.origin}/textures/${hash}` };
  }
  if (Object.keys(textures).length === 0) return undefined;

  const value = toBase64Json({
    timestamp: Date.now(),
    profileId: stripUuid(opts.profileUuid),
    profileName: opts.profileName,
    textures,
  });
  try {
    const signature = await signRsaSha1(env.signPrivateKeyPem, value);
    return { value, signature };
  } catch (err) {
    console.error("texturas sin firma (clave RSA inválida):", (err as Error)?.message ?? err);
    return { value };
  }
}
