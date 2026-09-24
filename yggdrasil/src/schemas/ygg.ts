// Contratos de entrada Yggdrasil (Zod estricto, propios de este servicio).
// Nada se reutiliza desde `api/`: el protocolo manda sus nombres de campos.

import { z } from "zod";

/**
 * El launcher StepLauncher guarda y reenvía tokens y UUID con guiones
 * (`normalizeAccessToken`): se aceptan con o sin guiones y se comparan
 * sin ellos (ley de Postel, sin romper el spec que usa sin guiones).
 */
function sinGuiones(v: unknown): unknown {
  return typeof v === "string" ? v.replace(/-/g, "") : v;
}

const uuidBare = z.preprocess(
  sinGuiones,
  z
    .string()
    .trim()
    .regex(/^[0-9a-f]{32}$/i, "UUID inválido.")
    .transform((s) => s.toLowerCase()),
);

/** `accessToken` opaco: con o sin guiones, 8–128 caracteres tras normalizar. */
const tokenYgg = z.preprocess(sinGuiones, z.string().trim().min(8).max(128));

const usernameLoose = z.string().trim().min(1, "Requerido.").max(254);

const passwordLoose = z.string().min(1, "Requerida.").max(512);

const clientTokenLoose = z.preprocess(sinGuiones, z.string().trim().min(1).max(128)).optional();

/** POST /authserver/authenticate */
export const authenticateSchema = z
  .object({
    username: usernameLoose,
    password: passwordLoose,
    clientToken: clientTokenLoose,
    requestUser: z.boolean().optional().default(false),
    agent: z
      .object({ name: z.string().trim().min(1).max(32), version: z.number().int().min(1).max(99) })
      .passthrough()
      .optional(),
  })
  .strict();

/** POST /authserver/refresh */
export const refreshSchema = z
  .object({
    accessToken: tokenYgg,
    clientToken: clientTokenLoose,
    requestUser: z.boolean().optional().default(false),
    selectedProfile: z
      .object({ id: uuidBare, name: z.string().trim().min(3).max(16) })
      .strict()
      .optional(),
  })
  .strict();

/** POST /authserver/validate y /authserver/invalidate */
export const tokenPairSchema = z
  .object({
    accessToken: tokenYgg,
    clientToken: clientTokenLoose,
  })
  .strict();

/** POST /authserver/signout */
export const signoutSchema = z
  .object({ username: usernameLoose, password: passwordLoose })
  .strict();

/** POST /sessionserver/session/minecraft/join */
export const joinSchema = z
  .object({
    accessToken: tokenYgg,
    selectedProfile: uuidBare,
    serverId: z.string().trim().min(1).max(128),
  })
  .strict();

/** GET /sessionserver/session/minecraft/hasJoined */
export const hasJoinedSchema = z
  .object({
    username: z.string().trim().min(3).max(16),
    serverId: z.string().trim().min(1).max(128),
    ip: z.string().trim().max(64).optional(),
  })
  .strict();

/** GET /sessionserver/session/minecraft/profile/:uuid */
export const profileParamSchema = z.object({ uuid: uuidBare }).strict();

export const profileQuerySchema = z
  .object({ unsigned: z.enum(["true", "false"]).optional().default("true") })
  .strict();

/** POST /api/profiles/minecraft (lote de nombres, 1–10). */
export const profilesBatchSchema = z.array(z.string().trim().min(3).max(16)).min(1).max(10);

/** GET /textures/:hash — SHA-256 del PNG en hex. */
export const textureHashSchema = z
  .object({ hash: z.string().trim().toLowerCase().regex(/^[0-9a-f]{64}$/, "Hash inválido.") })
  .strict();

/** GET /skins/MinecraftSkins/:file — `<usuario>.png` de la API legacy. */
export const legacySkinParamSchema = z
  .object({ file: z.string().trim().min(5).max(21) })
  .strict();

/** GET /skins/:name y /cloaks/:name — nombre con o sin `.png` (estilo Ely.by). */
export const skinNameSchema = z
  .object({ name: z.string().trim().min(3).max(21) })
  .strict();

export const cloakNameSchema = z
  .object({ name: z.string().trim().min(3).max(21) })
  .strict();

export type AuthenticateInput = z.infer<typeof authenticateSchema>;
export type RefreshInput = z.infer<typeof refreshSchema>;
