package dev.novastep.core.modloader.provider;

public final class LegacyFabricProvider extends AbstractFabricProvider {

    private static final String BASE = "https://meta.legacyfabric.net/v2";

    @Override
    public String name() {
        return "legacyfabric";
    }

    @Override
    protected String versionsEndpoint(String mcVersion) {
        return BASE + "/versions/loader/" + mcVersion;
    }

    @Override
    protected String profileEndpoint(String mcVersion, String loaderVersion) {
        return BASE + "/versions/loader/" + mcVersion + "/" + loaderVersion + "/profile/json";
    }
}
