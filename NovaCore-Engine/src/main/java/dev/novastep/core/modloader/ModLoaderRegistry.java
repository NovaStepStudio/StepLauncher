package dev.novastep.core.modloader;

import dev.novastep.core.modloader.provider.FabricProvider;
import dev.novastep.core.modloader.provider.ForgeProvider;
import dev.novastep.core.modloader.provider.LegacyFabricProvider;
import dev.novastep.core.modloader.provider.NeoForgeProvider;
import dev.novastep.core.modloader.provider.OptiFineProvider;
import dev.novastep.core.modloader.provider.QuiltProvider;

import java.util.Collection;
import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.Map;
import java.util.Optional;

public final class ModLoaderRegistry {

    private static final ModLoaderRegistry INSTANCE = new ModLoaderRegistry();

    private final Map<String, ModLoaderProvider> providers = new LinkedHashMap<>();

    private ModLoaderRegistry() {
        register(new FabricProvider());
        register(new QuiltProvider());
        register(new LegacyFabricProvider());
        register(new ForgeProvider());
        register(new NeoForgeProvider());
        register(new OptiFineProvider());
    }

    public static ModLoaderRegistry get() {
        return INSTANCE;
    }

    public void register(ModLoaderProvider provider) {
        providers.put(provider.name().toLowerCase(), provider);
    }

    public Optional<ModLoaderProvider> find(String name) {
        return Optional.ofNullable(providers.get(name.toLowerCase()));
    }

    public Collection<String> names() {
        return Collections.unmodifiableSet(providers.keySet());
    }
}
