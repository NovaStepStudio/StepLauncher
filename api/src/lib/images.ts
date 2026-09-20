// Análisis de imágenes por cabecera (sin fiarse del MIME declarado).
// Usado por el banner 1080p: formato real + dimensiones exactas desde los bytes.
// Soporta PNG, GIF, JPEG y WebP (VP8, VP8L y VP8X). Null = desconocido/corrupto.

export type ImageFormat = "png" | "gif" | "jpg" | "webp";

export interface ImageInfo {
  format: ImageFormat;
  width: number;
  height: number;
}

/** Primeros N bytes del archivo (las cabeceras viven al inicio; SOF JPEG < 64KB). */
async function headBytes(file: File, n = 65536): Promise<Uint8Array> {
  const buf = await file.slice(0, Math.min(file.size, n)).arrayBuffer();
  return new Uint8Array(buf);
}

function u16be(b: Uint8Array, o: number): number {
  return (b[o]! << 8) | b[o + 1]!;
}

function u16le(b: Uint8Array, o: number): number {
  return b[o]! | (b[o + 1]! << 8);
}

function u32be(b: Uint8Array, o: number): number {
  return (b[o]! * 0x1000000) + ((b[o + 1]! << 16) | (b[o + 2]! << 8) | b[o + 3]!);
}

function u32le(b: Uint8Array, o: number): number {
  return b[o]! | (b[o + 1]! << 8) | (b[o + 2]! << 16) | b[o + 3]! * 0x1000000;
}

function ascii(b: Uint8Array, o: number, len: number): string {
  let s = "";
  for (let i = 0; i < len; i++) s += String.fromCharCode(b[o + i]!);
  return s;
}

function parsePng(b: Uint8Array): ImageInfo | null {
  // Firma (8) + longitud IHDR (4) + "IHDR" (4) + ancho (4 BE) + alto (4 BE).
  if (b.length < 24 || ascii(b, 12, 4) !== "IHDR") return null;
  return { format: "png", width: u32be(b, 16), height: u32be(b, 20) };
}

function parseGif(b: Uint8Array): ImageInfo | null {
  // "GIF87a/89a" (6) + ancho (2 LE) + alto (2 LE).
  if (b.length < 10) return null;
  const sig = ascii(b, 0, 6);
  if (sig !== "GIF87a" && sig !== "GIF89a") return null;
  return { format: "gif", width: u16le(b, 6), height: u16le(b, 8) };
}

function parseJpeg(b: Uint8Array): ImageInfo | null {
  // SOI + segmentos; dimensiones en el marcador SOF (FFC0–C3, C5–C7, C9–CB, CD–CF).
  if (b.length < 4 || b[0] !== 0xff || b[1] !== 0xd8) return null;
  let o = 2;
  for (let steps = 0; steps < 100 && o + 4 <= b.length; steps++) {
    if (b[o] !== 0xff) return null;
    while (b[o] === 0xff) o++; // relleno
    const marker = b[o]!;
    o++;
    if (marker === 0xd8) continue; // SOI repetido
    if (marker === 0xd9) return null; // EOI sin SOF
    if (o + 2 > b.length) return null;
    if (
      marker >= 0xc0 && marker <= 0xcf &&
      marker !== 0xc4 && marker !== 0xc8 && marker !== 0xcc
    ) {
      if (o + 7 > b.length) return null;
      return { format: "jpg", width: u16be(b, o + 5), height: u16be(b, o + 3) };
    }
    const segLen = u16be(b, o);
    if (segLen < 2) return null;
    o += segLen;
  }
  return null;
}

function parseWebp(b: Uint8Array): ImageInfo | null {
  // "RIFF" tamaño "WEBP" + chunk (FourCC + tamaño + datos).
  if (b.length < 20 || ascii(b, 0, 4) !== "RIFF" || ascii(b, 8, 4) !== "WEBP") return null;
  const fourcc = ascii(b, 12, 4);
  const d = 20; // inicio de datos del chunk
  if (fourcc === "VP8X") {
    if (b.length < d + 10) return null;
    const w = (b[d + 4]! | (b[d + 5]! << 8) | (b[d + 6]! << 16)) + 1;
    const h = (b[d + 7]! | (b[d + 8]! << 8) | (b[d + 9]! << 16)) + 1;
    return { format: "webp", width: w, height: h };
  }
  if (fourcc === "VP8L") {
    if (b.length < d + 5 || b[d] !== 0x2f) return null;
    const b1 = b[d + 1]!;
    const b2 = b[d + 2]!;
    const b3 = b[d + 3]!;
    const b4 = b[d + 4]!;
    const w = (((b2 & 0x3f) << 8) | b1) + 1;
    const h = (((b4 & 0x0f) << 10) | (b3 << 2) | (b2 >> 6)) + 1;
    return { format: "webp", width: w, height: h };
  }
  if (fourcc === "VP8 ") {
    // Tag (3) + start code 9D 01 2A (3) + ancho (2, 14 bits) + alto (2, 14 bits).
    if (b.length < d + 10) return null;
    if (b[d + 3] !== 0x9d || b[d + 4] !== 0x01 || b[d + 5] !== 0x2a) return null;
    return {
      format: "webp",
      width: u16le(b, d + 6) & 0x3fff,
      height: u16le(b, d + 8) & 0x3fff,
    };
  }
  return null;
}

/** Formato y dimensiones reales del archivo. Null si no es PNG/GIF/JPEG/WebP válido. */
export async function imageInfo(file: File): Promise<ImageInfo | null> {
  if (file.size < 10) return null;
  const b = await headBytes(file);
  if (
    b.length >= 8 &&
    b[0] === 0x89 && b[1] === 0x50 && b[2] === 0x4e && b[3] === 0x47 &&
    b[4] === 0x0d && b[5] === 0x0a && b[6] === 0x1a && b[7] === 0x0a
  ) {
    return parsePng(b);
  }
  if (b.length >= 12) {
    const webp = parseWebp(b);
    if (webp) return webp;
  }
  if (b.length >= 6 && (ascii(b, 0, 6) === "GIF87a" || ascii(b, 0, 6) === "GIF89a")) {
    return parseGif(b);
  }
  if (b.length >= 2 && b[0] === 0xff && b[1] === 0xd8) {
    return parseJpeg(b);
  }
  return null;
}
