// Contratos de entrada v1 para archivos (skins, capas, avatar).
// Versionados: jamás reutilizar en v2. Los límites de tamaño/tipo se validan
// también en el handler (el multipart no lo cubre Zod).

import { z } from "zod";

/** Tipos de archivo que acepta la API. La quota diaria cubre skin + cape. */
export const fileKindSchema = z.enum(["skin", "cape", "avatar"]);
export type FileKind = z.infer<typeof fileKindSchema>;

/** Paginación para listar subidas propias. */
export const listUploadsSchema = z.object({
  limit: z.coerce.number().int().min(1).max(50).default(20),
  offset: z.coerce.number().int().min(0).default(0),
});
export type ListUploadsInput = z.infer<typeof listUploadsSchema>;
