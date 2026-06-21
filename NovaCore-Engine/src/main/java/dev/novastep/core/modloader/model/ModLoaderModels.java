package dev.novastep.core.modloader.model;

import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

public final class ModLoaderModels {
    private ModLoaderModels() {}

    public static final class LoaderVersion {
        public final String loaderVersion;
        public final String minecraftVersion;
        public final boolean stable;

        public LoaderVersion(String loaderVersion, String minecraftVersion, boolean stable) {
            this.loaderVersion    = loaderVersion;
            this.minecraftVersion = minecraftVersion;
            this.stable           = stable;
        }

        @Override
        public String toString() {
            return "LoaderVersion{loader='" + loaderVersion + "', mc='" + minecraftVersion + "', stable=" + stable + "}";
        }
    }

    public static final class InstalledLoader {
        public String loaderType;
        public String loaderVersion;
        public String minecraftVersion;
        public String versionJsonId;
        public String installerJarPath;
        public long   installedAt;

        public InstalledLoader() {}

        public InstalledLoader(
                String loaderType,
                String loaderVersion,
                String minecraftVersion,
                String versionJsonId,
                String installerJarPath) {
            this.loaderType       = loaderType;
            this.loaderVersion    = loaderVersion;
            this.minecraftVersion = minecraftVersion;
            this.versionJsonId    = versionJsonId;
            this.installerJarPath = installerJarPath;
            this.installedAt      = System.currentTimeMillis();
        }
    }

    public static final class ExecutionPlan {
        public final String      mainClass;
        public final List<Path>  additionalClasspath;
        public final List<String> additionalJvmArgs;
        public final List<String> additionalGameArgs;
        public final boolean     useModulePath;

        public ExecutionPlan(
                String mainClass,
                List<Path> additionalClasspath,
                List<String> additionalJvmArgs,
                List<String> additionalGameArgs,
                boolean useModulePath) {
            this.mainClass           = mainClass;
            this.additionalClasspath = Collections.unmodifiableList(new ArrayList<>(additionalClasspath));
            this.additionalJvmArgs   = Collections.unmodifiableList(new ArrayList<>(additionalJvmArgs));
            this.additionalGameArgs  = Collections.unmodifiableList(new ArrayList<>(additionalGameArgs));
            this.useModulePath       = useModulePath;
        }

        public static ExecutionPlan fromVersionJson(
                String mainClass,
                List<Path> classpath,
                List<String> jvmArgs,
                List<String> gameArgs) {
            return new ExecutionPlan(mainClass, classpath, jvmArgs, gameArgs, false);
        }

        public static ExecutionPlan forBootstrapLauncher(
                String mainClass,
                List<Path> classpath,
                List<String> jvmArgs,
                List<String> gameArgs) {
            return new ExecutionPlan(mainClass, classpath, jvmArgs, gameArgs, true);
        }
    }

    public static final class DownloadPlan {
        private final List<Entry> entries;
        private final boolean     requiresInstaller;
        private final Path        installerDestination;

        public DownloadPlan(List<Entry> entries, boolean requiresInstaller, Path installerDestination) {
            this.entries              = Collections.unmodifiableList(new ArrayList<>(entries));
            this.requiresInstaller    = requiresInstaller;
            this.installerDestination = installerDestination;
        }

        public static DownloadPlan profileOnly(List<Entry> entries) {
            return new DownloadPlan(entries, false, null);
        }

        public static DownloadPlan withInstaller(List<Entry> entries, Path installerDest) {
            return new DownloadPlan(entries, true, installerDest);
        }

        public List<Entry> entries() { return entries; }
        public boolean requiresInstaller() { return requiresInstaller; }
        public Path installerDestination() { return installerDestination; }

        public static final class Entry {
            public final String name;
            public final String url;
            public final Path   destination;
            public final long   size;
            public final String sha1;
            public final String category;

            public Entry(String name, String url, Path destination, long size, String sha1, String category) {
                this.name        = name;
                this.url         = url;
                this.destination = destination;
                this.size        = size;
                this.sha1        = sha1;
                this.category    = category;
            }

            public static Entry library(String name, String url, Path dest, long size, String sha1) {
                return new Entry(name, url, dest, size, sha1, "modloader_library");
            }

            public static Entry installer(String name, String url, Path dest) {
                return new Entry(name, url, dest, -1, null, "modloader_installer");
            }
        }
    }
}
