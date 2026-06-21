package dev.novastep.core.modloader.provider;

import com.fasterxml.jackson.core.type.TypeReference;
import dev.novastep.core.json.Json;
import dev.novastep.core.log.CoreLogger;
import dev.novastep.core.modloader.model.ModLoaderModels.LoaderVersion;

import java.io.IOException;
import java.util.*;

public final class ForgeProvider extends AbstractForgeProvider {

    private static final String LOG = "ForgeProvider";
    private static final String JSON_API = "https://files.minecraftforge.net/net/minecraftforge/forge/maven-metadata.json";
    private static final String MAVEN_BASE = "https://maven.minecraftforge.net/";

    @Override
    public String name() {
        return "forge";
    }

    @Override
    protected String installerUrl(String versionId) {
        return MAVEN_BASE + "net/minecraftforge/forge/" + versionId
                + "/forge-" + versionId + "-installer.jar";
    }

    @Override
    protected String mavenRepoBase() {
        return MAVEN_BASE;
    }

    @Override
    protected List<String> listAllVersions() throws IOException, InterruptedException {
        Map<String, List<String>> meta = fetchMeta();
        List<String> all = new ArrayList<>();
        for (List<String> v : meta.values())
            all.addAll(v);
        return all;
    }

    @Override
    protected List<String> filterForMinecraft(List<String> all, String mcVersion) {
        String prefix = mcVersion + "-";
        List<String> result = new ArrayList<>();
        for (String v : all) {
            if (v.startsWith(prefix))
                result.add(v);
        }
        return result;
    }

    @Override
    protected String versionIdForInstaller(String mcVersion, String loaderVersion) {
        if (loaderVersion.startsWith(mcVersion + "-"))
            return loaderVersion;
        return mcVersion + "-" + loaderVersion;
    }

    @Override
    public List<LoaderVersion> getVersions(String mcVersion)
            throws IOException, InterruptedException {

        Map<String, List<String>> meta = fetchMeta();
        List<String> raw = meta.get(mcVersion);

        if (raw == null || raw.isEmpty()) {
            CoreLogger.get().warn(LOG, "No Forge versions found for MC " + mcVersion);
            return Collections.emptyList();
        }

        List<String> ordered = new ArrayList<>(raw);
        Collections.reverse(ordered);

        List<LoaderVersion> result = new ArrayList<>(ordered.size());
        for (String fullId : ordered) {
            String loaderOnly = fullId.startsWith(mcVersion + "-")
                    ? fullId.substring(mcVersion.length() + 1)
                    : fullId;
            result.add(new LoaderVersion(loaderOnly, mcVersion, true));
        }
        return result;
    }

    // ─── Internal ─────────────────────────────────────────────────────────────

    private Map<String, List<String>> fetchMeta() throws IOException, InterruptedException {
        String json = get(JSON_API);
        Map<String, List<String>> meta = Json.read(json,
                new TypeReference<Map<String, List<String>>>() {
                });
        if (meta == null || meta.isEmpty())
            throw new IOException("Forge metadata JSON empty or unparseable: " + JSON_API);
        return meta;
    }
}
