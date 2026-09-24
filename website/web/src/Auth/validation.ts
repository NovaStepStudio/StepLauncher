// Validaciones de cliente, espejo de la API v1. La API tiene la última palabra.

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/;
const USERNAME_RE = /^[A-Za-z0-9._-]+$/;
const DISPLAY_RE = /^[\p{L}\p{N} _-]+$/u;

export function validarEmail(valor: string): string {
    const email = valor.trim().toLowerCase();
    if (!email) return 'Ingresá tu correo.';
    if (email.length > 254) return 'Ese correo es demasiado largo.';
    if (!EMAIL_RE.test(email)) return 'Ese correo no parece válido.';
    return '';
}

export function validarUsername(valor: string): string {
    const username = valor.trim();
    if (!username) return 'Elegí un nombre de usuario.';
    if (username.length < 3) return 'Mínimo 3 caracteres.';
    if (username.length > 20) return 'Máximo 20 caracteres.';
    if (!USERNAME_RE.test(username)) return 'Solo letras, números y . _ -';
    return '';
}

export function validarDisplayName(valor: string): string {
    const nombre = valor.trim();
    if (!nombre) return 'Elegí un nombre visible.';
    if (nombre.length < 3) return 'Mínimo 3 caracteres.';
    if (nombre.length > 32) return 'Máximo 32 caracteres.';
    if (!DISPLAY_RE.test(nombre)) return 'Solo letras, números, espacios y guiones.';
    return '';
}

export function validarPassword(valor: string): string {
    if (!valor) return 'Ingresá una contraseña.';
    if (valor.length < 8) return 'Mínimo 8 caracteres.';
    if (valor.length > 72) return 'Máximo 72 caracteres.';
    return '';
}

// Bio libre hasta 5000 caracteres; vacía significa sin bio (la API la borra).
export function validarBio(valor: string): string {
    if (valor.length > 5000) return 'Máximo 5000 caracteres.';
    return '';
}

export function validarIdentificador(valor: string): string {
    if (!valor.trim()) return 'Ingresá tu correo o tu usuario.';
    if (valor.trim().length < 3) return 'Mínimo 3 caracteres.';
    return '';
}

// Código de verificación del correo (OTP de 6 dígitos o hash del enlace).
export function validarCodigo(valor: string): string {
    const codigo = valor.trim();
    if (!codigo) return 'Ingresá el código que te llegó por correo.';
    if (codigo.length < 6) return 'Ese código es demasiado corto.';
    if (codigo.length > 4096) return 'Ese código es demasiado largo.';
    return '';
}
