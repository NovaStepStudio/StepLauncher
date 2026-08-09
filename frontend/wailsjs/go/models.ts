export namespace Config {
	
	export class BackgroundConfig {
	    type: string;
	    imagePath: string;
	    videoPath: string;
	    dynamicImages: string[];
	    dynamicOrder: string;
	    dynamicInterval: number;
	
	    static createFrom(source: any = {}) {
	        return new BackgroundConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.imagePath = source["imagePath"];
	        this.videoPath = source["videoPath"];
	        this.dynamicImages = source["dynamicImages"];
	        this.dynamicOrder = source["dynamicOrder"];
	        this.dynamicInterval = source["dynamicInterval"];
	    }
	}
	export class ExtraData {
	    assets: string;
	    accounts: string;
	    history: string;
	    profiles: string;
	    crashHistory: string;
	
	    static createFrom(source: any = {}) {
	        return new ExtraData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.assets = source["assets"];
	        this.accounts = source["accounts"];
	        this.history = source["history"];
	        this.profiles = source["profiles"];
	        this.crashHistory = source["crashHistory"];
	    }
	}
	export class RichPresenceConfig {
	    enabled?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RichPresenceConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	    }
	}
	export class IdleConfig {
	    autoCloseModals: boolean;
	    idleMinutes: number;
	    configCheckEnabled: boolean;
	    configCheckMinutes: number;
	
	    static createFrom(source: any = {}) {
	        return new IdleConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.autoCloseModals = source["autoCloseModals"];
	        this.idleMinutes = source["idleMinutes"];
	        this.configCheckEnabled = source["configCheckEnabled"];
	        this.configCheckMinutes = source["configCheckMinutes"];
	    }
	}
	export class ThemeColors {
	    sidebar: string;
	    modal: string;
	    buttons: string;
	    borderModal: string;
	    border: string;
	    progress: string;
	    playButton: string;
	    buttonPrimary: string;
	    error: string;
	    success: string;
	    tag: string;
	    warning: string;
	
	    static createFrom(source: any = {}) {
	        return new ThemeColors(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sidebar = source["sidebar"];
	        this.modal = source["modal"];
	        this.buttons = source["buttons"];
	        this.borderModal = source["borderModal"];
	        this.border = source["border"];
	        this.progress = source["progress"];
	        this.playButton = source["playButton"];
	        this.buttonPrimary = source["buttonPrimary"];
	        this.error = source["error"];
	        this.success = source["success"];
	        this.tag = source["tag"];
	        this.warning = source["warning"];
	    }
	}
	export class Personalization {
	    uiScale: number;
	    background: BackgroundConfig;
	    fontPrimary: string;
	    fontSecondary: string;
	    fontPrimaryColor: string;
	    fontSecondaryColor: string;
	    fontPrimarySize: number;
	    fontSecondarySize: number;
	    colors: ThemeColors;
	    recentColors: string[];
	    animations: boolean;
	    blur: boolean;
	    shadows: boolean;
	    textShadow: boolean;
	    textShadowIntensity: number;
	
	    static createFrom(source: any = {}) {
	        return new Personalization(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.uiScale = source["uiScale"];
	        this.background = this.convertValues(source["background"], BackgroundConfig);
	        this.fontPrimary = source["fontPrimary"];
	        this.fontSecondary = source["fontSecondary"];
	        this.fontPrimaryColor = source["fontPrimaryColor"];
	        this.fontSecondaryColor = source["fontSecondaryColor"];
	        this.fontPrimarySize = source["fontPrimarySize"];
	        this.fontSecondarySize = source["fontSecondarySize"];
	        this.colors = this.convertValues(source["colors"], ThemeColors);
	        this.recentColors = source["recentColors"];
	        this.animations = source["animations"];
	        this.blur = source["blur"];
	        this.shadows = source["shadows"];
	        this.textShadow = source["textShadow"];
	        this.textShadowIntensity = source["textShadowIntensity"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class LauncherConfig {
	    maxRamGB: number;
	    maxMbps: number;
	    concurrentDownloads: number;
	    hideLauncherOnLaunch: boolean;
	    verifyIntegrity?: boolean;
	    checkForUpdatesOnStart: boolean;
	    launchAfterInstall: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LauncherConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.maxRamGB = source["maxRamGB"];
	        this.maxMbps = source["maxMbps"];
	        this.concurrentDownloads = source["concurrentDownloads"];
	        this.hideLauncherOnLaunch = source["hideLauncherOnLaunch"];
	        this.verifyIntegrity = source["verifyIntegrity"];
	        this.checkForUpdatesOnStart = source["checkForUpdatesOnStart"];
	        this.launchAfterInstall = source["launchAfterInstall"];
	    }
	}
	export class MinecraftConfig {
	    hardwareEnabled: boolean;
	    hardwareAcceleration: boolean;
	    gpuType: string;
	    gpuPreset: string;
	    javaMode: string;
	    javaCustomPath: string;
	    proxyEnabled: boolean;
	    proxyHost: string;
	    proxyPort: number;
	    proxyUser: string;
	    proxyPass: string;
	    authVerify: boolean;
	    windowWidth: number;
	    windowHeight: number;
	    fullscreen: boolean;
	    javaArgs: string;
	    gameArgs: string;
	    offlineMode: boolean;
	    compatMode: boolean;
	    detailedLogs: boolean;
	
	    static createFrom(source: any = {}) {
	        return new MinecraftConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hardwareEnabled = source["hardwareEnabled"];
	        this.hardwareAcceleration = source["hardwareAcceleration"];
	        this.gpuType = source["gpuType"];
	        this.gpuPreset = source["gpuPreset"];
	        this.javaMode = source["javaMode"];
	        this.javaCustomPath = source["javaCustomPath"];
	        this.proxyEnabled = source["proxyEnabled"];
	        this.proxyHost = source["proxyHost"];
	        this.proxyPort = source["proxyPort"];
	        this.proxyUser = source["proxyUser"];
	        this.proxyPass = source["proxyPass"];
	        this.authVerify = source["authVerify"];
	        this.windowWidth = source["windowWidth"];
	        this.windowHeight = source["windowHeight"];
	        this.fullscreen = source["fullscreen"];
	        this.javaArgs = source["javaArgs"];
	        this.gameArgs = source["gameArgs"];
	        this.offlineMode = source["offlineMode"];
	        this.compatMode = source["compatMode"];
	        this.detailedLogs = source["detailedLogs"];
	    }
	}
	export class Config {
	    minecraftConfig: MinecraftConfig;
	    launcher: LauncherConfig;
	    personalization: Personalization;
	    idle: IdleConfig;
	    richPresence: RichPresenceConfig;
	    extraData: ExtraData;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.minecraftConfig = this.convertValues(source["minecraftConfig"], MinecraftConfig);
	        this.launcher = this.convertValues(source["launcher"], LauncherConfig);
	        this.personalization = this.convertValues(source["personalization"], Personalization);
	        this.idle = this.convertValues(source["idle"], IdleConfig);
	        this.richPresence = this.convertValues(source["richPresence"], RichPresenceConfig);
	        this.extraData = this.convertValues(source["extraData"], ExtraData);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	
	
	

}

export namespace Handlers {
	
	export class ScreenshotInfo {
	    name: string;
	    path: string;
	    size: number;
	    time?: string;
	
	    static createFrom(source: any = {}) {
	        return new ScreenshotInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.size = source["size"];
	        this.time = source["time"];
	    }
	}

}

export namespace accounts {
	
	export class AccountCredentials {
	    username: string;
	    uuid: string;
	    accessToken: string;
	    userType?: string;
	
	    static createFrom(source: any = {}) {
	        return new AccountCredentials(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.username = source["username"];
	        this.uuid = source["uuid"];
	        this.accessToken = source["accessToken"];
	        this.userType = source["userType"];
	    }
	}
	export class AccountInfo {
	    id: string;
	    name: string;
	    type: string;
	    username: string;
	    uuid: string;
	    authServerUrl?: string;
	    serverName?: string;
	    hasToken: boolean;
	    sessionValid: boolean;
	    createdAt: string;
	    lastUsed?: string;
	    customProperties?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new AccountInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.username = source["username"];
	        this.uuid = source["uuid"];
	        this.authServerUrl = source["authServerUrl"];
	        this.serverName = source["serverName"];
	        this.hasToken = source["hasToken"];
	        this.sessionValid = source["sessionValid"];
	        this.createdAt = source["createdAt"];
	        this.lastUsed = source["lastUsed"];
	        this.customProperties = source["customProperties"];
	    }
	}
	export class AuthlibLoginReq {
	    authServerUrl: string;
	    username: string;
	    password: string;
	
	    static createFrom(source: any = {}) {
	        return new AuthlibLoginReq(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.authServerUrl = source["authServerUrl"];
	        this.username = source["username"];
	        this.password = source["password"];
	    }
	}
	export class CreateAccountReq {
	    type: string;
	    name?: string;
	    username: string;
	    accessToken?: string;
	    authServerUrl?: string;
	    uuid?: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateAccountReq(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.name = source["name"];
	        this.username = source["username"];
	        this.accessToken = source["accessToken"];
	        this.authServerUrl = source["authServerUrl"];
	        this.uuid = source["uuid"];
	    }
	}

}

export namespace assets {
	
	export class FontSlot {
	    type: string;
	    name: string;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new FontSlot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.name = source["name"];
	        this.path = source["path"];
	    }
	}
	export class Assets {
	    fonts: FontSlot[];
	
	    static createFrom(source: any = {}) {
	        return new Assets(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fonts = this.convertValues(source["fonts"], FontSlot);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace cache {
	
	export class Info {
	    cacheDir: string;
	    totalEntries: number;
	    categories: Record<string, number>;
	    ttls: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new Info(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cacheDir = source["cacheDir"];
	        this.totalEntries = source["totalEntries"];
	        this.categories = source["categories"];
	        this.ttls = source["ttls"];
	    }
	}

}

export namespace downloader {
	
	export class DownloadFilter {
	    version: string;
	    client: boolean;
	    libraries: boolean;
	    natives: boolean;
	    assets: boolean;
	    java: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DownloadFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.client = source["client"];
	        this.libraries = source["libraries"];
	        this.natives = source["natives"];
	        this.assets = source["assets"];
	        this.java = source["java"];
	    }
	}
	export class FileProgress {
	    name: string;
	    section: string;
	    size: number;
	    downloaded: number;
	    percent: number;
	    state: string;
	
	    static createFrom(source: any = {}) {
	        return new FileProgress(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.section = source["section"];
	        this.size = source["size"];
	        this.downloaded = source["downloaded"];
	        this.percent = source["percent"];
	        this.state = source["state"];
	    }
	}
	export class SectionProgress {
	    name: string;
	    totalFiles: number;
	    doneFiles: number;
	    mbTotal: number;
	    mbDownloaded: number;
	
	    static createFrom(source: any = {}) {
	        return new SectionProgress(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.totalFiles = source["totalFiles"];
	        this.doneFiles = source["doneFiles"];
	        this.mbTotal = source["mbTotal"];
	        this.mbDownloaded = source["mbDownloaded"];
	    }
	}
	export class DownloadProgress {
	    mbDownloaded: number;
	    mbTotal: number;
	    percent: number;
	    state: string;
	    currentSection: string;
	    sectionsCompleted: string[];
	    sectionsTotal: number;
	    filesDownloaded: number;
	    filesTotal: number;
	    filesExisting: number;
	    currentFile: string;
	    currentUrl: string;
	    currentDest: string;
	    currentProgress: number;
	    log: string[];
	    sections: SectionProgress[];
	    activeFiles: FileProgress[];
	    queuedCount: number;
	    queuedPreview: string[];
	    speedMbps: number;
	
	    static createFrom(source: any = {}) {
	        return new DownloadProgress(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mbDownloaded = source["mbDownloaded"];
	        this.mbTotal = source["mbTotal"];
	        this.percent = source["percent"];
	        this.state = source["state"];
	        this.currentSection = source["currentSection"];
	        this.sectionsCompleted = source["sectionsCompleted"];
	        this.sectionsTotal = source["sectionsTotal"];
	        this.filesDownloaded = source["filesDownloaded"];
	        this.filesTotal = source["filesTotal"];
	        this.filesExisting = source["filesExisting"];
	        this.currentFile = source["currentFile"];
	        this.currentUrl = source["currentUrl"];
	        this.currentDest = source["currentDest"];
	        this.currentProgress = source["currentProgress"];
	        this.log = source["log"];
	        this.sections = this.convertValues(source["sections"], SectionProgress);
	        this.activeFiles = this.convertValues(source["activeFiles"], FileProgress);
	        this.queuedCount = source["queuedCount"];
	        this.queuedPreview = source["queuedPreview"];
	        this.speedMbps = source["speedMbps"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class LatestInfo {
	    release: string;
	    snapshot: string;
	
	    static createFrom(source: any = {}) {
	        return new LatestInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.release = source["release"];
	        this.snapshot = source["snapshot"];
	    }
	}
	export class ManifestVersion {
	    id: string;
	    url: string;
	    type: string;
	    releaseTime: string;
	
	    static createFrom(source: any = {}) {
	        return new ManifestVersion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.url = source["url"];
	        this.type = source["type"];
	        this.releaseTime = source["releaseTime"];
	    }
	}
	export class Manifest {
	    versions: ManifestVersion[];
	    latest: LatestInfo;
	
	    static createFrom(source: any = {}) {
	        return new Manifest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.versions = this.convertValues(source["versions"], ManifestVersion);
	        this.latest = this.convertValues(source["latest"], LatestInfo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	

}

export namespace engine {
	
	export class DownloadInfo {
	    id: string;
	    version: string;
	    state: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new DownloadInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.version = source["version"];
	        this.state = source["state"];
	        this.error = source["error"];
	    }
	}
	export class EngineInfo {
	    name: string;
	    version: string;
	    author: string;
	    goVersion: string;
	    os: string;
	    arch: string;
	    numCpu: number;
	    numGoroutine: number;
	    launcherName?: string;
	    launcherVersion?: string;
	
	    static createFrom(source: any = {}) {
	        return new EngineInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	        this.author = source["author"];
	        this.goVersion = source["goVersion"];
	        this.os = source["os"];
	        this.arch = source["arch"];
	        this.numCpu = source["numCpu"];
	        this.numGoroutine = source["numGoroutine"];
	        this.launcherName = source["launcherName"];
	        this.launcherVersion = source["launcherVersion"];
	    }
	}
	export class GameResp {
	    id: string;
	    pid: number;
	    version: string;
	    status: string;
	    startTime?: string;
	    exitCode: number;
	    logPath?: string;
	    crashLog?: string;
	    crashReason?: string;
	    crashCategory?: string;
	
	    static createFrom(source: any = {}) {
	        return new GameResp(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.pid = source["pid"];
	        this.version = source["version"];
	        this.status = source["status"];
	        this.startTime = source["startTime"];
	        this.exitCode = source["exitCode"];
	        this.logPath = source["logPath"];
	        this.crashLog = source["crashLog"];
	        this.crashReason = source["crashReason"];
	        this.crashCategory = source["crashCategory"];
	    }
	}
	export class InstalledVersion {
	    id: string;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new InstalledVersion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	    }
	}
	export class ModLoaderInstallResult {
	    sessionId: string;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new ModLoaderInstallResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	        this.status = source["status"];
	    }
	}
	export class VersionInfo {
	    id: string;
	    type: string;
	    url: string;
	    time: string;
	
	    static createFrom(source: any = {}) {
	        return new VersionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.url = source["url"];
	        this.time = source["time"];
	    }
	}
	export class VersionStats {
	    version: string;
	    playCount: number;
	    lastPlayed: number;
	    totalPlayed: number;
	
	    static createFrom(source: any = {}) {
	        return new VersionStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.playCount = source["playCount"];
	        this.lastPlayed = source["lastPlayed"];
	        this.totalPlayed = source["totalPlayed"];
	    }
	}

}

export namespace engineconfig {
	
	export class Config {
	    maxCores?: number;
	    maxRam?: number;
	    cacheDir?: string;
	    logDir?: string;
	    workDir?: string;
	    launcherName?: string;
	    launcherVersion?: string;
	    instancesDir?: string;
	    sharedDir?: string;
	    maxMbps?: number;
	    minMbps?: number;
	    cacheTTLManifest?: string;
	    cacheTTLAssets?: string;
	    cacheTTLVersions?: string;
	    cacheTTLModloader?: string;
	    cacheTTLJava?: string;
	    cacheTTLDefault?: string;
	    hardwareEnabled: boolean;
	    hardwareAcceleration: boolean;
	    gpuType: string;
	    gpuPreset: string;
	    javaMode: string;
	    javaCustomPath: string;
	    proxyEnabled: boolean;
	    proxyHost: string;
	    proxyPort: number;
	    proxyUser: string;
	    proxyPass: string;
	    authVerify: boolean;
	    windowWidth: number;
	    windowHeight: number;
	    fullscreen: boolean;
	    javaArgs: string;
	    gameArgs: string;
	    offlineMode: boolean;
	    compatMode: boolean;
	    detailedLogs: boolean;
	    concurrentDownloads: number;
	    verifyIntegrity: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.maxCores = source["maxCores"];
	        this.maxRam = source["maxRam"];
	        this.cacheDir = source["cacheDir"];
	        this.logDir = source["logDir"];
	        this.workDir = source["workDir"];
	        this.launcherName = source["launcherName"];
	        this.launcherVersion = source["launcherVersion"];
	        this.instancesDir = source["instancesDir"];
	        this.sharedDir = source["sharedDir"];
	        this.maxMbps = source["maxMbps"];
	        this.minMbps = source["minMbps"];
	        this.cacheTTLManifest = source["cacheTTLManifest"];
	        this.cacheTTLAssets = source["cacheTTLAssets"];
	        this.cacheTTLVersions = source["cacheTTLVersions"];
	        this.cacheTTLModloader = source["cacheTTLModloader"];
	        this.cacheTTLJava = source["cacheTTLJava"];
	        this.cacheTTLDefault = source["cacheTTLDefault"];
	        this.hardwareEnabled = source["hardwareEnabled"];
	        this.hardwareAcceleration = source["hardwareAcceleration"];
	        this.gpuType = source["gpuType"];
	        this.gpuPreset = source["gpuPreset"];
	        this.javaMode = source["javaMode"];
	        this.javaCustomPath = source["javaCustomPath"];
	        this.proxyEnabled = source["proxyEnabled"];
	        this.proxyHost = source["proxyHost"];
	        this.proxyPort = source["proxyPort"];
	        this.proxyUser = source["proxyUser"];
	        this.proxyPass = source["proxyPass"];
	        this.authVerify = source["authVerify"];
	        this.windowWidth = source["windowWidth"];
	        this.windowHeight = source["windowHeight"];
	        this.fullscreen = source["fullscreen"];
	        this.javaArgs = source["javaArgs"];
	        this.gameArgs = source["gameArgs"];
	        this.offlineMode = source["offlineMode"];
	        this.compatMode = source["compatMode"];
	        this.detailedLogs = source["detailedLogs"];
	        this.concurrentDownloads = source["concurrentDownloads"];
	        this.verifyIntegrity = source["verifyIntegrity"];
	    }
	}

}

export namespace history {
	
	export class CrashEntry {
	    id: string;
	    timestamp: number;
	    version: string;
	    exit_code: number;
	    crash_reason?: string;
	    crash_category?: string;
	    launcherLogPath?: string;
	    minecraftLogPath?: string;
	    jvmLogPath?: string;
	
	    static createFrom(source: any = {}) {
	        return new CrashEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = source["timestamp"];
	        this.version = source["version"];
	        this.exit_code = source["exit_code"];
	        this.crash_reason = source["crash_reason"];
	        this.crash_category = source["crash_category"];
	        this.launcherLogPath = source["launcherLogPath"];
	        this.minecraftLogPath = source["minecraftLogPath"];
	        this.jvmLogPath = source["jvmLogPath"];
	    }
	}
	export class Entry {
	    id: string;
	    timestamp: number;
	    instance_id?: string;
	    version: string;
	    player_name: string;
	    play_time_seconds: number;
	    exit_code: number;
	    crash_reason?: string;
	    modloader?: string;
	    max_ram?: number;
	
	    static createFrom(source: any = {}) {
	        return new Entry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = source["timestamp"];
	        this.instance_id = source["instance_id"];
	        this.version = source["version"];
	        this.player_name = source["player_name"];
	        this.play_time_seconds = source["play_time_seconds"];
	        this.exit_code = source["exit_code"];
	        this.crash_reason = source["crash_reason"];
	        this.modloader = source["modloader"];
	        this.max_ram = source["max_ram"];
	    }
	}

}

export namespace instance {
	
	export class AddVersionReq {
	    version: string;
	    client?: boolean;
	    libraries?: boolean;
	    natives?: boolean;
	    assets?: boolean;
	    java?: boolean;
	    maxRetries?: number;
	    concurrency?: number;
	    skipVerify?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AddVersionReq(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.client = source["client"];
	        this.libraries = source["libraries"];
	        this.natives = source["natives"];
	        this.assets = source["assets"];
	        this.java = source["java"];
	        this.maxRetries = source["maxRetries"];
	        this.concurrency = source["concurrency"];
	        this.skipVerify = source["skipVerify"];
	    }
	}
	export class InstanceLaunchConfig {
	    version?: string;
	    javaExec?: string;
	    minRam?: number;
	    maxRam?: number;
	    useOfficialJava?: boolean;
	    fullscreen?: boolean;
	    hardwareAcceleration?: boolean;
	    gcPreset?: string;
	    gpuPreference?: string;
	    customResolution?: boolean;
	    resWidth?: number;
	    resHeight?: number;
	    demoUser?: boolean;
	    windowTitle?: string;
	    userType?: string;
	    advanced?: launcher.AdvancedConfig;
	    gameDir?: string;
	    assetsDir?: string;
	    librariesDir?: string;
	    nativesDir?: string;
	    versionsDir?: string;
	    runtimeDir?: string;
	    cacheDir?: string;
	    assetIndexId?: string;
	    disableAssets?: boolean;
	    serverAddress?: string;
	    serverPort?: number;
	    maxMetaspaceSize?: number;
	    stackSize?: number;
	    allowServerList?: boolean;
	    allowMultiplayer?: boolean;
	    allowChat?: boolean;
	    allowRealms?: boolean;
	    framerateLimit?: number;
	    logKeepDays?: number;
	    logMaxFiles?: number;
	    preLaunchCommand?: string;
	    postLaunchCommand?: string;
	    environmentVars?: Record<string, string>;
	    javaArgs?: string[];
	    gameArgs?: string[];
	    skipLibraryCheck?: boolean;
	    skipAssetCheck?: boolean;
	    skipNativeExtract?: boolean;
	    skipVersionDownload?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new InstanceLaunchConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.javaExec = source["javaExec"];
	        this.minRam = source["minRam"];
	        this.maxRam = source["maxRam"];
	        this.useOfficialJava = source["useOfficialJava"];
	        this.fullscreen = source["fullscreen"];
	        this.hardwareAcceleration = source["hardwareAcceleration"];
	        this.gcPreset = source["gcPreset"];
	        this.gpuPreference = source["gpuPreference"];
	        this.customResolution = source["customResolution"];
	        this.resWidth = source["resWidth"];
	        this.resHeight = source["resHeight"];
	        this.demoUser = source["demoUser"];
	        this.windowTitle = source["windowTitle"];
	        this.userType = source["userType"];
	        this.advanced = this.convertValues(source["advanced"], launcher.AdvancedConfig);
	        this.gameDir = source["gameDir"];
	        this.assetsDir = source["assetsDir"];
	        this.librariesDir = source["librariesDir"];
	        this.nativesDir = source["nativesDir"];
	        this.versionsDir = source["versionsDir"];
	        this.runtimeDir = source["runtimeDir"];
	        this.cacheDir = source["cacheDir"];
	        this.assetIndexId = source["assetIndexId"];
	        this.disableAssets = source["disableAssets"];
	        this.serverAddress = source["serverAddress"];
	        this.serverPort = source["serverPort"];
	        this.maxMetaspaceSize = source["maxMetaspaceSize"];
	        this.stackSize = source["stackSize"];
	        this.allowServerList = source["allowServerList"];
	        this.allowMultiplayer = source["allowMultiplayer"];
	        this.allowChat = source["allowChat"];
	        this.allowRealms = source["allowRealms"];
	        this.framerateLimit = source["framerateLimit"];
	        this.logKeepDays = source["logKeepDays"];
	        this.logMaxFiles = source["logMaxFiles"];
	        this.preLaunchCommand = source["preLaunchCommand"];
	        this.postLaunchCommand = source["postLaunchCommand"];
	        this.environmentVars = source["environmentVars"];
	        this.javaArgs = source["javaArgs"];
	        this.gameArgs = source["gameArgs"];
	        this.skipLibraryCheck = source["skipLibraryCheck"];
	        this.skipAssetCheck = source["skipAssetCheck"];
	        this.skipNativeExtract = source["skipNativeExtract"];
	        this.skipVersionDownload = source["skipVersionDownload"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CreateInstanceReq {
	    name: string;
	    version?: string;
	    title?: string;
	    description?: string;
	    icon?: string;
	    banner?: string;
	    background?: string;
	    accentColor?: string;
	    group?: string;
	    tags?: string[];
	    favorite?: boolean;
	    launchConfig?: InstanceLaunchConfig;
	
	    static createFrom(source: any = {}) {
	        return new CreateInstanceReq(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.icon = source["icon"];
	        this.banner = source["banner"];
	        this.background = source["background"];
	        this.accentColor = source["accentColor"];
	        this.group = source["group"];
	        this.tags = source["tags"];
	        this.favorite = source["favorite"];
	        this.launchConfig = this.convertValues(source["launchConfig"], InstanceLaunchConfig);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class InstanceInfo {
	    name: string;
	    title: string;
	    versions: string[];
	    favorite: boolean;
	    group: string;
	    lastPlayed: string;
	    playTime: number;
	
	    static createFrom(source: any = {}) {
	        return new InstanceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.title = source["title"];
	        this.versions = source["versions"];
	        this.favorite = source["favorite"];
	        this.group = source["group"];
	        this.lastPlayed = source["lastPlayed"];
	        this.playTime = source["playTime"];
	    }
	}
	
	export class InstanceLaunchResult {
	    id: string;
	    pid: number;
	    version: string;
	    status: string;
	    logPath: string;
	
	    static createFrom(source: any = {}) {
	        return new InstanceLaunchResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.pid = source["pid"];
	        this.version = source["version"];
	        this.status = source["status"];
	        this.logPath = source["logPath"];
	    }
	}
	export class InstanceMetadata {
	    id: string;
	    name: string;
	    title: string;
	    description: string;
	    icon: string;
	    banner: string;
	    background: string;
	    accentColor: string;
	    group: string;
	    tags: string[];
	    favorite: boolean;
	    createdAt: string;
	    lastPlayed: string;
	    playTime: number;
	    versions: string[];
	    configPath: string;
	
	    static createFrom(source: any = {}) {
	        return new InstanceMetadata(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.icon = source["icon"];
	        this.banner = source["banner"];
	        this.background = source["background"];
	        this.accentColor = source["accentColor"];
	        this.group = source["group"];
	        this.tags = source["tags"];
	        this.favorite = source["favorite"];
	        this.createdAt = source["createdAt"];
	        this.lastPlayed = source["lastPlayed"];
	        this.playTime = source["playTime"];
	        this.versions = source["versions"];
	        this.configPath = source["configPath"];
	    }
	}
	export class UpdateMetadataReq {
	    title?: string;
	    description?: string;
	    icon?: string;
	    banner?: string;
	    background?: string;
	    accentColor?: string;
	    group?: string;
	    tags?: string[];
	    favorite?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UpdateMetadataReq(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.description = source["description"];
	        this.icon = source["icon"];
	        this.banner = source["banner"];
	        this.background = source["background"];
	        this.accentColor = source["accentColor"];
	        this.group = source["group"];
	        this.tags = source["tags"];
	        this.favorite = source["favorite"];
	    }
	}
	export class VerifyIssue {
	    type: string;
	    file: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new VerifyIssue(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.file = source["file"];
	        this.message = source["message"];
	    }
	}
	export class VerifyResult {
	    valid: boolean;
	    version: string;
	    issues: VerifyIssue[];
	
	    static createFrom(source: any = {}) {
	        return new VerifyResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.valid = source["valid"];
	        this.version = source["version"];
	        this.issues = this.convertValues(source["issues"], VerifyIssue);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace launcher {
	
	export class AuthLibConfig {
	    enabled: boolean;
	    injectorPath?: string;
	    authServerUrl?: string;
	    preVerifyServer: boolean;
	    preVerifyTimeout?: number;
	    skipServerCheck: boolean;
	    username?: string;
	    serverToken?: string;
	
	    static createFrom(source: any = {}) {
	        return new AuthLibConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.injectorPath = source["injectorPath"];
	        this.authServerUrl = source["authServerUrl"];
	        this.preVerifyServer = source["preVerifyServer"];
	        this.preVerifyTimeout = source["preVerifyTimeout"];
	        this.skipServerCheck = source["skipServerCheck"];
	        this.username = source["username"];
	        this.serverToken = source["serverToken"];
	    }
	}
	export class AdvancedConfig {
	    authlib?: AuthLibConfig;
	    javaExec?: string;
	    useOfficialJava: boolean;
	    useSystemJava: boolean;
	    minRam: number;
	    maxRam: number;
	    gcPreset?: string;
	    gpuPreference?: string;
	    hardwareAcceleration?: boolean;
	    fullscreen: boolean;
	    customResolution: boolean;
	    resWidth: number;
	    resHeight: number;
	    windowTitle?: string;
	    demoUser: boolean;
	    userType?: string;
	    logLevel?: string;
	    runtimeDir?: string;
	    gameDir?: string;
	    assetsDir?: string;
	    librariesDir?: string;
	    versionsDir?: string;
	    nativesBaseDir?: string;
	    nativesDir?: string;
	    cacheDir?: string;
	    workingDir?: string;
	    assetIndexId?: string;
	    assetIndexType?: string;
	    assetIndexUrl?: string;
	    assetIndexVirtual?: boolean;
	    disableAssets: boolean;
	    forceReindex: boolean;
	    serverAddress?: string;
	    serverPort: number;
	    quickPlayPath?: string;
	    minecraftLogConfig?: string;
	    maxMetaspaceSize?: number;
	    stackSize?: number;
	    javaArgs?: string[];
	    gameArgs?: string[];
	    disableLibraries: boolean;
	    skipLibraryCheck: boolean;
	    skipAssetCheck: boolean;
	    skipNativeExtract: boolean;
	    skipVersionDownload: boolean;
	    profileVersion?: string;
	    baseVersion?: string;
	    environmentVars?: Record<string, string>;
	    jvmFlags?: Record<string, boolean>;
	    downloadConcurrency?: number;
	    maxRetries?: number;
	    connectionTimeout?: number;
	    proxyHost?: string;
	    proxyPort?: number;
	    proxyUser?: string;
	    proxyPass?: string;
	    libraryCustomRepo?: string;
	    assetCustomUrl?: string;
	    forceRedownload: boolean;
	    javaModulePath?: string;
	    javaAddModules?: string[];
	    javaAddExports?: string[];
	    javaAddOpens?: string[];
	    directMemorySize?: number;
	    reservedCodeCache?: number;
	    metaspaceSize?: number;
	    allowServerList?: boolean;
	    allowMultiplayer?: boolean;
	    allowChat?: boolean;
	    allowRealms?: boolean;
	    framerateLimit?: number;
	    renderer?: string;
	    preLaunchCommand?: string;
	    postLaunchCommand?: string;
	    logKeepDays?: number;
	    logMaxFiles?: number;
	    gameLogLines?: number;
	    libraryExcludePatterns?: string[];
	
	    static createFrom(source: any = {}) {
	        return new AdvancedConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.authlib = this.convertValues(source["authlib"], AuthLibConfig);
	        this.javaExec = source["javaExec"];
	        this.useOfficialJava = source["useOfficialJava"];
	        this.useSystemJava = source["useSystemJava"];
	        this.minRam = source["minRam"];
	        this.maxRam = source["maxRam"];
	        this.gcPreset = source["gcPreset"];
	        this.gpuPreference = source["gpuPreference"];
	        this.hardwareAcceleration = source["hardwareAcceleration"];
	        this.fullscreen = source["fullscreen"];
	        this.customResolution = source["customResolution"];
	        this.resWidth = source["resWidth"];
	        this.resHeight = source["resHeight"];
	        this.windowTitle = source["windowTitle"];
	        this.demoUser = source["demoUser"];
	        this.userType = source["userType"];
	        this.logLevel = source["logLevel"];
	        this.runtimeDir = source["runtimeDir"];
	        this.gameDir = source["gameDir"];
	        this.assetsDir = source["assetsDir"];
	        this.librariesDir = source["librariesDir"];
	        this.versionsDir = source["versionsDir"];
	        this.nativesBaseDir = source["nativesBaseDir"];
	        this.nativesDir = source["nativesDir"];
	        this.cacheDir = source["cacheDir"];
	        this.workingDir = source["workingDir"];
	        this.assetIndexId = source["assetIndexId"];
	        this.assetIndexType = source["assetIndexType"];
	        this.assetIndexUrl = source["assetIndexUrl"];
	        this.assetIndexVirtual = source["assetIndexVirtual"];
	        this.disableAssets = source["disableAssets"];
	        this.forceReindex = source["forceReindex"];
	        this.serverAddress = source["serverAddress"];
	        this.serverPort = source["serverPort"];
	        this.quickPlayPath = source["quickPlayPath"];
	        this.minecraftLogConfig = source["minecraftLogConfig"];
	        this.maxMetaspaceSize = source["maxMetaspaceSize"];
	        this.stackSize = source["stackSize"];
	        this.javaArgs = source["javaArgs"];
	        this.gameArgs = source["gameArgs"];
	        this.disableLibraries = source["disableLibraries"];
	        this.skipLibraryCheck = source["skipLibraryCheck"];
	        this.skipAssetCheck = source["skipAssetCheck"];
	        this.skipNativeExtract = source["skipNativeExtract"];
	        this.skipVersionDownload = source["skipVersionDownload"];
	        this.profileVersion = source["profileVersion"];
	        this.baseVersion = source["baseVersion"];
	        this.environmentVars = source["environmentVars"];
	        this.jvmFlags = source["jvmFlags"];
	        this.downloadConcurrency = source["downloadConcurrency"];
	        this.maxRetries = source["maxRetries"];
	        this.connectionTimeout = source["connectionTimeout"];
	        this.proxyHost = source["proxyHost"];
	        this.proxyPort = source["proxyPort"];
	        this.proxyUser = source["proxyUser"];
	        this.proxyPass = source["proxyPass"];
	        this.libraryCustomRepo = source["libraryCustomRepo"];
	        this.assetCustomUrl = source["assetCustomUrl"];
	        this.forceRedownload = source["forceRedownload"];
	        this.javaModulePath = source["javaModulePath"];
	        this.javaAddModules = source["javaAddModules"];
	        this.javaAddExports = source["javaAddExports"];
	        this.javaAddOpens = source["javaAddOpens"];
	        this.directMemorySize = source["directMemorySize"];
	        this.reservedCodeCache = source["reservedCodeCache"];
	        this.metaspaceSize = source["metaspaceSize"];
	        this.allowServerList = source["allowServerList"];
	        this.allowMultiplayer = source["allowMultiplayer"];
	        this.allowChat = source["allowChat"];
	        this.allowRealms = source["allowRealms"];
	        this.framerateLimit = source["framerateLimit"];
	        this.renderer = source["renderer"];
	        this.preLaunchCommand = source["preLaunchCommand"];
	        this.postLaunchCommand = source["postLaunchCommand"];
	        this.logKeepDays = source["logKeepDays"];
	        this.logMaxFiles = source["logMaxFiles"];
	        this.gameLogLines = source["gameLogLines"];
	        this.libraryExcludePatterns = source["libraryExcludePatterns"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class LaunchConfig {
	    Version: string;
	    Username: string;
	    InstanceID: string;
	    UUID: string;
	    AccessToken: string;
	    XUID: string;
	    ClientID: string;
	    LauncherName: string;
	    LauncherVersion: string;
	    LogDir: string;
	    Advanced?: AdvancedConfig;
	    Profile: string;
	
	    static createFrom(source: any = {}) {
	        return new LaunchConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Version = source["Version"];
	        this.Username = source["Username"];
	        this.InstanceID = source["InstanceID"];
	        this.UUID = source["UUID"];
	        this.AccessToken = source["AccessToken"];
	        this.XUID = source["XUID"];
	        this.ClientID = source["ClientID"];
	        this.LauncherName = source["LauncherName"];
	        this.LauncherVersion = source["LauncherVersion"];
	        this.LogDir = source["LogDir"];
	        this.Advanced = this.convertValues(source["Advanced"], AdvancedConfig);
	        this.Profile = source["Profile"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace main {
	
	export class AddInstanceVersionResult {
	    downloadId: string;
	    version: string;
	
	    static createFrom(source: any = {}) {
	        return new AddInstanceVersionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.downloadId = source["downloadId"];
	        this.version = source["version"];
	    }
	}
	export class CreateInstanceResult {
	    metadata?: instance.InstanceMetadata;
	    downloadId: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateInstanceResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.metadata = this.convertValues(source["metadata"], instance.InstanceMetadata);
	        this.downloadId = source["downloadId"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class GetInstanceResult {
	    metadata?: instance.InstanceMetadata;
	    config?: instance.InstanceLaunchConfig;
	
	    static createFrom(source: any = {}) {
	        return new GetInstanceResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.metadata = this.convertValues(source["metadata"], instance.InstanceMetadata);
	        this.config = this.convertValues(source["config"], instance.InstanceLaunchConfig);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class HistoryStatsResult {
	    stats: engine.VersionStats[];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new HistoryStatsResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.stats = this.convertValues(source["stats"], engine.VersionStats);
	        this.total = source["total"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class InstanceDownloadStatusResult {
	    id: string;
	    version: string;
	    state: string;
	
	    static createFrom(source: any = {}) {
	        return new InstanceDownloadStatusResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.version = source["version"];
	        this.state = source["state"];
	    }
	}
	export class RecommendedRAMResult {
	    minRam: number;
	    maxRam: number;
	    gcPreset: string;
	
	    static createFrom(source: any = {}) {
	        return new RecommendedRAMResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.minRam = source["minRam"];
	        this.maxRam = source["maxRam"];
	        this.gcPreset = source["gcPreset"];
	    }
	}

}

export namespace modloader {
	
	export class ExecutionPlan {
	    mainClass: string;
	    additionalClasspath: string[];
	    additionalJvmArgs: string[];
	    additionalGameArgs: string[];
	    useModulePath: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ExecutionPlan(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mainClass = source["mainClass"];
	        this.additionalClasspath = source["additionalClasspath"];
	        this.additionalJvmArgs = source["additionalJvmArgs"];
	        this.additionalGameArgs = source["additionalGameArgs"];
	        this.useModulePath = source["useModulePath"];
	    }
	}
	export class InstalledLoader {
	    loaderType: string;
	    loaderVersion: string;
	    minecraftVersion: string;
	    versionJsonId: string;
	    installerJarPath?: string;
	    installedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new InstalledLoader(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.loaderType = source["loaderType"];
	        this.loaderVersion = source["loaderVersion"];
	        this.minecraftVersion = source["minecraftVersion"];
	        this.versionJsonId = source["versionJsonId"];
	        this.installerJarPath = source["installerJarPath"];
	        this.installedAt = source["installedAt"];
	    }
	}
	export class LoaderVersion {
	    loaderVersion: string;
	    minecraftVersion: string;
	    stable: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LoaderVersion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.loaderVersion = source["loaderVersion"];
	        this.minecraftVersion = source["minecraftVersion"];
	        this.stable = source["stable"];
	    }
	}

}

export namespace profile {
	
	export class Profile {
	    name: string;
	    version?: string;
	    lastVersionId?: string;
	    gameDir?: string;
	    javaExec?: string;
	    javaArgs?: string;
	    resWidth?: number;
	    resHeight?: number;
	    fullscreen: boolean;
	    modLoader?: string;
	    modLoaderVersion?: string;
	    icon?: string;
	    createdAt: string;
	    lastUsed?: string;
	    customProperties?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new Profile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	        this.lastVersionId = source["lastVersionId"];
	        this.gameDir = source["gameDir"];
	        this.javaExec = source["javaExec"];
	        this.javaArgs = source["javaArgs"];
	        this.resWidth = source["resWidth"];
	        this.resHeight = source["resHeight"];
	        this.fullscreen = source["fullscreen"];
	        this.modLoader = source["modLoader"];
	        this.modLoaderVersion = source["modLoaderVersion"];
	        this.icon = source["icon"];
	        this.createdAt = source["createdAt"];
	        this.lastUsed = source["lastUsed"];
	        this.customProperties = source["customProperties"];
	    }
	}

}

