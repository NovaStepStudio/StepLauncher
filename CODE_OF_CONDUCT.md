# Código de Conducta

## Nuestro Compromiso

Con el objetivo de fomentar un ambiente abierto, inclusivo y respetuoso, nos comprometemos como **colaboradores**, **mantenedores** y **miembros de la comunidad** a hacer de la participación en StepLauncher y sus proyectos asociados una experiencia libre de acoso para todos, independientemente de su:

- Edad, complexión física o discapacidad.
- Identidad de género, expresión u orientación sexual.
- Nacionalidad, etnia, religión o nivel socioeconómico.
- Experiencia técnica o nivel de habilidad.

---

## Nuestros Estándares

### Comportamiento Aceptable ✅

- Usar un lenguaje acogedor e inclusivo.
- Respetar los diferentes puntos de vista y experiencias.
- Aceptar críticas constructivas con gracia y profesionalismo.
- Enfocarse en lo que es mejor para la comunidad en general.
- Mostrar empatía hacia los demás miembros de la comunidad.
- **Priorizar la colaboración y el código por sobre el ego.**
- **Cuidar los recursos compartidos**: la configuración y los registros de los usuarios son sensibles; nunca compartirlos en público sin antes verificar que no contengan tokens, credenciales o datos personales.

### Comportamiento Inaceptable ❌

- **Acoso, intimidación o discriminación** de cualquier tipo, incluyendo comentarios sexistas, racistas, homofóbicos, transfóbicos o excluyentes.
- **Lenguaje o imágenes sexuales** en espacios públicos del proyecto (issues, PRs, Discord, foros).
- **Troleo, insultos, ataques personales o comentarios despectivos.**
- **Acoso público o privado**, incluyendo mensajes no solicitados, doxeo o persecución.
- **Publicación de información privada de terceros** sin su consentimiento explícito (direcciones, correos electrónicos, datos personales).
- **Spam, flooding, publicidad no solicitada o promoción de proyectos no relacionados.**
- **Compartir tokens de cuentas, credenciales o archivos de configuración reales** de usuarios (`launcher_accounts.json`, `launcher_config.json`, tokens de sesión) fuera del contexto privado y necesario.
- **Cualquier otra conducta que sea razonablemente considerada inapropiada en un entorno profesional y colaborativo.**
- **Abuso del sistema de reportes** o denuncias falsas.

---

## Comportamiento Técnico

Además del comportamiento interpersonal, esperamos que los colaboradores mantengan ciertos estándares técnicos alineados con la arquitectura del proyecto (`AGENTS.md` y los registros de `Changelogs/`):

### ✅ Buenas Prácticas

- Mantener el código limpio, documentado y siguiendo las convenciones del proyecto.
- **Respetar la arquitectura**: bindings finos en `app.go` que delegan en `internal/Handlers` (motor en `internal/Handlers/Engine`, integrado en el programa), sin romper la separación de capas.
- **Cuidar la concurrencia**: usar `sync.RWMutex`, respetar el orden de adquisición y nunca dejar un `Lock()` llamando a funciones que vuelvan a tomar el mismo mutex (aplica a `internal/Config`, `internal/Core/` y los handlers).
- **No bloquear la UI**: los bindings de Wails deben ser rápidos; todo trabajo lento (I/O, red) va en su propia goroutine con eventos/callbacks, nunca esperando indefinidamente.
- **Verificar siempre**: `go build ./...` en el backend y `bun run build` (con `vue-tsc`) en el frontend antes de enviar un PR.
- **Registrar en `Changelogs/`**: cada error corregido y cada cambio relevante, siguiendo la convención de `Changelogs/README.md`, ANTES de dar la tarea por terminada.
- **No exponer secretos**: tokens de cuentas, contraseñas y rutas sensibles nunca deben loguearse ni llegar a la interfaz (solo la vista `AccountInfo` sin credenciales).
- Escribir mensajes de commit descriptivos en español o inglés.
- Reportar bugs con información detallada (versión, sistema operativo, pasos para reproducir, logs).

### ❌ Malas Prácticas

- Introducir código malicioso, backdoors, malware o vulnerabilidades intencionalmente.
- Filtrar o loguear tokens, contraseñas o datos privados de cuentas, o debilitar el cifrado del almacenamiento.
- Introducir deadlocks o bloqueos del hilo principal (self-deadlock con `sync.RWMutex`, `WaitGroup` a la espera de la misma goroutine, etc.).
- Cambiar la estructura de bindings, agregar campos o renombrar métodos públicos sin justificar el impacto en el frontend generado (`frontend/wailsjs`).
- Realizar cambios masivos sin coordinación con el equipo.
- Ignorar revisiones de código o comentarios de los mantenedores.
- Hacer fork con fines maliciosos o de suplantación de identidad.

---

## Distribución y Redistribución

StepLauncher es **software libre bajo GPL-3.0**, pero su redistribución está sujeta a condiciones propias del proyecto que TODO colaborador, fork o derivado debe respetar:

### Condiciones para redistribuir

- ❌ **NO se permite redistribuir** StepLauncher, ni versiones modificadas o derivadas del mismo, **sin una explicación** de qué es y qué cambia, o **sin un panel visible** en la aplicación donde se indique explícitamente que está **construido con la base de StepLauncher** (con créditos a NovaStepStudio y enlace al repositorio original).
- ✅ **Toda redistribución debe pasar primero por el Discord oficial de StepLauncher** (comunicar la intención al equipo) **y por GitHub Issues** (abrir una issue en este repositorio notificando la redistribución, con el vínculo al proyecto derivado).
- ✅ **Cumplir la GPL-3.0**: ofrecer el código fuente, mantener los avisos de licencia y autoría, y no restringir los derechos de los usuarios finales.
- ❌ **Prohibido** suplantar a StepLauncher o a NovaStepStudio, usar su nombre o marca para engañar usuarios, o lucrarse del proyecto sin cumplir la GPL-3.0.

**Quien redistribuya sin cumplir estas condiciones está fuera de la comunidad del proyecto**: los mantenedores lo notificarán y tomarán las medidas legales y comunitarias que correspondan.

---

## Responsabilidades de los Mantenedores

Los mantenedores del proyecto tienen la responsabilidad de:

1. **Aclarar y hacer cumplir** los estándares de comportamiento aceptable.
2. **Tomar decisiones correctivas** que sean apropiadas y justas ante cualquier comportamiento inaceptable.
3. **Revisar y responder** a los reportes de violaciones de este código dentro de un plazo razonable.
4. **Editar, cerrar o eliminar** comentarios, commits, issues y otras contribuciones que violen este código.
5. **Aplicar medidas temporales o permanentes** contra cualquier colaborador por comportamientos que consideren inapropiados, amenazantes, ofensivos o dañinos.

---

## Alcance

Este Código de Conducta aplica en **todos los espacios del proyecto**, incluyendo:

- Repositorio de GitHub (issues, PRs, discussions, wikis).
- Servidor oficial de Discord.
- Redes sociales oficiales del proyecto.
- Eventos o reuniones relacionadas con el proyecto.
- Comunicaciones privadas dentro del contexto del proyecto.

También aplica cuando un individuo representa oficialmente al proyecto en espacios públicos.

---

## Aplicación

### Reportar Violaciones

Si presencias o experimentas algún comportamiento que viole este código de conducta (incluidas las normas de distribución y redistribución), repórtalo a través de:

| Canal | Detalle |
|---|---|
| **Email** | stepnicka012@gmail.com |
| **Discord** | Contactá directamente con los administradores del servidor oficial |
| **GitHub** | Enviá un mensaje privado a los mantenedores del repositorio o abrí una issue |

**Todos los reportes serán revisados e investigados** de manera confidencial, justa y oportuna. El equipo del proyecto está obligado a mantener la confidencialidad de la persona que reporta el incidente.

### Proceso de Resolución

1. **Reporte:** Se recibe el reporte y se acusa recibo en un plazo máximo de 72 horas.
2. **Investigación:** Se investiga el incidente de manera imparcial.
3. **Decisión:** Se determina la acción correctiva apropiada.
4. **Notificación:** Se informa a las partes involucradas sobre la decisión.
5. **Seguimiento:** Se da seguimiento para asegurar que la resolución sea efectiva.

### Consecuencias

Las consecuencias por comportamiento inaceptable pueden incluir:

| Acción | Descripción |
|---|---|
| **Corrección** | Advertencia privada por escrito explicando la falta y los cambios esperados |
| **Advertencia** | Advertencia pública con consecuencias si continúa el comportamiento |
| **Expulsión temporal** | Suspensión temporal de la participación en el proyecto y espacios asociados |
| **Expulsión permanente** | Prohibición permanente de participar en el proyecto y espacios asociados |
| **Bloqueo de distribución** | Notificación oficial de incumplimiento de las normas de redistribución, con acciones legales si corresponde |

Las consecuencias se determinarán caso por caso, considerando la gravedad, frecuencia e intencionalidad del incidente.

---

## Atribución

Este Código de Conducta está adaptado del [Contributor Covenant](https://www.contributor-covenant.org), versión 2.1, disponible en [https://www.contributor-covenant.org/version/2/1/code_of_conduct.html](https://www.contributor-covenant.org/version/2/1/code_of_conduct.html).

---

<div align="center">
  <sub>© 2026 NovaStepStudio — Santiago Stepnicka</sub>
</div>