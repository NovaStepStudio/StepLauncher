// Validación de entrada con Zod: el cuerpo/query que no cumple el contrato se rechaza.
// Evita inyecciones, tipos rotos y que un endpoint contamine a otro.

import type { Context } from "hono";
import { validator } from "hono/validator";
import type { z } from "zod";
import type { AppEnv } from "../types/app";

/** Valida `application/json` contra un esquema Zod. Devuelve 400 con detalles por campo. */
// Sin anotar el retorno a propósito: se infiere el tipo del validador de Hono
// para que `c.req.valid("json")` llegue tipado al handler.
export const validateJson = <S extends z.ZodTypeAny>(schema: S) => {
  return validator("json", (value, c: Context<AppEnv>) => {
    const parsed = schema.safeParse(value);
    if (!parsed.success) {
      return c.json(
        {
          success: false,
          error: {
            code: "validation_error",
            message: "Datos inválidos.",
            details: parsed.error.flatten().fieldErrors,
            requestId: c.get("requestId") as string | undefined,
          },
        },
        400,
      );
    }
    return parsed.data;
  });
};

/** Valida query params contra un esquema Zod (paginación, filtros). */
// Igual que arriba: retorno inferido para tipar `c.req.valid("query")`.
export const validateQuery = <S extends z.ZodTypeAny>(schema: S) => {
  return validator("query", (value, c: Context<AppEnv>) => {
    const parsed = schema.safeParse(value);
    if (!parsed.success) {
      return c.json(
        {
          success: false,
          error: {
            code: "validation_error",
            message: "Parámetros inválidos.",
            details: parsed.error.flatten().fieldErrors,
            requestId: c.get("requestId") as string | undefined,
          },
        },
        400,
      );
    }
    return parsed.data;
  });
};

/** Valida path params (ej: `:id` UUID) contra un esquema Zod. */
// Igual que arriba: retorno inferido para tipar `c.req.valid("param")`.
export const validateParam = <S extends z.ZodTypeAny>(schema: S) => {
  return validator("param", (value, c: Context<AppEnv>) => {
    const parsed = schema.safeParse(value);
    if (!parsed.success) {
      return c.json(
        {
          success: false,
          error: {
            code: "validation_error",
            message: "Parámetro inválido.",
            details: parsed.error.flatten().fieldErrors,
            requestId: c.get("requestId") as string | undefined,
          },
        },
        400,
      );
    }
    return parsed.data;
  });
};
