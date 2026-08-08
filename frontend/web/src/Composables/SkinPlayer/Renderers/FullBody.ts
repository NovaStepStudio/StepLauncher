import type { SkinBitmapSource } from "../Utils";
import type { RendererOptions } from "./Structures/RendererOptions";
import { canvasToBlob, checkSlim } from "../Utils";

export async function getFullBodyLegacy(image: SkinBitmapSource, { inputScale, layers, outputScale }: RendererOptions): Promise<Blob> {

    const canvas = document.createElement('canvas');
    canvas.width = 16 * outputScale;
    canvas.height = 32 * outputScale;
    const ctx = canvas.getContext('2d')!;
    ctx.imageSmoothingEnabled = false;

    ctx.drawImage(image, 4 * inputScale, 20 * inputScale, 4 * inputScale, 12 * inputScale, 4 * outputScale, 20 * outputScale, 4 * outputScale, 12 * outputScale);

    ctx.save();
    ctx.scale(-1, 1);
    ctx.drawImage(image, 4 * inputScale, 20 * inputScale, 4 * inputScale, 12 * inputScale, -12 * outputScale, 20 * outputScale, 4 * outputScale, 12 * outputScale);
    ctx.restore();

    ctx.drawImage(image, 44 * inputScale, 20 * inputScale, 4 * inputScale, 12 * inputScale, 0, 8 * outputScale, 4 * outputScale, 12 * outputScale);

    ctx.save();
    ctx.scale(-1, 1);
    ctx.drawImage(image, 44 * inputScale, 20 * inputScale, 4 * inputScale, 12 * inputScale, -16 * outputScale, 8 * outputScale, 4 * outputScale, 12 * outputScale);
    ctx.restore();

    ctx.drawImage(image, 8 * inputScale, 8 * inputScale, 8 * inputScale, 8 * inputScale, 4 * outputScale, 0, 8 * outputScale, 8 * outputScale);
    if (layers) ctx.drawImage(image, 40 * inputScale, 8 * inputScale, 8 * inputScale, 8 * inputScale, 4 * outputScale, 0, 8 * outputScale, 8 * outputScale);

    ctx.drawImage(image, 20 * inputScale, 20 * inputScale, 8 * inputScale, 12 * inputScale, 4 * outputScale, 8 * outputScale, 8 * outputScale, 12 * outputScale);

    return await canvasToBlob(canvas);

}

export async function getFullBodyModern(image: SkinBitmapSource, { inputScale, layers, outputScale }: RendererOptions): Promise<Blob> {

    const canvas = document.createElement('canvas');
    canvas.width = 16 * outputScale;
    canvas.height = 32 * outputScale;
    const ctx = canvas.getContext('2d')!;
    ctx.imageSmoothingEnabled = false;

    ctx.drawImage(image, 8 * inputScale, 8 * inputScale, 8 * inputScale, 8 * inputScale, 4 * outputScale, 0, 8 * outputScale, 8 * outputScale);
    if (layers) ctx.drawImage(image, 40 * inputScale, 8 * inputScale, 8 * inputScale, 8 * inputScale, 4 * outputScale, 0, 8 * outputScale, 8 * outputScale);

    ctx.drawImage(image, 20 * inputScale, 20 * inputScale, 8 * inputScale, 12 * inputScale, 4 * outputScale, 8 * outputScale, 8 * outputScale, 12 * outputScale);
    if (layers) ctx.drawImage(image, 20 * inputScale, 36 * inputScale, 8 * inputScale, 12 * inputScale, 4 * outputScale, 8 * outputScale, 8 * outputScale, 12 * outputScale);

    if (await checkSlim(image)) {

        ctx.drawImage(image, 44 * inputScale, 20 * inputScale, 3 * inputScale, 12 * inputScale, 1 * outputScale, 8 * outputScale, 3 * outputScale, 12 * outputScale);
        if (layers) ctx.drawImage(image, 44 * inputScale, 36 * inputScale, 3 * inputScale, 12 * inputScale, 1 * outputScale, 8 * outputScale, 3 * outputScale, 12 * outputScale);

        ctx.drawImage(image, 36 * inputScale, 52 * inputScale, 3 * inputScale, 12 * inputScale, 12 * outputScale, 8 * outputScale, 3 * outputScale, 12 * outputScale);
        if (layers) ctx.drawImage(image, 52 * inputScale, 52 * inputScale, 3 * inputScale, 12 * inputScale, 12 * outputScale, 8 * outputScale, 3 * outputScale, 12 * outputScale);

    } else {

        ctx.drawImage(image, 44 * inputScale, 20 * inputScale, 4 * inputScale, 12 * inputScale, 0, 8 * outputScale, 4 * outputScale, 12 * outputScale);
        if (layers) ctx.drawImage(image, 44 * inputScale, 36 * inputScale, 4 * inputScale, 12 * inputScale, 0, 8 * outputScale, 4 * outputScale, 12 * outputScale);

        ctx.drawImage(image, 36 * inputScale, 52 * inputScale, 4 * inputScale, 12 * inputScale, 12 * outputScale, 8 * outputScale, 4 * outputScale, 12 * outputScale);
        if (layers) ctx.drawImage(image, 52 * inputScale, 52 * inputScale, 4 * inputScale, 12 * inputScale, 12 * outputScale, 8 * outputScale, 4 * outputScale, 12 * outputScale);

    }

    ctx.drawImage(image, 20 * inputScale, 52 * inputScale, 4 * inputScale, 12 * inputScale, 8 * outputScale, 20 * outputScale, 4 * outputScale, 12 * outputScale);
    if (layers) ctx.drawImage(image, 4 * inputScale, 52 * inputScale, 4 * inputScale, 12 * inputScale, 8 * outputScale, 20 * outputScale, 4 * outputScale, 12 * outputScale);

    ctx.drawImage(image, 4 * inputScale, 20 * inputScale, 4 * inputScale, 12 * inputScale, 4 * outputScale, 20 * outputScale, 4 * outputScale, 12 * outputScale);
    if (layers) ctx.drawImage(image, 4 * inputScale, 36 * inputScale, 4 * inputScale, 12 * inputScale, 4 * outputScale, 20 * outputScale, 4 * outputScale, 12 * outputScale);

    return await canvasToBlob(canvas);

}