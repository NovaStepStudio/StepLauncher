// Contratos de entrada v1 para notificaciones de la cuenta.
// Versionados: jamás reutilizar en v2. Las notificaciones las crea el servidor;
// el jugador solo las lista, las marca como leídas o las descarta.

import { z } from "zod";

/** Listado propio con filtro de no leídas y paginación. */
export const listNotificationsSchema = z.object({
  limit: z.coerce.number().int().min(1).max(50).default(20),
  offset: z.coerce.number().int().min(0).default(0),
  unreadOnly: z.coerce.boolean().default(false),
});
export type ListNotificationsInput = z.infer<typeof listNotificationsSchema>;

/** `:id` de notificación (UUID). */
export const notificationIdSchema = z.object({
  id: z.string().uuid("Identificador inválido."),
});
export type NotificationIdInput = z.infer<typeof notificationIdSchema>;
