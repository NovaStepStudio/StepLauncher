package dev.novastep.core.minecraft;

import dev.novastep.core.json.JacksonCompatibilityAdapter;
import dev.novastep.core.minecraft.version.VersionInfo;

import java.util.List;
import java.util.Map;

public final class RuleEvaluator {

    private RuleEvaluator() {
    }

    public record FeatureSet(boolean demo, boolean quickPlay, String quickPlayMode, boolean customResolution) {
        public static final FeatureSet DEFAULT = new FeatureSet(false, false, null, true);
    }

    public static boolean isLibraryAllowed(List<VersionInfo.Library.Rule> rules) {
        if (rules == null || rules.isEmpty()) {
            return true;
        }

        boolean result = false;
        for (VersionInfo.Library.Rule rule : rules) {
            boolean matches = rule.os == null || evaluateOs(rule.os);
            if (matches) {
                result = "allow".equals(rule.action);
            }
        }
        return result;
    }

    public static boolean evaluateRules(JacksonCompatibilityAdapter.JsonNode rulesNode, FeatureSet featureSet) {
        if (rulesNode == null || !rulesNode.isArray() || rulesNode.isEmpty()) {
            return true;
        }

        boolean result = false;
        for (JacksonCompatibilityAdapter.JsonNode ruleNode : rulesNode.elements()) {
            if (!ruleNode.isObject()) continue;
            String action = text(ruleNode, "action", "allow");
            boolean matches;
            if (ruleNode.has("features")) {
                matches = evaluateFeatures(ruleNode.get("features"), featureSet);
            } else if (ruleNode.has("os")) {
                matches = evaluateOs(ruleNode.get("os"));
            } else {
                matches = true;
            }
            if (matches) {
                result = "allow".equals(action);
            }
        }
        return result;
    }

    public static boolean evaluateRules(List<Object> rules, FeatureSet featureSet) {
        if (rules == null || rules.isEmpty()) {
            return true;
        }

        boolean result = false;
        for (Object entry : rules) {
            if (!(entry instanceof Map<?, ?> rawMap)) continue;
            @SuppressWarnings("unchecked")
            Map<String, Object> rule = (Map<String, Object>) rawMap;
            String action = String.valueOf(rule.getOrDefault("action", "allow"));
            boolean matches;
            if (rule.containsKey("features")) {
                matches = evaluateFeatures(rule.get("features"), featureSet);
            } else if (rule.containsKey("os")) {
                matches = evaluateOs(rule.get("os"));
            } else {
                matches = true;
            }
            if (matches) {
                result = "allow".equals(action);
            }
        }
        return result;
    }

    public static boolean evaluateOs(Object rawOs) {
        if (rawOs instanceof JacksonCompatibilityAdapter.JsonNode node) {
            return osNameMatches(text(node, "name", null)) && osArchMatches(text(node, "arch", null));
        }
        if (!(rawOs instanceof Map<?, ?> rawMap)) {
            return true;
        }
        @SuppressWarnings("unchecked")
        Map<String, Object> os = (Map<String, Object>) rawMap;
        return osNameMatches(value(os.get("name"))) && osArchMatches(value(os.get("arch")));
    }

    public static boolean evaluateFeatures(Object rawFeatures, FeatureSet featureSet) {
        if (rawFeatures instanceof JacksonCompatibilityAdapter.JsonNode node) {
            return evaluateFeaturesFromJson(node, featureSet);
        }
        if (!(rawFeatures instanceof Map<?, ?> rawMap)) {
            return true;
        }
        @SuppressWarnings("unchecked")
        Map<String, Object> features = (Map<String, Object>) rawMap;
        return evaluateFeaturesFromMap(features, featureSet);
    }

    private static boolean evaluateFeaturesFromJson(JacksonCompatibilityAdapter.JsonNode featuresNode, FeatureSet featureSet) {
        if (featuresNode == null || !featuresNode.isObject()) return true;
        if (featuresNode.has("has_custom_resolution")) return featureSet.customResolution();
        if (featuresNode.has("is_demo_user"))           return featuresNode.get("is_demo_user").asBoolean() == featureSet.demo();
        if (featuresNode.has("is_quick_play_singleplayer")) return "singleplayer".equals(featureSet.quickPlayMode());
        if (featuresNode.has("is_quick_play_multiplayer"))  return "multiplayer".equals(featureSet.quickPlayMode());
        if (featuresNode.has("is_quick_play_realms"))      return "realms".equals(featureSet.quickPlayMode());
        if (featuresNode.has("has_quick_plays_support"))   return featureSet.quickPlay();
        return true;
    }

    private static boolean evaluateFeaturesFromMap(Map<String, Object> features, FeatureSet featureSet) {
        if (features.containsKey("has_custom_resolution")) return featureSet.customResolution();
        if (features.containsKey("is_demo_user"))           return Boolean.parseBoolean(value(features.get("is_demo_user"))) == featureSet.demo();
        if (features.containsKey("is_quick_play_singleplayer")) return "singleplayer".equals(featureSet.quickPlayMode());
        if (features.containsKey("is_quick_play_multiplayer"))  return "multiplayer".equals(featureSet.quickPlayMode());
        if (features.containsKey("is_quick_play_realms"))      return "realms".equals(featureSet.quickPlayMode());
        if (features.containsKey("has_quick_plays_support"))   return featureSet.quickPlay();
        return true;
    }

    public static String currentOsName() {
        String os = System.getProperty("os.name", "").toLowerCase();
        if (os.contains("win")) return "windows";
        if (os.contains("mac")) return "osx";
        return "linux";
    }

    public static String currentArch() {
        String arch = System.getProperty("os.arch", "").toLowerCase();
        if (arch.equals("amd64") || arch.equals("x86_64")) return "x86_64";
        if (arch.equals("aarch64") || arch.equals("arm64")) return "arm64";
        if (arch.equals("x86") || arch.equals("i386") || arch.equals("i686")) return "x86";
        return arch;
    }

    public static boolean osNameMatches(String name) {
        if (name == null || name.isBlank()) return true;
        String os = System.getProperty("os.name", "").toLowerCase();
        return switch (name) {
            case "windows" -> os.contains("win");
            case "osx"     -> os.contains("mac");
            case "linux"   -> !os.contains("win") && !os.contains("mac");
            default        -> false;
        };
    }

    public static boolean osArchMatches(String arch) {
        if (arch == null || arch.isBlank()) return true;
        String current = System.getProperty("os.arch", "").toLowerCase();
        return switch (arch) {
            case "x86"   -> current.equals("x86") || current.equals("i386") || current.equals("i686");
            case "x64"   -> current.equals("amd64") || current.equals("x86_64");
            case "arm64" -> current.equals("aarch64") || current.equals("arm64");
            default      -> current.contains(arch.toLowerCase());
        };
    }

    private static String text(JacksonCompatibilityAdapter.JsonNode node, String field, String fallback) {
        if (node == null || !node.has(field) || node.get(field).isNull()) return fallback;
        return node.get(field).asText();
    }

    private static String value(Object value) {
        return value == null ? null : String.valueOf(value);
    }
}