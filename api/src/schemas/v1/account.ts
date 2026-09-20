// Contratos de entrada v1 para cuentas. Versionados: jamás reutilizar en v2.
// v2 copiará y evolucionará este archivo en `src/schemas/v2/`.

import { z } from "zod";

/** Nombre visible del jugador: letras, números, espacios y guiones. Sin HTML ni rutas. */
export const displayNameSchema = z
  .string()
  .trim()
  .min(3, "Mínimo 3 caracteres.")
  .max(32, "Máximo 32 caracteres.")
  .regex(/^[\p{L}\p{N} _-]+$/u, "Solo letras, números, espacios y guiones.");

/** Nombre de usuario único: 3–20 caracteres, letras, números y . _ - */
export const usernameSchema = z
  .string()
  .trim()
  .min(3, "Mínimo 3 caracteres.")
  .max(20, "Máximo 20 caracteres.")
  .regex(/^[A-Za-z0-9._-]+$/, "Solo letras, números y . _ -");

/** Descripción del jugador: texto libre hasta 5000 caracteres (vacío = sin bio). */
export const bioSchema = z.string().max(5000, "Máximo 5000 caracteres.");

/** Última versión de Minecraft Java jugada (ej: 1.21, 1.20.4, 24w14a). */
export const mcVersionSchema = z
  .string()
  .trim()
  .min(1, "Versión vacía.")
  .max(32, "Máximo 32 caracteres.")
  .regex(/^[A-Za-z0-9._+-]+$/, "Formato de versión de Minecraft inválido (ej: 1.21).");

/** UUID de Minecraft Java enlazado (premium). Vacío = desvincular. */
export const mcUuidSchema = z
  .string()
  .trim()
  .max(36)
  .refine((v) => v === "" || /^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$/.test(v), {
    message: "UUID de Minecraft inválido.",
  });

/** Actualización parcial del perfil propio. Todo opcional, pero al menos un campo. */
export const updateProfileSchema = z
  .object({
    displayName: displayNameSchema.optional(),
    username: usernameSchema.optional(),
    bio: bioSchema.optional(),
    lastMcVersion: mcVersionSchema.optional(),
    mcUuid: mcUuidSchema.optional(),
    // Estado en línea (booleano interno cambiable por el dueño).
    isOnline: z.boolean().optional(),
  })
  .strict()
  .refine((v) => Object.keys(v).length > 0, { message: "Nada que actualizar." });

export type UpdateProfileInput = z.infer<typeof updateProfileSchema>;

/** Presencia propia: encender/apagar el estado en línea (heartbeat del launcher). */
export const presenceUpdateSchema = z
  .object({
    isOnline: z.boolean(),
  })
  .strict();

export type PresenceUpdateInput = z.infer<typeof presenceUpdateSchema>;

/** Cambio de email: la confirmación viaja al correo NUEVO (ver docs). */
export const emailChangeSchema = z
  .object({
    newEmail: z.string().trim().toLowerCase().email("Email inválido.").max(254),
  })
  .strict();

export type EmailChangeInput = z.infer<typeof emailChangeSchema>;

/** Cambio de contraseña: exige la actual (se verifica) + la nueva (8–72). */
export const passwordChangeSchema = z
  .object({
    currentPassword: z.string().min(1, "Contraseña actual requerida.").max(72),
    newPassword: z.string().min(8, "Mínimo 8 caracteres.").max(72, "Máximo 72 caracteres."),
  })
  .strict()
  .refine((v) => v.currentPassword !== v.newPassword, {
    message: "La nueva contraseña debe ser distinta.",
  });

export type PasswordChangeInput = z.infer<typeof passwordChangeSchema>;
