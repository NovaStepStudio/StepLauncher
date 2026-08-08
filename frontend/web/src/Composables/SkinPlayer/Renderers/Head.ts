import type { SkinBitmapSource } from "../Utils";
import type { RendererOptions } from "./Structures/RendererOptions";
import { canvasToBlob } from "../Utils";

export async function getHead(image: SkinBitmapSource, { inputScale, layers, outputScale }: RendererOptions): Promise<Blob> {

    const canvas = document.createElement('canvas');
    canvas.width = 8 * outputScale;
    canvas.height = 8 * outputScale;
    const ctx = canvas.getContext('2d')!;
    ctx.imageSmoothingEnabled = false;

    ctx.drawImage(image, 8 * inputScale, 8 * inputScale, 8 * inputScale, 8 * inputScale, 0, 0, 8 * outputScale, 8 * outputScale);
    if (layers) ctx.drawImage(image, 40 * inputScale, 8 * inputScale, 8 * inputScale, 8 * inputScale, 0, 0, 8 * outputScale, 8 * outputScale);

    return await canvasToBlob(canvas);

}