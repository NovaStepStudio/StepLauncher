// Notificaciones v1: bandeja propia del jugador (listar, leer, descartar).
// Todo con `requireAuth()` + RLS: un token solo ve las de SU cuenta.

import { Hono } from "hono";
import type { AppEnv } from "../../types/app";
import { getEnv } from "../../env";
import { ok, fail } from "../../lib/respond";
import { supabaseForUser } from "../../lib/supabase";
import { requireAuth, getBearerToken } from "../../middleware/auth";
import { rateLimit, rateLimitPresets } from "../../middleware/rate-limit";
import { validateQuery, validateParam } from "../../middleware/validate";
import { listNotificationsSchema, notificationIdSchema } from "../../schemas/v1/notifications";

export const notificationRoutes = new Hono<AppEnv>();

interface NotificationRow {
  id: string;
  title: string;
  body: string;
  read: boolean;
  created_at: string;
}

function notificationShape(row: NotificationRow) {
  return {
    id: row.id,
    title: row.title,
    body: row.body,
    read: row.read,
    createdAt: row.created_at,
  };
}

// GET /v1/notifications — bandeja propia (+ nº de no leídas).
notificationRoutes.get(
  "/",
  rateLimit(rateLimitPresets.account),
  requireAuth(),
  validateQuery(listNotificationsSchema),
  async (c) => {
    const user = c.get("user");
    const token = getBearerToken(c);
    if (!user || !token) {
      return fail(c, { code: "unauthorized", message: "Autenticación requerida.", status: 401 });
    }
    let env;
    try {
      env = getEnv(c);
    } catch {
      return fail(c, { code: "server_misconfigured", message: "Servicio no disponible.", status: 500 });
    }
    const q = c.req.valid("query");
    const supabase = supabaseForUser(env, token);

    let query = supabase
      .from("notifications")
      .select("id, title, body, read, created_at")
      .eq("user_id", user.id)
      .order("created_at", { ascending: false })
      .range(q.offset, q.offset + q.limit - 1);
    if (q.unreadOnly) query = query.eq("read", false);

    const [{ data, error }, { count: unreadCount }] = await Promise.all([
      query,
      supabase
        .from("notifications")
        .select("id", { count: "exact", head: true })
        .eq("user_id", user.id)
        .eq("read", false),
    ]);
    if (error) {
      console.error(`[${c.get("requestId")}] notifications list:`, error.message);
      return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
    }
    return ok(c, {
      notifications: ((data ?? []) as NotificationRow[]).map(notificationShape),
      unreadCount: unreadCount ?? 0,
      limit: q.limit,
      offset: q.offset,
    });
  },
);

// PATCH /v1/notifications/:id/read — marcar una como leída.
notificationRoutes.patch(
  "/:id/read",
  rateLimit(rateLimitPresets.account),
  requireAuth(),
  validateParam(notificationIdSchema),
  async (c) => {
    const user = c.get("user");
    const token = getBearerToken(c);
    if (!user || !token) {
      return fail(c, { code: "unauthorized", message: "Autenticación requerida.", status: 401 });
    }
    let env;
    try {
      env = getEnv(c);
    } catch {
      return fail(c, { code: "server_misconfigured", message: "Servicio no disponible.", status: 500 });
    }
    const { data, error } = await supabaseForUser(env, token)
      .from("notifications")
      .update({ read: true })
      .eq("id", c.req.valid("param").id)
      .eq("user_id", user.id)
      .select("id, title, body, read, created_at")
      .maybeSingle();
    if (error) {
      console.error(`[${c.get("requestId")}] notifications read:`, error.message);
      return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
    }
    if (!data) {
      return fail(c, { code: "not_found", message: "Recurso no encontrado.", status: 404 });
    }
    return ok(c, notificationShape(data as NotificationRow));
  },
);

// POST /v1/notifications/read-all — marcar todas como leídas.
notificationRoutes.post(
  "/read-all",
  rateLimit(rateLimitPresets.account),
  requireAuth(),
  async (c) => {
    const user = c.get("user");
    const token = getBearerToken(c);
    if (!user || !token) {
      return fail(c, { code: "unauthorized", message: "Autenticación requerida.", status: 401 });
    }
    let env;
    try {
      env = getEnv(c);
    } catch {
      return fail(c, { code: "server_misconfigured", message: "Servicio no disponible.", status: 500 });
    }
    const { error, count } = await supabaseForUser(env, token)
      .from("notifications")
      .update({ read: true }, { count: "exact" })
      .eq("user_id", user.id)
      .eq("read", false);
    if (error) {
      console.error(`[${c.get("requestId")}] notifications read-all:`, error.message);
      return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
    }
    return ok(c, { markedRead: count ?? 0 });
  },
);

// DELETE /v1/notifications/:id — descartar una notificación propia.
notificationRoutes.delete(
  "/:id",
  rateLimit(rateLimitPresets.account),
  requireAuth(),
  validateParam(notificationIdSchema),
  async (c) => {
    const user = c.get("user");
    const token = getBearerToken(c);
    if (!user || !token) {
      return fail(c, { code: "unauthorized", message: "Autenticación requerida.", status: 401 });
    }
    let env;
    try {
      env = getEnv(c);
    } catch {
      return fail(c, { code: "server_misconfigured", message: "Servicio no disponible.", status: 500 });
    }
    const { data, error } = await supabaseForUser(env, token)
      .from("notifications")
      .delete()
      .eq("id", c.req.valid("param").id)
      .eq("user_id", user.id)
      .select("id")
      .maybeSingle();
    if (error) {
      console.error(`[${c.get("requestId")}] notifications delete:`, error.message);
      return fail(c, { code: "internal_error", message: "Error interno.", status: 500 });
    }
    if (!data) {
      return fail(c, { code: "not_found", message: "Recurso no encontrado.", status: 404 });
    }
    return ok(c, { deleted: true });
  },
);
