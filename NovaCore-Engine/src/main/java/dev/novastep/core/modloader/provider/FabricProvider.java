package dev.novastep.core.modloader.provider;

public final class FabricProvider extends AbstractFabricProvider {

    private static final String BASE = "https://meta.fabricmc.net/v2";

    @Override
    public String name() {
        return "fabric";
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
