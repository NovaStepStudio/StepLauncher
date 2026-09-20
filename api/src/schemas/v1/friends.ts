// Contratos de entrada v1 para amistades y búsqueda de jugadores.
// Versionados: jamás reutilizar en v2.

import { z } from "zod";

/** Enviar solicitud: email, nombre de usuario o UUID (user_id o mc_uuid). */
export const sendFriendRequestSchema = z
  .object({
    identifier: z.string().trim().min(1, "Identificador requerido.").max(254),
  })
  .strict();
export type SendFriendRequestInput = z.infer<typeof sendFriendRequestSchema>;

/** Listar solicitudes: recibidas, enviadas o todas. */
export const listRequestsSchema = z.object({
  type: z.enum(["incoming", "sent", "all"]).default("incoming"),
});
export type ListRequestsInput = z.infer<typeof listRequestsSchema>;

/** `:id` de solicitud (UUID). */
export const requestIdSchema = z.object({
  id: z.string().uuid("Identificador inválido."),
});

/** `:friendId` para romper amistad (UUID de user_id). */
export const friendIdSchema = z.object({
  friendId: z.string().uuid("Identificador inválido."),
});

/** Bloquear: email, nombre de usuario o UUID (la privacidad no impide bloquear). */
export const blockUserSchema = z
  .object({
    identifier: z.string().trim().min(1, "Identificador requerido.").max(254),
  })
  .strict();
export type BlockUserInput = z.infer<typeof blockUserSchema>;

/** `:userId` para desbloquear o consultar bloqueo (UUID de user_id). */
export const blockUserIdSchema = z.object({
  userId: z.string().uuid("Identificador inválido."),
});

/** Búsqueda de jugadores por nombre (respeta `searchable`). */
export const userSearchSchema = z.object({
  q: z.string().trim().min(1, "Búsqueda vacía.").max(40),
  limit: z.coerce.number().int().min(1).max(20).default(10),
});
export type UserSearchInput = z.infer<typeof userSearchSchema>;
