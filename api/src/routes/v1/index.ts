// Router v1: punto único donde se cuelgan los dominios de la versión.
// Añadir un dominio nuevo = crear su archivo y una línea `route()` aquí.
// v2 será una carpeta `src/routes/v2/` nueva; v1 no se modifica para romper.

import { Hono } from "hono";
import type { AppEnv } from "../../types/app";
import { healthRoutes } from "./health";
import { accountRoutes } from "./accounts";
import { authRoutes } from "./auth";
import { fileRoutes } from "./files";
import { notificationRoutes } from "./notifications";
import { friendRoutes } from "./friends";
import { userRoutes } from "./users";
import { communityRoutes } from "./community";

export const v1 = new Hono<AppEnv>();

v1.route("/health", healthRoutes);
v1.route("/auth", authRoutes);
v1.route("/accounts", accountRoutes);
v1.route("/files", fileRoutes);
v1.route("/notifications", notificationRoutes);
v1.route("/friends", friendRoutes);
v1.route("/users", userRoutes);
v1.route("/community", communityRoutes);
