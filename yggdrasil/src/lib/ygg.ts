// Formas exactas del protocolo Yggdrasil (Mojang + authlib-injector).
// Nada de sobre `ok()`/`fail()`: aquí mandan los campos del spec.
// Ver `docs/base-yggdrasil.md` para el origen de cada campo.

import type { Context } from "hono";
import type { YggEnv } from "../types/app";
import { stripUuid } from "./crypto";

/** Error Yggdrasil: `{ error, errorMessage, cause? }` con su HTTP. */
export function yggError(
  c: Context<YggEnv>,
  status: 400 | 401 | 403 | 404 | 429 | 500,
  error: string,
  errorMessage: string,
  cause?: string,
) {
  return c.json(
    {
      error,
      errorMessage,
      ...(cause ? { cause } : {}),
    },
    status,
  );
}

/** 403 genérico de credenciales (no revela si el usuario existe). */
export function invalidCredentials(c: Context<YggEnv>) {
  return yggError(
    c,
    403,
    "ForbiddenOperationException",
    "Invalid credentials. Invalid username or password.",
  );
}

/** 403 genérico de token inválido (no distingue expirado de revocado). */
export function invalidToken(c: Context<YggEnv>) {
  return yggError(c, 403, "ForbiddenOperationException", "Invalid token.");
}

/** Usuario serializado (solo cuando `requestUser=true`). */
export function serializeUser(id: string) {
  return { id: stripUuid(id), properties: [] as Array<{ name: string; value: string }> };
}

/** Perfil sin propiedades (listas y respuestas de login). */
export function serializeProfile(uuid: string, name: string) {
  return { id: stripUuid(uuid), name };
}

/**
 * Perfil completo para `hasJoined`/`profile`: incluye `textures` + firma.
 * `texturesValue` es el base64 del JSON de texturas; `signature` su firma RSA.
 * `properties` SIEMPRE va como arreglo (vacío si no hay texturas): así lo
 * emiten Mojang y el propio authlib-injector (`YggdrasilResponseBuilder`), y
 * así lo exige su parser (`parseGameProfile` revienta si falta la clave).
 * Sin texturas, el juego/launcher usan la skin por defecto.
 */
export function serializeProfileFull(
  uuid: string,
  name: string,
  texturesValue?: string,
  signature?: string,
) {
  const properties: Array<{ name: string; value: string; signature?: string }> = [];
  if (texturesValue) {
    properties.push({
      name: "textures",
      value: texturesValue,
      ...(signature ? { signature } : {}),
    });
  }
  return { id: stripUuid(uuid), name, properties };
}
