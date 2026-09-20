import type { SkinBitmapSource } from "../Utils";
import type { SkinImageOptions } from "./Structures/SkinImageOptions";
import { getHead } from "../Renderers/Head";

export async function renderHead(skin: SkinBitmapSource, options?: SkinImageOptions): Promise<Blob> {

    const inputScale = Math.floor(skin.width / 64);
    const outputScale = options?.scale ?? 1;
    const layers = options?.layers ?? true;

    return await getHead(skin, { inputScale, outputScale, layers });

}