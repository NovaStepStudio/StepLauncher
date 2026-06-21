package dev.novastep.core.modloader.provider;

import dev.novastep.core.modloader.model.ModLoaderModels.LoaderVersion;
import dev.novastep.core.modloader.resolver.NeoForgeVersionResolver;

import java.io.IOException;
import java.util.ArrayList;
import java.util.List;

public final class NeoForgeProvider extends AbstractForgeProvider {

    private static final String MAVEN_META = "https://maven.neoforged.net/releases/net/neoforged/neoforge/maven-metadata.xml";
    private static final String MAVEN_BASE = "https://maven.neoforged.net/releases/";

    @Override
    public String name() {
        return "neoforge";
    }

    @Override
    protected String installerUrl(String versionId) {
        return MAVEN_BASE + "net/neoforged/neoforge/" + versionId + "/neoforge-" + versionId + "-installer.jar";
    }

    @Override
    protected String mavenRepoBase() {
        return MAVEN_BASE;
    }

    @Override
    protected List<String> listAllVersions() throws IOException, InterruptedException {
        return parseMavenMetadataVersions(get(MAVEN_META));
    }

    @Override
    protected List<String> filterForMinecraft(List<String> all, String mcVersion) {
        return NeoForgeVersionResolver.filterVersionsForMinecraft(all, mcVersion);
    }

    @Override
    protected String versionIdForInstaller(String mcVersion, String loaderVersion) {
        return loaderVersion;
    }

    @Override
    public List<LoaderVersion> getVersions(String mcVersion) throws IOException, InterruptedException {
        List<String> all      = listAllVersions();
        List<String> filtered = filterForMinecraft(all, mcVersion);
        List<LoaderVersion> result = new ArrayList<>();
        for (String v : filtered) {
            String resolvedMc = NeoForgeVersionResolver.neoForgeVersionToMcVersion(v);
            result.add(new LoaderVersion(v, resolvedMc, true));
        }
        return result;
    }
}
