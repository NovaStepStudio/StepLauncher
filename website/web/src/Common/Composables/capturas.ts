// Capturas de previsualización por dominio (`web/assets/capturesPreview/`).
// Imports estáticos para que Vite las incluya en el bundle final.
// Cada carga de la web elige una al azar del dominio y queda fija
// (no hay rotación dinámica).
import mainMenu from '../../../assets/capturesPreview/MainMenu/MainMenu.png';
import mainMenu001 from '../../../assets/capturesPreview/MainMenu/MainMenu-001.png';
import instances001 from '../../../assets/capturesPreview/Instances/Instances-001.png';
import instances002 from '../../../assets/capturesPreview/Instances/Instances-002.png';
import modsExplorer from '../../../assets/capturesPreview/Mods/ModsExplorer.png';
import skinLayers from '../../../assets/capturesPreview/Mods/3DSkinLayers-Preview.png';
import freshAnimations from '../../../assets/capturesPreview/Mods/FreshAnimations-Preview.png';
import eva3NowPlaying from '../../../assets/capturesPreview/MusicPanel/Eva3MusicPanelNowPlaying.png';
import exploreMusic from '../../../assets/capturesPreview/MusicPanel/ExploreMusicPanel.png';
import homeMusic from '../../../assets/capturesPreview/MusicPanel/HomeMusicPanel.png';
import homeMusic2 from '../../../assets/capturesPreview/MusicPanel/HomeMusicPanel2.png';
import libraryMusic from '../../../assets/capturesPreview/MusicPanel/LibrarySectionMusicPanel.png';
import listMusic from '../../../assets/capturesPreview/MusicPanel/ListReproductionMusicPanel.png';
import nowPlayingMusic from '../../../assets/capturesPreview/MusicPanel/NowPlayingMusicPanel.png';
import playMenu from '../../../assets/capturesPreview/Playing/PlayMenu.png';
import playMenu2 from '../../../assets/capturesPreview/Playing/PlayMenu2.png';
import playing121 from '../../../assets/capturesPreview/Playing/Playing121Mc.png';
import previewStyle from '../../../assets/capturesPreview/Personalization/PreviewStyle.png';
import personConfig from '../../../assets/capturesPreview/Personalization/PersonalizationConfig.png';
import personConfig2 from '../../../assets/capturesPreview/Personalization/PersonalizationConfig2.png';
import welcome001 from '../../../assets/capturesPreview/Welcome/Welcome-001.png';
import welcome002 from '../../../assets/capturesPreview/Welcome/Welcome-002.png';
import welcome003 from '../../../assets/capturesPreview/Welcome/Welcome-003.png';

export type DominioCaptura =
    | 'MainMenu'
    | 'Instances'
    | 'Mods'
    | 'MusicPanel'
    | 'Playing'
    | 'Personalization'
    | 'Welcome';

const POR_DOMINIO: Record<DominioCaptura, string[]> = {
    MainMenu: [mainMenu, mainMenu001],
    Instances: [instances001, instances002],
    Mods: [modsExplorer, skinLayers, freshAnimations],
    MusicPanel: [eva3NowPlaying, exploreMusic, homeMusic, homeMusic2, libraryMusic, listMusic, nowPlayingMusic],
    Playing: [playMenu, playMenu2, playing121],
    Personalization: [previewStyle, personConfig, personConfig2],
    Welcome: [welcome001, welcome002, welcome003],
};

// Una al azar por carga: se calcula en el setup del componente y no cambia más.
export function capturaAlAzar(dominio: DominioCaptura): string {
    const lista = POR_DOMINIO[dominio];
    if (lista.length === 0) return '';
    return lista[Math.floor(Math.random() * lista.length)] ?? lista[0] ?? '';
}
