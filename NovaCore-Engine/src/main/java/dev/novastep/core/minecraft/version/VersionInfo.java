package dev.novastep.core.minecraft.version;

import dev.novastep.core.minecraft.RuleEvaluator;

import java.util.List;
import java.util.Map;

public class VersionInfo {

    public String id;
    public String type;
    public String mainClass;
    public String assets;
    public String inheritsFrom;
    public Arguments arguments;
    public String minecraftArguments;
    public Downloads downloads;
    public List<Library> libraries;
    public AssetIndex assetIndex;
    public JavaVersion javaVersion;
    public LoggingConfig logging;
    public int minimumLauncherVersion;

    public static class Downloads {
        public Artifact client;
        public Artifact server;
        public Artifact client_mappings;
        public Artifact server_mappings;
    }

    public static class Artifact {
        public String sha1;
        public long   size;
        public String url;
        public String path;
    }

    public static class Library {
        public String name;
        public LibDownloads downloads;
        public Map<String, String> natives;
        public List<Rule> rules;

        public boolean isAllowed() {
            return RuleEvaluator.isLibraryAllowed(rules);
        }

        public String getNativeClassifier() {
            if (natives == null) return null;
            return natives.get(RuleEvaluator.currentOsName());
        }

        public static class LibDownloads {
            public Artifact artifact;
            public Map<String, Artifact> classifiers;
        }

        public static class Rule {
            public String action;
            public Map<String, String> os;
        }
    }

    public static class AssetIndex {
        public String id;
        public String sha1;
        public long size;
        public long totalSize;
        public String url;
    }

    public static class JavaVersion {
        public String component;
        public int majorVersion;
    }

    public static class LoggingConfig {
        public ClientLogging client;
        public static class ClientLogging {
            public String argument;
            public Artifact file;
            public String type;
        }
    }

    public static class Arguments {
        public List<Object> game;
        public List<Object> jvm;
    }
}
