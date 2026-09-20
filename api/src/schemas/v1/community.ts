// Contratos de entrada v1 para la comunidad (previsualización de perfiles).
// Versionados: jamás reutilizar en v2.

import { z } from "zod";

/** `:identifier`: UUID de cuenta, nombre de usuario o UUID de Minecraft. */
export const profileIdentifierSchema = z.object({
  identifier: z.string().trim().min(1, "Identificador requerido.").max(254),
});
export type ProfileIdentifierInput = z.infer<typeof profileIdentifierSchema>;
