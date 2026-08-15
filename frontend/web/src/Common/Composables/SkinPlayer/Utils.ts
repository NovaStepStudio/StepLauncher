export type SkinBitmapSource = ImageBitmap | HTMLImageElement | HTMLCanvasElement;

export interface StrictSkinImageOptions {
    scale: number,
    layers: boolean
}

export async function loadSkin(source: Blob | string): Promise<ImageBitmap> {
    const blob = typeof source === "string" ? await (await fetch(source)).blob() : source;
    return await createImageBitmap(blob);
}

export async function checkSlim(skin: SkinBitmapSource): Promise<boolean> {

    const canvas = document.createElement('canvas');
    canvas.width = 64;
    canvas.height = 64;
    const ctx = canvas.getContext('2d', { willReadFrequently: true })!;
    ctx.imageSmoothingEnabled = false;

    ctx.drawImage(skin, 0, 0, 64, skin.height === Math.floor(skin.width / 2) ? 32 : 64);
    return ctx.getImageData(55, 20, 1, 1).data.every(e => e === 0);

}

export function canvasToBlob(canvas: HTMLCanvasElement): Promise<Blob> {
    return new Promise((resolve, reject) => {
        canvas.toBlob((blob) => {
            if (blob) resolve(blob);
            else reject(new Error("No se pudo generar la imagen PNG"));
        }, "image/png");
    });
}