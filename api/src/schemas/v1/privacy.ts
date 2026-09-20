// Contratos de entrada v1 para privacidad de la cuenta.
// Versionados: jamás reutilizar en v2. Todo opcional, al menos un campo.

import { z } from "zod";

/** Opciones de privacidad: quién puede encontrarte y qué puede llegarte. */
export const privacyUpdateSchema = z
  .object({
    searchable: z.boolean().optional(),
    allowEmailSearch: z.boolean().optional(),
    receiveFriendRequests: z.boolean().optional(),
    receiveNotifications: z.boolean().optional(),
  })
  .strict()
  .refine((v) => Object.keys(v).length > 0, { message: "Nada que actualizar." });

export type PrivacyUpdateInput = z.infer<typeof privacyUpdateSchema>;
