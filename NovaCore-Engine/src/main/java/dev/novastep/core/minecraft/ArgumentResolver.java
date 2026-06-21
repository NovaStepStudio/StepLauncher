package dev.novastep.core.minecraft;

import dev.novastep.core.minecraft.version.VersionInfo;
import dev.novastep.core.server.request.LaunchRequest;
import dev.novastep.core.json.JacksonCompatibilityAdapter.JsonNode;

import java.io.File;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Set;

public class ArgumentResolver {

    private static final Set<String> QUICKPLAY_ARGS = Set.of(
            "--quickPlaySingleplayer", "--quickPlayMultiplayer",
            "--quickPlayRealms", "--quickPlayPath");

    public static class LaunchContext {
        public final String username;
        public final String version;
        public final String gameDir;
        public final String assetsDir;
        public final String assetIndex;
        public final String uuid;
        public final String accessToken;
        public final String userType;
        public final String clientId;
        public final String xuid;
        public final String versionType;
        public final String nativesDir;
        public final String libraryDir;
        public String classpathStr;
        public final String launcherName;
        public final String launcherVersion;
        public final int width;
        public final int height;

        public LaunchContext(String username, String version, String gameDir, String assetsDir,
                             String assetIndex, String uuid, String accessToken, String userType,
                             String clientId, String xuid, String versionType, String nativesDir,
                             String classpathStr, String launcherName, String launcherVersion,
                             int width, int height, String libraryDir) {
            this.username = username;
            this.version = version;
            this.gameDir = gameDir;
            this.assetsDir = assetsDir;
            this.assetIndex = assetIndex;
            this.uuid = uuid;
            this.accessToken = accessToken;
            this.userType = userType;
            this.clientId = clientId;
            this.xuid = xuid;
            this.versionType = versionType;
            this.nativesDir = nativesDir;
            this.classpathStr = classpathStr;
            this.launcherName = launcherName;
            this.launcherVersion = launcherVersion;
            this.width = width;
            this.height = height;
            this.libraryDir = libraryDir;
        }
    }

    public static ArgumentResolver fromRequest(LaunchRequest req, VersionInfo info,
                                               Path instancePath, String vanillaVersionId) {
        Path librariesPath = req.resolvedLibrariesPath().toAbsolutePath();
        String nativesDir = instancePath.toAbsolutePath()
                .resolve("versions")
                .resolve(vanillaVersionId)
                .resolve("natives")
                .toString();

        LaunchContext context = new LaunchContext(
                req.resolvedUsername(),
                vanillaVersionId,
                req.resolvedGameDir(),
                req.resolvedAssetsPath().toString(),
                info.assetIndex != null ? info.assetIndex.id : "legacy",
                req.resolvedUuid(),
                req.resolvedAccessToken(),
                req.resolvedUserType(),
                req.resolvedClientId(),
                req.resolvedXuid(),
                info.type != null ? info.type : "release",
                nativesDir,
                "",
                req.resolvedLauncherName(),
                req.resolvedLauncherVersion(),
                req.resolvedWidth(),
                req.resolvedHeight(),
                librariesPath + File.separator
        );
        return new ArgumentResolver(info, context).configureFromRequest(req);
    }

    public static ArgumentResolver fromRequest(LaunchRequest req, VersionInfo info, Path instancePath) {
        return fromRequest(req, info, instancePath, info.id);
    }

    public boolean demo = false;
    public boolean quickPlay = false;
    public String quickPlayMode = null;
    public String quickPlayValue = null;

    private final VersionInfo versionInfo;
    private final LaunchContext context;

    public ArgumentResolver(VersionInfo versionInfo, LaunchContext context) {
        this.versionInfo = versionInfo;
        this.context = context;
    }

    public ArgumentResolver configureFromRequest(LaunchRequest req) {
        if (req.features != null) {
            this.demo = req.features.isDemoMode();
            this.quickPlay = req.features.hasQuickPlay();
            if (req.features.quickPlay != null) {
                this.quickPlayMode = req.features.quickPlay.mode;
                this.quickPlayValue = req.features.quickPlay.value;
            }
        }
        return this;
    }

    public List<String> buildJvmArgs(ClasspathBuilder classpathBuilder) {
        context.classpathStr = classpathBuilder.buildClasspathString();
        return resolveJvmArgs();
    }

    public List<String> buildGameArgs() {
        return resolveGameArgs();
    }

    public List<String> resolveJvmArgs() {
        List<String> args = new ArrayList<>();
        if (versionInfo.arguments != null && versionInfo.arguments.jvm != null) {
            processArgList(versionInfo.arguments.jvm, args, false);
        } else {
            args.add("-Djava.library.path=" + context.nativesDir);
            args.add("-Dminecraft.launcher.brand=" + context.launcherName);
            args.add("-Dminecraft.launcher.version=" + context.launcherVersion);
            args.add("-cp");
            args.add(context.classpathStr);
        }
        return args;
    }

    public List<String> resolveGameArgs() {
        List<String> args = new ArrayList<>();
        if (versionInfo.arguments != null && versionInfo.arguments.game != null) {
            processArgList(versionInfo.arguments.game, args, true);
        } else if (versionInfo.minecraftArguments != null) {
            for (String token : versionInfo.minecraftArguments.split(" ")) {
                if (!token.isBlank()) {
                    args.add(substitute(token));
                }
            }
        }

        if (quickPlay && quickPlayMode != null && quickPlayValue != null) {
            boolean alreadyPresent = args.stream().anyMatch(value -> value.startsWith("--quickPlay"));
            if (!alreadyPresent) {
                switch (quickPlayMode) {
                    case "singleplayer" -> {
                        args.add("--quickPlaySingleplayer");
                        args.add(quickPlayValue);
                    }
                    case "multiplayer" -> {
                        args.add("--quickPlayMultiplayer");
                        args.add(quickPlayValue);
                    }
                    case "realms" -> {
                        args.add("--quickPlayRealms");
                        args.add(quickPlayValue);
                    }
                }
            }
        }

        return args;
    }

    @SuppressWarnings("unchecked")
    private void processArgList(List<Object> values, List<String> target, boolean gameArgs) {
        for (Object entry : values) {
            if (entry instanceof String value) {
                if (!gameArgs || !shouldSkipGameArg(value)) {
                    target.add(substitute(value));
                }
                continue;
            }

            if (entry instanceof JsonNode node) {
                processConditionalNode(node, target, gameArgs);
                continue;
            }

            if (entry instanceof Map<?, ?> rawMap) {
                processConditionalMap((Map<String, Object>) rawMap, target, gameArgs);
            }
        }
    }

    private void processConditionalNode(JsonNode node, List<String> target, boolean gameArgs) {
        if (node.isTextual()) {
            String value = node.asText();
            if (!gameArgs || !shouldSkipGameArg(value)) {
                target.add(substitute(value));
            }
            return;
        }
        if (!node.isObject()) {
            return;
        }
        if (node.has("rules") && !RuleEvaluator.evaluateRules(node.get("rules"), featureSet())) {
            return;
        }
        JsonNode value = node.get("value");
        appendNodeValue(value, target, gameArgs);
    }

    @SuppressWarnings("unchecked")
    private void processConditionalMap(Map<String, Object> map, List<String> target, boolean gameArgs) {
        Object rules = map.get("rules");
        if (rules instanceof List<?> rawRules && !RuleEvaluator.evaluateRules((List<Object>) rawRules, featureSet())) {
            return;
        }
        Object value = map.get("value");
        if (value instanceof String single) {
            if (!gameArgs || !shouldSkipGameArg(single)) {
                target.add(substitute(single));
            }
        } else if (value instanceof List<?> list) {
            for (Object item : list) {
                String token = String.valueOf(item);
                if (!gameArgs || !shouldSkipGameArg(token)) {
                    target.add(substitute(token));
                }
            }
        }
    }

    private void appendNodeValue(JsonNode value, List<String> target, boolean gameArgs) {
        if (value == null) {
            return;
        }
        if (value.isTextual()) {
            String token = value.asText();
            if (!gameArgs || !shouldSkipGameArg(token)) {
                target.add(substitute(token));
            }
            return;
        }
        if (value.isArray()) {
            for (JsonNode item : value) {
                String token = item.asText();
                if (!gameArgs || !shouldSkipGameArg(token)) {
                    target.add(substitute(token));
                }
            }
        }
    }

    private RuleEvaluator.FeatureSet featureSet() {
        return new RuleEvaluator.FeatureSet(demo, quickPlay, quickPlayMode, true);
    }

    private boolean shouldSkipGameArg(String arg) {
        if ("--demo".equals(arg) && !demo) {
            return true;
        }
        return QUICKPLAY_ARGS.contains(arg) && !quickPlay;
    }

    public String substitute(String template) {
        return template
                .replace("${auth_player_name}", context.username)
                .replace("${version_name}", context.version)
                .replace("${game_directory}", context.gameDir)
                .replace("${assets_root}", context.assetsDir)
                .replace("${game_assets}", context.assetsDir)
                .replace("${assets_index_name}", context.assetIndex)
                .replace("${auth_uuid}", context.uuid)
                .replace("${auth_access_token}", context.accessToken)
                .replace("${auth_session}", context.accessToken)
                .replace("${user_type}", context.userType)
                .replace("${user_properties}", "{}")
                .replace("${auth_xuid}", context.xuid)
                .replace("${clientid}", context.clientId)
                .replace("${version_type}", context.versionType)
                .replace("${natives_directory}", context.nativesDir)
                .replace("${launcher_name}", context.launcherName)
                .replace("${launcher_version}", context.launcherVersion)
                .replace("${classpath}", context.classpathStr)
                .replace("${classpath_separator}", File.pathSeparator)
                .replace("${library_directory}", context.libraryDir)
                .replace("${resolution_width}", String.valueOf(context.width))
                .replace("${resolution_height}", String.valueOf(context.height));
    }
}
