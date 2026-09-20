// Búsqueda pública de jugadores v1 (respeta `searchable` a nivel SQL).
// Si el usuario desactivó "poder buscarte", NO aparece aquí.

import { Hono } from "hono";
import type { AppEnv } from "../../types/app";
import { getEnv } from "../../env";
import { ok, fail } from "../../lib/respond";
import { supabaseAdmin } from "../../lib/supabase";
import { requireAuth } from "../../middleware/auth";
import { rateLimit, rateLimitPresets } from "../../middleware/rate-limit";
import { validateQuery } from "../../middleware/validate";
import { userSearchSchema } from "../../schemas/v1/friends";

export const userRoutes = new Hono<AppEnv>();

// GET /v1/users/search?q=ste&limit=10 — buscar jugadores visibles.
userRoutes.get(
  "/search",
  rateLimit(rateLimitPresets.account),
  requireAuth(),
  validateQuery(userSearchSchema),
  async (c) => {
    const user = c.get("user");
    if (!user) {
      return fail(c, { code: "unauthorized", message: "Autenticación requerida.", status: 401 });
    }
    let env;
    try {
      env = getEnv(c);
    } catch {
      return fail(c, { code: "server_misconfigured", message: "Servicio no disponible.", status: 500 });
    }
    const q = c.req.valid("query");
    const { data, error } = await supabaseAdmin(env).rpc("search_users", {
      p_query: q.q,
      p_limit: q.limit,
    });
    if (error) {
      console.error(`[${c.get("requestId")}] users search:`, error.message);
      return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
    }
    const rows = (data ?? []) as Array<{
      username: string;
      display_name: string;
      avatar_url: string | null;
      mc_uuid: string | null;
      is_online: boolean;
    }>;
    return ok(c, {
      users: rows.map((r) => ({
        username: r.username,
        displayName: r.display_name,
        avatarUrl: r.avatar_url,
        mcUuid: r.mc_uuid,
        isOnline: r.is_online ?? false,
      })),
    });
  },
);
