import type { SkinBitmapSource } from "../Utils";
import type { SkinImageOptions } from "./Structures/SkinImageOptions";
import { getFullBodyLegacy, getFullBodyModern } from "../Renderers/FullBody";

export async function renderFullBody(skin: SkinBitmapSource, options?: SkinImageOptions): Promise<Blob> {

    const inputScale = Math.floor(skin.width / 64);
    const outputScale = options?.scale ?? 1;
    const layers = options?.layers ?? true;
    if (skin.height === Math.floor(skin.width / 2)) {
        return await getFullBodyLegacy(skin, { inputScale, outputScale, layers });
    }
    return await getFullBodyModern(skin, { inputScale, outputScale, layers });

}