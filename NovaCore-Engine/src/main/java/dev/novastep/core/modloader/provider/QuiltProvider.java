package dev.novastep.core.modloader.provider;

public final class QuiltProvider extends AbstractFabricProvider {

    private static final String BASE = "https://meta.quiltmc.org/v3";

    @Override
    public String name() {
        return "quilt";
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
