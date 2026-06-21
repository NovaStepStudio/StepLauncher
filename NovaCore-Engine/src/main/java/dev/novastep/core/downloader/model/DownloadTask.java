package dev.novastep.core.downloader.model;

import java.nio.file.Path;

public final class DownloadTask {

    public final String sessionId;
    public final String category;
    public final String name;
    public final String url;
    public final Path destination;
    public final long expectedSize;
    public final String sha1;

    public DownloadTask(String sessionId, String category, String name,
            String url, Path destination, long expectedSize, String sha1) {
        this.sessionId = sessionId;
        this.category = category;
        this.name = name;
        this.url = url;
        this.destination = destination;
        this.expectedSize = expectedSize;
        this.sha1 = sha1;
    }

    public static DownloadTask client(String sessionId, String version,
            String url, Path dest, long size, String sha1) {
        return new DownloadTask(sessionId, "client", version + ".jar", url, dest, size, sha1);
    }

    public static DownloadTask library(String sessionId, String name,
            String url, Path dest, long size, String sha1) {
        return new DownloadTask(sessionId, "library", name, url, dest, size, sha1);
    }

    public static DownloadTask asset(String sessionId, String logicalName,
            String url, Path dest, long size, String sha1) {
        return new DownloadTask(sessionId, "asset", logicalName, url, dest, size, sha1);
    }

    public static DownloadTask nativeLib(String sessionId, String name,
            String url, Path dest, long size, String sha1) {
        return new DownloadTask(sessionId, "native", name, url, dest, size, sha1);
    }

    public static DownloadTask assetIndex(String sessionId, String indexId,
            String url, Path dest, long size, String sha1) {
        return new DownloadTask(sessionId, "asset_index", indexId + ".json", url, dest, size, sha1);
    }

    @Override
    public String toString() {
        return String.format("DownloadTask{category='%s', name='%s', size=%d}", category, name, expectedSize);
    }
}
