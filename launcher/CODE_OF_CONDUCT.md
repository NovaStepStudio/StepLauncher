# Código de Conducta — StepLauncher

> **Última actualización: 19 de mayo de 2026** · En sintonía con la [Política de Privacidad](https://steplauncher.pages.dev/PrivacyPolicy) y los [Términos y Condiciones](https://steplauncher.pages.dev/TermsAndConditions) del ecosistema StepLauncher.

Bienvenido a **StepLauncher**, el launcher más orgánico y versátil para **Minecraft: Java Edition**, creado por **NovaStepStudio** y potenciado por **NovaCore Engine** (Wails v3 + Go + Vue 3). Este código existe para que colaborar, jugar y crear aquí sea una experiencia segura, respetuosa y divertida para todos — ya seas contribuidor, jugador, diseñador de skins o mantenedor.

Al participar en cualquier espacio del proyecto aceptas este Código y los [Términos](https://steplauncher.pages.dev/TermsAndConditions). Si no estás de acuerdo, por favor no utilices la plataforma.

---

## 1. Nuestro compromiso

Como **colaboradores, mantenedores y miembros de la comunidad** nos comprometemos a hacer de StepLauncher — incluyendo el **launcher de escritorio**, el **panel web (`/Dashboard`)**, la **API pública**, Discord y GitHub — un entorno **abierto, inclusivo y libre de acoso** para todos, sin importar:

- Edad, complexión física o discapacidad.
- Identidad o expresión de género, orientación sexual.
- Nacionalidad, etnia, idioma, religión o nivel socioeconómico.
- Experiencia técnica, rol (jugador, modder, artista) o nivel de habilidad.

Creemos firmemente que **tú eres dueño de tu cuenta y de tu contenido**. StepLauncher actúa únicamente como guardián técnico: puedes ver, editar, exportar o eliminar todos tus datos en cualquier momento desde tu panel de control.

---

## 2. Nuestros estándares

### 2.1 Comportamiento esperado ✅

- Usar lenguaje acogedor, empático e inclusivo. Saluda, agradece y celebra el trabajo ajeno.
- Respetar puntos de vista, experiencias y niveles diferentes. Asumir buena fe primero.
- Aceptar críticas constructivas con profesionalismo y ofrecer feedback accionable, no descalificaciones.
- Priorizar lo que es mejor para la comunidad por sobre el ego personal.
- **Colaborar antes que competir**: compartir conocimiento, documentar y hacer reviews amables.
- Cuidar los recursos compartidos — canales de Discord, issues, PRs y la infraestructura (Supabase, API, rate limits).

### 2.2 Comportamiento inaceptable ❌

No se tolerará, ni en espacios públicos ni privados del proyecto:

- **Acoso, intimidación, discriminación** o comentarios excluyentes (sexistas, racistas, homófobos, tránsfobos, etc.).
- **Lenguaje o imágenes sexuales** no solicitadas en issues, PRs, Discord, skins, capas, biografías o avatares.
- **Troleo, insultos, ataques personales** y descalificaciones.
- **Acoso público o privado**: mensajes no solicitados, persecución, doxeo o publicación de datos privados de terceros sin consentimiento (dirección, email, IP, etc.).
- **Contenido NSFW, gore, violento, ilegal o que viole derechos de terceros** — incluyendo skins, capas, kits, biografías y nombres de usuario.
- **Spam, flooding, publicidad no solicitada** o promoción de proyectos no relacionados.
- **Suplantación de identidad**, robo de cuentas o intento de vulnerar la seguridad de otros usuarios.
- **Evasión de baneos** mediante cuentas alternativas.
- **Abuso del sistema de reportes** con denuncias falsas o malintencionadas.
- **Comercializar o revender el acceso a StepLauncher** con fines maliciosos o para vulnerar la plataforma.

> Estas prohibiciones son las mismas que rigen la **sección 4 (Uso prohibido)** y **sección 11 (Cosméticos)** de los Términos. Su incumplimiento puede derivar en suspensión o eliminación inmediata de la cuenta.

---

## 3. Estándares técnicos y de seguridad

Además de la convivencia, esperamos un nivel técnico alineado con la arquitectura descrita en `AGENTS.md` y `Changelogs/`:

### ✅ Buenas prácticas

- Mantener el código **limpio, documentado y tipado** (Go + `vue-tsc`) y respetar la arquitectura feature-first por dominios.
- **Respetar la separación de capas**: bindings finos en `app.go` que delegan en `internal/Handlers` / `internal/Handlers/Engine` y `internal/Core`. No romper contratos de `frontend/wailsjs`.
- **Concurrencia segura**: usar `sync.RWMutex` correctamente, respetar el orden de adquisición y nunca mantener un `Lock()` durante I/O o callbacks (aplica a `internal/Config`, `internal/Core` y handlers). Evitar self-deadlocks.
- **No bloquear la UI**: los bindings de Wails deben retornar rápido; todo trabajo lento (descargas, red, escaneo de música) va en goroutines con eventos/callbacks.
- **Verificar siempre** antes de un PR: `go vet ./...` / `go build ./...` en backend y `bun run build` (con `vue-tsc`) en frontend.
- **Registrar en `Changelogs/`** cada fix y cambio relevante siguiendo `Changelogs/README.md` antes de dar la tarea por terminada.
- **Seguridad por defecto**: no loguear ni exponer secretos. Las contraseñas se almacenan con **PBKDF2 + salt única (100.000 iteraciones, SHA-256)**, sesiones con tokens aleatorios y expiración automática, comunicaciones por **TLS 1.3**. Las IPs nunca se guardan en logs públicos sin enmascarar.
- Mensajes de commit descriptivos (español o inglés) y PRs pequeños y enfocados.
- Reportar bugs con versión, SO, pasos para reproducir y logs relevantes.

### ❌ Malas prácticas

- Introducir código malicioso, backdoors, malware o vulnerabilidades intencionales.
- Filtrar o loguear tokens, contraseñas o datos privados, o debilitar el cifrado del almacenamiento (`launcher_accounts.json`, `launcher_config.json`).
- Introducir deadlocks o bloqueos del hilo principal (`sync.RWMutex` mal usado, `WaitGroup` esperado por la misma goroutine).
- Cambiar bindings, añadir campos o renombrar métodos públicos sin justificar el impacto en el frontend generado.
- Hacer `scraping` masivo, saturar endpoints o evadir `rate limiting` de la API.
- Hacer ingeniería inversa, descompilar o extraer el código del launcher o la API sin autorización.
- Cambios masivos sin coordinación o ignorar revisiones de los mantenedores.
- Forks con fines maliciosos o de suplantación de identidad.

---

## 4. Privacidad y protección de datos

StepLauncher es **privacy-first**. Lo que aplica en la [Política de Privacidad](https://steplauncher.pages.dev/PrivacyPolicy) aplica aquí:

- **Tus datos son tuyos.** Puedes acceder, rectificar, portar o **eliminar tu cuenta y todos tus datos** de forma irreversible desde `/Dashboard` (o solicitando copia a `privacy@steplauncher.com`). La eliminación borra skins, capas, kits, amigos, likes, seguimientos y enlaces; las copias de seguridad se purgan en 30 días.
- **Nunca compartas credenciales reales** en público: tokens de sesión, `launcher_accounts.json` o `launcher_config.json`. Si necesitas ayuda, comparte logs sanitizados.
- **Qué recopilamos** (resumen): cuenta (hash de contraseña), perfil (bio, avatar, `skin_url`/`cape_url`), cosméticos, preferencias de privacidad, amigos/solicitudes/likes, enlaces sociales, sesión (token, fuente web/launcher), presencia (estado, mensaje), estadísticas de juego y datos de Minecraft (UUID, username). Vía OAuth (Google/Microsoft) solo recibimos ID único, email, nombre y avatar — **nunca tu contraseña de Google/Microsoft**.
- **Qué no hacemos**: no vendemos datos, no usamos cookies de rastreo ni publicidad conductual. Solo usamos `oauth_state` (CSRF, 10 min), `token`/`user` en `localStorage` y `launcher_token` en almacenamiento seguro. Compartimos datos solo con proveedores esenciales: **Supabase** (BD/almacenamiento), **Google OAuth**, **Microsoft / Xbox Live** y **Minecraft Services**.
- **Retención**: cuenta hasta que la elimines (`DELETE /me`); sesiones 1 día (web) / 7 días (launcher); cosméticos hasta que los borres; logs 30 días (anonimizados a los 7).
- **Tus derechos (RGPD / CPRA / LGPD)**: acceder, rectificar, eliminar, restringir, portar, oponerte y retirar el consentimiento (desvincular OAuth). Respuesta en **máximo 30 días** vía `privacy@steplauncher.com`.
- **Menores**: StepLauncher no está dirigido a menores de 13 años (o la edad de consentimiento digital de tu país). Detectamos y eliminamos esas cuentas.
- **Seguridad**: si encuentras una vulnerabilidad, no la publiques. Repórtala de forma privada (ver §7).

---

## 5. Contenido social y cosméticos

Todo lo que subes (skins, capas, kits, biografías) **sigue siendo tuyo**, pero al publicarlo otorgas a StepLauncher una licencia mundial, no exclusiva y libre de regalías para mostrarlo dentro de la plataforma. Nos reservamos el derecho de **moderar, ocultar o eliminar** cualquier contenido que viole los Términos sin previo aviso.

Recuerda:

- No subas contenido plagiado ni suplantando a otros usuarios o cuentas oficiales.
- No incruste malware, scripts o código ejecutable en archivos subidos.
- Configura tu privacidad (quién ve tu skin, estado online, estadísticas) desde el panel.

---

## 6. Distribución y redistribución

StepLauncher es **software libre bajo GPL-3.0**, pero su marca y redistribución tienen condiciones propias que todo colaborador, fork o derivado debe respetar:

- ❌ **No redistribuyas** StepLauncher ni derivados **sin explicar qué es y qué cambia**, y **sin un panel visible** en la app que indique explícitamente que está **construido con la base de StepLauncher** (créditos a **NovaStepStudio** + enlace al repositorio original).
- ✅ **Toda redistribución debe notificarse primero** vía **Discord oficial** (comunicar intención al equipo) **y GitHub Issues** (abrir una issue en este repositorio con el vínculo al proyecto derivado).
- ✅ **Cumplir GPL-3.0**: ofrecer código fuente, mantener avisos de licencia y autoría, y no restringir derechos del usuario final.
- ❌ Prohibido suplantar a StepLauncher o NovaStepStudio, usar su nombre/marca/logo para engañar, o lucrarse sin cumplir la GPL-3.0.

Quien incumpla quedará fuera de la comunidad; los mantenedores notificarán y tomarán medidas legales y comunitarias.

---

## 7. Moderación, baneos y apelaciones

Para mantener una comunidad sana, el equipo de moderación puede actuar de forma **automática** (patrones de abuso, rate limits) o **manual**:

| Medida | Cuándo aplica |
|---|---|
| **Suspensión temporal de funciones** | Spam leve, uso indebido de API o social |
| **Baneo temporal** | Acoso, toxicidad, contenido inapropiado, abuso de API, evasión de límites |
| **Baneo permanente y eliminación de cuenta** | Infracción grave, malware, NSFW, ilegal, reincidencia |
| **Eliminación de contenido** | Skins/capas/bios que violen las normas |
| **Revocación de sesiones/tokens** | En cualquier baneo; se revocan todas las sesiones activas |

**Efectos de un baneo:** no puedes iniciar sesión ni usar el launcher, tus datos permanecen almacenados pero inaccesibles hasta que expire el baneo o se elimine la cuenta, y recibirás notificación del motivo y duración cuando sea posible.

**Apelaciones:** puedes contactar al equipo para revisión. Nos reservamos el derecho de no divulgar detalles específicos por seguridad, pero intentaremos responder de forma justa y oportuna. Si tu cuenta fue eliminada por infracción grave tras revisión manual, la decisión es definitiva. Recuerda que **tú puedes eliminar tu propia cuenta en cualquier momento** sin intervención administrativa.

---

## 8. Alcance

Este Código aplica en **todos los espacios del proyecto**, incluyendo:

- Repositorio de GitHub (issues, PRs, discussions, wikis, commits).
- Servidor oficial de Discord y redes sociales oficiales.
- Launcher de escritorio, panel web, API y eventos relacionados.
- Comunicaciones privadas **dentro del contexto del proyecto** y cuando alguien represente oficialmente al proyecto en público.

También rige el comportamiento en espacios externos si afecta la seguridad o la reputación de la comunidad StepLauncher.

---

## 9. Responsabilidades de los mantenedores

Los mantenedores se comprometen a:

1. **Aclarar y hacer cumplir** los estándares de comportamiento aceptable.
2. **Tomar decisiones correctivas** proporcionales, justas y consistentes.
3. **Revisar y responder** a reportes en un plazo razonable (acuse de recibo en **≤ 72 h**).
4. **Editar, cerrar o eliminar** comentarios, commits, issues y contribuciones que violen este código.
5. **Aplicar medidas temporales o permanentes** contra colaboradores por conductas inapropiadas.

Toda acción se documentará con transparencia cuando sea posible sin comprometer privacidad o seguridad.

---

## 10. Aplicación

### 10.1 Reportar violaciones

Si presencias o experimentas una violación (incluidas normas de distribución, privacidad o contenido), repórtala por el canal que te sea más cómodo. **Todos los reportes se tratan de forma confidencial**:

| Canal | Detalle |
|---|---|
| **Email privacidad** | `privacy@steplauncher.com` |
| **Email seguridad** | `security@steplauncher.com` |
| **Email legal / general** | `legal@steplauncher.com` · `stepnicka012@gmail.com` |
| **Discord** | Mensaje directo a administradores/moderadores del servidor oficial |
| **GitHub** | Mensaje privado a mantenedores o issue confidencial |

> **Consejo:** no publiques pruebas con datos sensibles en público. Comparte capturas/logs sanitizados y envía los detalles completos por un canal privado.

### 10.2 Proceso de resolución

1. **Reporte y acuse** — recepción y confirmación en ≤ 72 h.
2. **Investigación imparcial** — el equipo revisa evidencias y contexto.
3. **Decisión** — se determina la acción correctiva proporcional.
4. **Notificación** — se informa a las partes sobre la resolución (y motivo/duración en caso de baneo).
5. **Seguimiento y apelación** — se monitorea y se atiende apelación si se solicita.

### 10.3 Consecuencias

| Acción | Descripción |
|---|---|
| **Corrección** | Advertencia privada por escrito y guía de cambios esperados |
| **Advertencia** | Apercibimiento con consecuencias si reincide |
| **Suspensión temporal** | Pérdida temporal de funciones o acceso (vía baneo temporal + revocación de tokens) |
| **Expulsión permanente** | Prohibición de participar en proyecto y espacios asociados |
| **Bloqueo de distribución** | Notificación oficial de incumplimiento y acciones legales si corresponde |

Las decisiones consideran gravedad, frecuencia, intencionalidad e historial. La reincidencia o el engaño agravan la sanción.

---

## 11. Atribución y documentos relacionados

Este Código está adaptado del **[Contributor Covenant](https://www.contributor-covenant.org) v2.1** ([texto original](https://www.contributor-covenant.org/version/2/1/code_of_conduct.html)) y ampliado con las normas específicas del ecosistema StepLauncher.

Consulta también:

- [Política de Privacidad](https://steplauncher.pages.dev/PrivacyPolicy) — qué datos recopilamos, cómo los usamos y tus derechos.
- [Términos y Condiciones](https://steplauncher.pages.dev/TermsAndConditions) — uso permitido/prohibido, moderación, propiedad intelectual y descargos.
- `LICENSE.md` (GPL-3.0) — derechos y obligaciones legales del código.
- `AGENTS.md` y `Changelogs/README.md` — flujo de trabajo y registro de cambios.

Las preguntas sobre este Código pueden enviarse a cualquiera de los canales del §10.1.

---

<div align="center">

**StepLauncher — NovaStepStudio · Powered by NovaCore Engine**

[🌐 Página principal](https://steplauncher.pages.dev) · [Política de Privacidad](https://steplauncher.pages.dev/PrivacyPolicy) · [Términos](https://steplauncher.pages.dev/TermsAndConditions) · [Discord](https://steplauncher.pages.dev) · [Repositorio](https://github.com/NovaStepStudio/StepLauncher)

<sub>© 2026 NovaStepStudio — Santiago Stepnicka · No afiliado a Mojang Studios ni a Microsoft. Minecraft es marca de Mojang Studios y Microsoft.</sub>

</div>
