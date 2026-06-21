package dev.novastep.core.server.request;

import dev.novastep.core.util.SystemResources;
import java.nio.file.Path;
import java.util.List;
import java.util.Map;

public class LaunchRequest {

    public String version;
    public String instancePath;
    public String sharedPath;
    public String javaPath;
    public String configPath;

    public Boolean hardwareAcceleration;
    public Boolean disableHardwareAcceleration;
    public String gcPreset;
    public String gpuPreference;

    public LaunchFeatures features;
    public AuthConfig auth;
    public AuthlibInjector authlibInjector;
    public JvmConfig jvm;
    public WindowConfig window;
    public LauncherBranding launcher;
    public GameCustomization game;

    public static class AuthConfig {
        public String username;
        public String uuid;
        public String accessToken;
        public String userType;
        public String clientId;
        public String xuid;
    }

    public static class AuthlibInjector {
        public Boolean enabled;
        public String jarPath;
        public String serverUrl;
    }

    public static class JvmConfig {
        public Integer minMemoryMb;
        public Integer maxMemoryMb;
        public List<String> extraArgs;
        public List<String> prependArgs;
    }

    public static class WindowConfig {
        public Integer width;
        public Integer height;
        public Boolean fullscreen;
    }

    public static class LaunchFeatures {
        public Boolean demo;
        public QuickPlay quickPlay;

        public boolean isDemoMode() {
            return Boolean.TRUE.equals(demo);
        }

        public boolean hasQuickPlay() {
            return quickPlay != null && quickPlay.mode != null;
        }

        public static class QuickPlay {
            public String mode;
            public String value;
        }
    }

    public static class LauncherBranding {
        public String name;
        public String version;
    }

    public static class GameCustomization {
        public String gameDir;
        public List<String> extraGameArgs;
        public Map<String, String> extraJvmProperties;
        public String serverHost;
        public Integer serverPort;
    }

    public String resolvedInstancePath() {
        if (instancePath == null || instancePath.isBlank())
            return Path.of(System.getProperty("user.dir")).resolve("instances").resolve("default").toAbsolutePath()
                    .toString();
        return Path.of(instancePath).toAbsolutePath().toString();
    }

    public Path resolvedLibrariesPath() {
        if (sharedPath != null && !sharedPath.isBlank())
            return Path.of(sharedPath).toAbsolutePath().resolve("libraries");
        return Path.of(resolvedInstancePath()).resolve("libraries");
    }

    public Path resolvedAssetsPath() {
        if (sharedPath != null && !sharedPath.isBlank())
            return Path.of(sharedPath).toAbsolutePath().resolve("assets");
        return Path.of(resolvedInstancePath()).resolve("assets");
    }

    public Path resolvedRuntimePath() {
        if (sharedPath != null && !sharedPath.isBlank())
            return Path.of(sharedPath).toAbsolutePath().resolve("runtime");
        return Path.of(resolvedInstancePath()).resolve("runtime");
    }

    public String resolvedLauncherName() {
        return (launcher != null && launcher.name != null && !launcher.name.isBlank())
                ? launcher.name
                : "StepLauncher";
    }

    public String resolvedLauncherVersion() {
        return (launcher != null && launcher.version != null && !launcher.version.isBlank())
                ? launcher.version
                : dev.novastep.core.CoreVersion.get();
    }

    public String resolvedJavaPath() {
        return javaPath != null ? javaPath : "java";
    }

    public int resolvedWidth() {
        return window != null && window.width != null ? window.width : 854;
    }

    public int resolvedHeight() {
        return window != null && window.height != null ? window.height : 480;
    }

    public boolean isFullscreen() {
        return window != null && Boolean.TRUE.equals(window.fullscreen);
    }

    public String resolvedUsername() {
        return auth != null && auth.username != null ? auth.username : "Player";
    }

    public String resolvedUuid() {
        return auth != null && auth.uuid != null ? auth.uuid : "00000000-0000-0000-0000-000000000000";
    }

    public String resolvedAccessToken() {
        return auth != null && auth.accessToken != null ? auth.accessToken : "0";
    }

    public String resolvedUserType() {
        return auth != null && auth.userType != null ? auth.userType : "msa";
    }

    public String resolvedClientId() {
        return auth != null && auth.clientId != null ? auth.clientId : "";
    }

    public String resolvedXuid() {
        return auth != null && auth.xuid != null ? auth.xuid : "";
    }

    public int resolvedMinMemory() {
        int req = jvm != null && jvm.minMemoryMb != null ? jvm.minMemoryMb : 0;
        return SystemResources.safeRam(req, resolvedMaxMemoryRaw())[0];
    }

    public int resolvedMaxMemory() {
        int req = jvm != null && jvm.maxMemoryMb != null ? jvm.maxMemoryMb : 0;
        int min = jvm != null && jvm.minMemoryMb != null ? jvm.minMemoryMb : 0;
        return SystemResources.safeRam(min, req)[1];
    }

    private int resolvedMaxMemoryRaw() {
        return jvm != null && jvm.maxMemoryMb != null ? jvm.maxMemoryMb : 0;
    }

    public boolean isAuthlibEnabled() {
        return authlibInjector != null && Boolean.TRUE.equals(authlibInjector.enabled)
                && authlibInjector.jarPath != null && authlibInjector.serverUrl != null;
    }

    public boolean isHardwareAccelerationDisabled() {
        if (disableHardwareAcceleration != null)
            return Boolean.TRUE.equals(disableHardwareAcceleration);
        if (hardwareAcceleration != null)
            return !Boolean.TRUE.equals(hardwareAcceleration);
        return false;
    }

    public String resolvedGameDir() {
        if (game != null && game.gameDir != null && !game.gameDir.isBlank())
            return Path.of(game.gameDir).toAbsolutePath().toString();
        return Path.of(resolvedInstancePath()).resolve("game").toAbsolutePath().toString();
    }

    public String validate() {
        if (version == null || version.isBlank())
            return "'version' es requerido";
        if (instancePath == null || instancePath.isBlank())
            return "'instancePath' es requerido";
        if (isAuthlibEnabled()) {
            if (!new java.io.File(authlibInjector.jarPath).exists())
                return "authlib-injector.jar no encontrado: " + authlibInjector.jarPath;
        }
        return null;
    }
}
