package dev.novastep.core.modloader.model;

import java.util.List;
import java.util.Map;

public final class InstallProfile {

    public int                  spec;
    public String               profile;
    public String               version;
    public String               minecraft;
    public String               serverJarPath;
    public Map<String, DataVal> data;
    public List<Processor>      processors;
    public List<Library>        libraries;

    public static final class DataVal {
        public String client;
        public String server;
    }

    public static final class Processor {
        public List<String>        sides;
        public String              jar;
        public List<String>        classpath;
        public List<String>        args;
        public Map<String, String> outputs;

        public boolean isClientSide() {
            return sides == null || sides.isEmpty() || sides.contains("client");
        }
    }

    public static final class Library {
        public String      name;
        public LibDownload downloads;

        public static final class LibDownload {
            public Artifact artifact;

            public static final class Artifact {
                public String sha1;
                public long   size;
                public String url;
                public String path;
            }
        }
    }
}
