// Contratos de entrada v1 para autenticación (registro, login, sesiones).
// Versionados: jamás reutilizar en v2. Las contraseñas nunca se registran en logs.

import { z } from "zod";
import { usernameSchema } from "./account";

/** Registro: pide email + usuario + contraseña (los tres, siempre). */
export const registerSchema = z
  .object({
    email: z.string().trim().toLowerCase().email("Email inválido.").max(254),
    username: usernameSchema,
    password: passwordSchema(),
  })
  .strict();

/** Login: acepta email O usuario como identificador + contraseña. */
export const loginSchema = z
  .object({
    identifier: z.string().trim().min(3, "Mínimo 3 caracteres.").max(254),
    password: z.string().min(1, "Contraseña requerida.").max(72),
  })
  .strict();

/** Renovación de sesión con el refresh_token guardado en el dispositivo. */
export const refreshSchema = z
  .object({
    refreshToken: z.string().min(10, "Token inválido.").max(4096),
  })
  .strict();

/** Email genérico para reenvíos y recupero (minúsculas, sin enumerar usuarios). */
const emailSchema = z.string().trim().toLowerCase().email("Email inválido.").max(254);

/** Reenvío de confirmación: registro o cambio de email pendientes. */
export const resendSchema = z
  .object({
    email: emailSchema,
    type: z.enum(["signup", "email_change"]).default("signup"),
  })
  .strict();

/** Solicitud de recupero: siempre responde éxito genérico (sin enumerar). */
export const recoverSchema = z
  .object({
    email: emailSchema,
  })
  .strict();

/** Confirmación con código/OTP del correo (6–1024 caracteres del enlace o código). */
export const confirmSchema = z
  .object({
    email: emailSchema,
    token: z.string().trim().min(6, "Código inválido.").max(1024),
    type: z.enum(["signup", "email_change", "recovery"]),
  })
  .strict();

/** Nueva contraseña tras recupero verificado (el token prueba el correo). */
export const resetPasswordSchema = z
  .object({
    email: emailSchema,
    token: z.string().trim().min(6, "Código inválido.").max(1024),
    newPassword: passwordSchema(),
  })
  .strict();

/** Nueva contraseña con la sesión de recupero (el Bearer ES del enlace recovery). */
export const recoveryPasswordSchema = z
  .object({
    newPassword: passwordSchema(),
  })
  .strict();

/** Contraseña nueva: 8–72 caracteres (límite de bcrypt). Sin complejidad teatral. */
export function passwordSchema() {
  return z.string().min(8, "Mínimo 8 caracteres.").max(72, "Máximo 72 caracteres.");
}

export type RegisterInput = z.infer<typeof registerSchema>;
export type LoginInput = z.infer<typeof loginSchema>;
export type RefreshInput = z.infer<typeof refreshSchema>;
export type ResendInput = z.infer<typeof resendSchema>;
export type RecoverInput = z.infer<typeof recoverSchema>;
export type ConfirmInput = z.infer<typeof confirmSchema>;
export type ResetPasswordInput = z.infer<typeof resetPasswordSchema>;
export type RecoveryPasswordInput = z.infer<typeof recoveryPasswordSchema>;
