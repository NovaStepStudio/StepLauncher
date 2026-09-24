// Validación de entrada con Zod para el protocolo Yggdrasil.
// El cuerpo/query que no cumple el contrato se rechaza con forma Yggdrasil:
// `{ error: "IllegalArgumentException", errorMessage }` (400).

import type { Context } from "hono";
import { validator } from "hono/validator";
import type { z } from "zod";
import type { YggEnv } from "../types/app";

/** Valida `application/json` contra un esquema Zod. */
export const validateJson = <S extends z.ZodTypeAny>(schema: S) => {
  return validator("json", (value, c: Context<YggEnv>) => {
    const parsed = schema.safeParse(value);
    if (!parsed.success) {
      const first = parsed.error.issues[0];
      const detail = first ? `${first.path.join(".")}: ${first.message}` : "Cuerpo inválido.";
      return c.json(
        { error: "IllegalArgumentException", errorMessage: `Datos inválidos. ${detail}` },
        400,
      );
    }
    return parsed.data;
  });
};

/** Valida query params contra un esquema Zod. */
export const validateQuery = <S extends z.ZodTypeAny>(schema: S) => {
  return validator("query", (value, c: Context<YggEnv>) => {
    const parsed = schema.safeParse(value);
    if (!parsed.success) {
      return c.json(
        { error: "IllegalArgumentException", errorMessage: "Parámetros inválidos." },
        400,
      );
    }
    return parsed.data;
  });
};

/** Valida path params (ej: `:uuid`) contra un esquema Zod. */
export const validateParam = <S extends z.ZodTypeAny>(schema: S) => {
  return validator("param", (value, c: Context<YggEnv>) => {
    const parsed = schema.safeParse(value);
    if (!parsed.success) {
      return c.json(
        { error: "IllegalArgumentException", errorMessage: "Parámetro inválido." },
        400,
      );
    }
    return parsed.data;
  });
};
