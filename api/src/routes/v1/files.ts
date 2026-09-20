// Archivos v1: skins y capas de Minecraft (privados) con quota diaria.
// Límite: 10 skins+capas por día y usuario (ver MAX_UPLOADS_PER_DAY).
// Los buckets son privados: la descarga es por URL firmada de 1h (GET /:id/url).

import { Hono, type Context } from "hono";
import type { AppEnv } from "../../types/app";
import { getEnv, type SafeEnv } from "../../env";
import { ok, fail } from "../../lib/respond";
import { supabaseForUser } from "../../lib/supabase";
import { requireAuth, getBearerToken } from "../../middleware/auth";
import { rateLimit, rateLimitPresets } from "../../middleware/rate-limit";
import { uploadQuota, MAX_UPLOADS_PER_DAY } from "../../middleware/upload-quota";
import { validateQuery } from "../../middleware/validate";
import { listUploadsSchema } from "../../schemas/v1/files";

export const fileRoutes = new Hono<AppEnv>();

type UploadKind = "skin" | "cape";
const KIND_BUCKET: Record<UploadKind, string> = { skin: "skins", cape: "capes" };
/** Skins/capas son PNG de 64x64: 1 MB sobra y frena abuso de almacenamiento. */
const MAX_FILE_BYTES = 1024 * 1024;

interface UploadRow {
  id: string;
  kind: string;
  bucket: string;
  path: string;
  size_bytes: number;
  mime: string;
  created_at: string;
}

function uploadShape(row: UploadRow) {
  return {
    id: row.id,
    kind: row.kind,
    sizeBytes: row.size_bytes,
    mime: row.mime,
    createdAt: row.created_at,
  };
}

/** Verifica firma PNG (89 50 4E 47 0D 0A 1A 0A). No fiarse del MIME declarado. */
async function isPng(file: File): Promise<boolean> {
  const bytes = new Uint8Array(await file.slice(0, 8).arrayBuffer());
  return (
    bytes.length === 8 &&
    bytes[0] === 0x89 &&
    bytes[1] === 0x50 &&
    bytes[2] === 0x4e &&
    bytes[3] === 0x47 &&
    bytes[4] === 0x0d &&
    bytes[5] === 0x0a &&
    bytes[6] === 0x1a &&
    bytes[7] === 0x0a
  );
}

function authed(env: SafeEnv, token: string) {
  return supabaseForUser(env, token);
}

/** Subida genérica para skin/cape (multipart, campo `file`). */
async function handleUpload(c: Context<AppEnv>, kind: UploadKind) {
  const user = c.get("user");
  const token = getBearerToken(c);
  // Inalcanzable tras `requireAuth()`, pero se valida por seguridad.
  if (!user || !token) {
    return fail(c, { code: "unauthorized", message: "Autenticación requerida.", status: 401 });
  }
  let env;
  try {
    env = getEnv(c);
  } catch {
    return fail(c, { code: "server_misconfigured", message: "Servicio no disponible.", status: 500 });
  }

  const body = await c.req.parseBody();
  const raw = body["file"];
  const file = Array.isArray(raw) ? raw[0] : raw;
  if (!(file instanceof File) || file.size === 0) {
    return fail(c, { code: "validation_error", message: "Falta el archivo (campo `file`).", status: 400 });
  }
  if (file.size > MAX_FILE_BYTES) {
    return fail(c, { code: "file_too_large", message: "Máximo 1 MB por archivo.", status: 400 });
  }
  if (!(await isPng(file))) {
    return fail(c, { code: "invalid_file_type", message: "Solo PNG válido.", status: 400 });
  }

  const bucket = KIND_BUCKET[kind];
  const path = `${user.id}/${kind}-${Date.now()}.png`;
  const supabase = authed(env, token);

  const { error: upError } = await supabase.storage
    .from(bucket)
    .upload(path, file, { contentType: "image/png", upsert: false });
  if (upError) {
    console.error(`[${c.get("requestId")}] upload ${kind}: error storage:`, upError.message);
    return fail(c, { code: "upload_failed", message: "No se pudo guardar el archivo.", status: 500 });
  }

  const { data: row, error: dbError } = await supabase
    .from("file_uploads")
    .insert({
      user_id: user.id,
      kind,
      bucket,
      path,
      size_bytes: file.size,
      mime: "image/png",
    })
    .select("id, kind, bucket, path, size_bytes, mime, created_at")
    .single();
  if (dbError || !row) {
    // Limpieza best-effort del huérfano en storage (sin transacción entre ambos).
    await supabase.storage.from(bucket).remove([path]);
    console.error(`[${c.get("requestId")}] upload ${kind}: error ledger:`, dbError?.message);
    return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
  }

  return ok(c, uploadShape(row as UploadRow), 201);
}

// POST /v1/files/skin — subir skin (PNG ≤ 1 MB). Cuenta para la quota diaria.
fileRoutes.post(
  "/skin",
  rateLimit(rateLimitPresets.account),
  requireAuth(),
  uploadQuota(),
  (c) => handleUpload(c, "skin"),
);

// POST /v1/files/cape — subir capa (PNG ≤ 1 MB). Cuenta para la quota diaria.
fileRoutes.post(
  "/cape",
  rateLimit(rateLimitPresets.account),
  requireAuth(),
  uploadQuota(),
  (c) => handleUpload(c, "cape"),
);

// GET /v1/files — listar subidas propias (paginado).
fileRoutes.get("/", rateLimit(rateLimitPresets.account), requireAuth(), validateQuery(listUploadsSchema), async (c) => {
  const user = c.get("user");
  const token = getBearerToken(c);
  if (!user || !token) {
    return fail(c, { code: "unauthorized", message: "Autenticación requerida.", status: 401 });
  }
  let env;
  try {
    env = getEnv(c);
  } catch {
    return fail(c, { code: "server_misconfigured", message: "Servicio no disponible.", status: 500 });
  }
  const q = c.req.valid("query");
  const { data, error } = await authed(env, token)
    .from("file_uploads")
    .select("id, kind, bucket, path, size_bytes, mime, created_at")
    .eq("user_id", user.id)
    .order("created_at", { ascending: false })
    .range(q.offset, q.offset + q.limit - 1);
  if (error) {
    console.error(`[${c.get("requestId")}] files list:`, error.message);
    return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
  }
  return ok(c, {
    uploads: ((data ?? []) as UploadRow[]).map(uploadShape),
    limit: q.limit,
    offset: q.offset,
    dailyQuota: { limit: MAX_UPLOADS_PER_DAY, window: "24h", kinds: ["skin", "cape"] },
  });
});

// GET /v1/files/:id/url — URL firmada (1h) para descargar un archivo propio.
fileRoutes.get("/:id/url", rateLimit(rateLimitPresets.account), requireAuth(), async (c) => {
  const user = c.get("user");
  const token = getBearerToken(c);
  if (!user || !token) {
    return fail(c, { code: "unauthorized", message: "Autenticación requerida.", status: 401 });
  }
  let env;
  try {
    env = getEnv(c);
  } catch {
    return fail(c, { code: "server_misconfigured", message: "Servicio no disponible.", status: 500 });
  }
  const supabase = authed(env, token);
  const { data: row } = await supabase
    .from("file_uploads")
    .select("bucket, path")
    .eq("id", c.req.param("id"))
    .eq("user_id", user.id)
    .maybeSingle();
  if (!row) {
    return fail(c, { code: "not_found", message: "Recurso no encontrado.", status: 404 });
  }
  const { data: signed, error } = await supabase.storage
    .from((row as { bucket: string }).bucket)
    .createSignedUrl((row as { path: string }).path, 3600);
  if (error || !signed) {
    console.error(`[${c.get("requestId")}] files signed url:`, error?.message);
    return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
  }
  return ok(c, { url: signed.signedUrl, expiresIn: 3600 });
});

// DELETE /v1/files/:id — borrar un archivo propio (libera quota del día).
fileRoutes.delete("/:id", rateLimit(rateLimitPresets.account), requireAuth(), async (c) => {
  const user = c.get("user");
  const token = getBearerToken(c);
  if (!user || !token) {
    return fail(c, { code: "unauthorized", message: "Autenticación requerida.", status: 401 });
  }
  let env;
  try {
    env = getEnv(c);
  } catch {
    return fail(c, { code: "server_misconfigured", message: "Servicio no disponible.", status: 500 });
  }
  const supabase = authed(env, token);
  const { data: row } = await supabase
    .from("file_uploads")
    .select("bucket, path")
    .eq("id", c.req.param("id"))
    .eq("user_id", user.id)
    .maybeSingle();
  if (!row) {
    return fail(c, { code: "not_found", message: "Recurso no encontrado.", status: 404 });
  }
  const r = row as { bucket: string; path: string };
  const { error: rmError } = await supabase.storage.from(r.bucket).remove([r.path]);
  if (rmError) {
    console.error(`[${c.get("requestId")}] files delete storage:`, rmError.message);
    return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
  }
  await supabase.from("file_uploads").delete().eq("id", c.req.param("id")).eq("user_id", user.id);
  return ok(c, { deleted: true });
});
