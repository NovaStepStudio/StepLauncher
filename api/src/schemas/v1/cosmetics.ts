// Contratos de entrada v1 para cosméticos propios (equipar/desequipar).
// Versionados: jamás reutilizar en v2. Solo lo que el jugador YA posee.

import { z } from "zod";

/** Equipar o desequipar un cosmético propio por su id. */
export const cosmeticEquipSchema = z
  .object({
    cosmeticId: z.string().uuid("Identificador inválido."),
  })
  .strict();
export type CosmeticEquipInput = z.infer<typeof cosmeticEquipSchema>;
