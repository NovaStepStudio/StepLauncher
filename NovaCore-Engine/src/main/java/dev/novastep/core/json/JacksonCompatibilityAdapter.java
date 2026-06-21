package dev.novastep.core.json;

public abstract class JacksonCompatibilityAdapter {

    public abstract static class JsonNode implements Iterable<JsonNode> {
        protected final JsonParserLite.JsonValue value;

        protected JsonNode(JsonParserLite.JsonValue value) {
            this.value = value;
        }

        public abstract boolean isNull();
        public abstract boolean isMissingNode();
        public abstract boolean isBoolean();
        public abstract boolean isNumber();
        public abstract boolean isString();
        public abstract boolean isTextual();
        public abstract boolean isArray();
        public abstract boolean isObject();
        public abstract boolean isEmpty();

        public abstract JsonNode get(String key);
        public abstract JsonNode get(int index);
        public abstract JsonNode path(String key);
        public abstract boolean has(String key);
        public abstract boolean hasNonNull(String key);

        public abstract String asText();
        public abstract String asText(String defaultValue);
        public abstract int asInt();
        public abstract int asInt(int defaultValue);
        public abstract long asLong();
        public abstract long asLong(long defaultValue);
        public abstract double asDouble();
        public abstract double asDouble(double defaultValue);
        public abstract boolean asBoolean();
        public abstract boolean asBoolean(boolean defaultValue);

        public abstract int size();
        public abstract Iterable<JsonNode> elements();
        public abstract Iterable<String> fieldNames();
        public abstract java.util.Iterator<java.util.Map.Entry<String, JsonNode>> fields();

        @Override
        public java.util.Iterator<JsonNode> iterator() {
            return elements().iterator();
        }
    }

    private static class NullNode extends JsonNode {
        NullNode() {
            super(new JsonParserLite.JsonNull());
        }

        @Override public boolean isNull()        { return true; }
        @Override public boolean isMissingNode() { return true; }
        @Override public boolean isBoolean()     { return false; }
        @Override public boolean isNumber()      { return false; }
        @Override public boolean isString()      { return false; }
        @Override public boolean isTextual()     { return false; }
        @Override public boolean isArray()       { return false; }
        @Override public boolean isObject()      { return false; }
        @Override public boolean isEmpty()       { return true; }
        @Override public JsonNode get(String key) { return NULL; }
        @Override public JsonNode get(int index) { return NULL; }
        @Override public JsonNode path(String key) { return NULL; }
        @Override public boolean has(String key) { return false; }
        @Override public boolean hasNonNull(String key) { return false; }
        @Override public java.util.Iterator<java.util.Map.Entry<String, JsonNode>> fields() { return java.util.Collections.emptyIterator(); }
        @Override public String asText() { return "null"; }
        @Override public String asText(String defaultValue) { return defaultValue; }
        @Override public int asInt() { return 0; }
        @Override public int asInt(int defaultValue) { return defaultValue; }
        @Override public long asLong() { return 0L; }
        @Override public long asLong(long defaultValue) { return defaultValue; }
        @Override public double asDouble() { return 0.0; }
        @Override public double asDouble(double defaultValue) { return defaultValue; }
        @Override public boolean asBoolean() { return false; }
        @Override public boolean asBoolean(boolean defaultValue) { return defaultValue; }
        @Override public int size() { return 0; }
        @Override public Iterable<JsonNode> elements() { return java.util.List.of(); }
        @Override public Iterable<String> fieldNames() { return java.util.List.of(); }
    }

    private static class BooleanNode extends JsonNode {
        BooleanNode(JsonParserLite.JsonValue value) { super(value); }

        @Override public boolean isNull()        { return false; }
        @Override public boolean isMissingNode() { return false; }
        @Override public boolean isBoolean()     { return true; }
        @Override public boolean isNumber()      { return false; }
        @Override public boolean isString()      { return false; }
        @Override public boolean isTextual()     { return false; }
        @Override public boolean isArray()       { return false; }
        @Override public boolean isObject()      { return false; }
        @Override public boolean isEmpty()       { return false; }
        @Override public JsonNode get(String key) { return NULL; }
        @Override public JsonNode get(int index) { return NULL; }
        @Override public JsonNode path(String key) { return NULL; }
        @Override public boolean has(String key) { return false; }
        @Override public boolean hasNonNull(String key) { return false; }
        @Override public java.util.Iterator<java.util.Map.Entry<String, JsonNode>> fields() { return java.util.Collections.emptyIterator(); }
        @Override public String asText() { return String.valueOf(value.asBoolean()); }
        @Override public String asText(String defaultValue) { return asText(); }
        @Override public int asInt() { return value.asBoolean() ? 1 : 0; }
        @Override public int asInt(int defaultValue) { return asInt(); }
        @Override public long asLong() { return value.asBoolean() ? 1L : 0L; }
        @Override public long asLong(long defaultValue) { return asLong(); }
        @Override public double asDouble() { return value.asBoolean() ? 1.0 : 0.0; }
        @Override public double asDouble(double defaultValue) { return asDouble(); }
        @Override public boolean asBoolean() { return value.asBoolean(); }
        @Override public boolean asBoolean(boolean defaultValue) { return asBoolean(); }
        @Override public int size() { return 0; }
        @Override public Iterable<JsonNode> elements() { return java.util.List.of(); }
        @Override public Iterable<String> fieldNames() { return java.util.List.of(); }
    }

    private static class NumberNode extends JsonNode {
        NumberNode(JsonParserLite.JsonValue value) { super(value); }

        @Override public boolean isNull()        { return false; }
        @Override public boolean isMissingNode() { return false; }
        @Override public boolean isBoolean()     { return false; }
        @Override public boolean isNumber()      { return true; }
        @Override public boolean isString()      { return false; }
        @Override public boolean isTextual()     { return false; }
        @Override public boolean isArray()       { return false; }
        @Override public boolean isObject()      { return false; }
        @Override public boolean isEmpty()       { return false; }
        @Override public JsonNode get(String key) { return NULL; }
        @Override public JsonNode get(int index) { return NULL; }
        @Override public JsonNode path(String key) { return NULL; }
        @Override public boolean has(String key) { return false; }
        @Override public boolean hasNonNull(String key) { return false; }
        @Override public java.util.Iterator<java.util.Map.Entry<String, JsonNode>> fields() { return java.util.Collections.emptyIterator(); }
        @Override public String asText() { return value.asString(); }
        @Override public String asText(String defaultValue) { return asText(); }
        @Override public int asInt() { return value.asInt(); }
        @Override public int asInt(int defaultValue) { return asInt(); }
        @Override public long asLong() { return value.asLong(); }
        @Override public long asLong(long defaultValue) { return asLong(); }
        @Override public double asDouble() { return value.asDouble(); }
        @Override public double asDouble(double defaultValue) { return asDouble(); }
        @Override public boolean asBoolean() { return value.asDouble() != 0; }
        @Override public boolean asBoolean(boolean defaultValue) { return asBoolean(); }
        @Override public int size() { return 0; }
        @Override public Iterable<JsonNode> elements() { return java.util.List.of(); }
        @Override public Iterable<String> fieldNames() { return java.util.List.of(); }
    }

    private static class StringNode extends JsonNode {
        StringNode(JsonParserLite.JsonValue value) { super(value); }

        @Override public boolean isNull()        { return false; }
        @Override public boolean isMissingNode() { return false; }
        @Override public boolean isBoolean()     { return false; }
        @Override public boolean isNumber()      { return false; }
        @Override public boolean isString()      { return true; }
        @Override public boolean isTextual()     { return true; }
        @Override public boolean isArray()       { return false; }
        @Override public boolean isObject()      { return false; }
        @Override public boolean isEmpty()       { return value.asString().isEmpty(); }
        @Override public JsonNode get(String key) { return NULL; }
        @Override public JsonNode get(int index) { return NULL; }
        @Override public JsonNode path(String key) { return NULL; }
        @Override public boolean has(String key) { return false; }
        @Override public boolean hasNonNull(String key) { return false; }
        @Override public java.util.Iterator<java.util.Map.Entry<String, JsonNode>> fields() { return java.util.Collections.emptyIterator(); }
        @Override public String asText() { return value.asString(); }
        @Override public String asText(String defaultValue) { return asText(); }
        @Override public int asInt() { try { return Integer.parseInt(value.asString()); } catch (Exception e) { return 0; } }
        @Override public int asInt(int defaultValue) { try { return Integer.parseInt(value.asString()); } catch (Exception e) { return defaultValue; } }
        @Override public long asLong() { try { return Long.parseLong(value.asString()); } catch (Exception e) { return 0L; } }
        @Override public long asLong(long defaultValue) { try { return Long.parseLong(value.asString()); } catch (Exception e) { return defaultValue; } }
        @Override public double asDouble() { try { return Double.parseDouble(value.asString()); } catch (Exception e) { return 0.0; } }
        @Override public double asDouble(double defaultValue) { try { return Double.parseDouble(value.asString()); } catch (Exception e) { return defaultValue; } }
        @Override public boolean asBoolean() { return !value.asString().isEmpty() && !value.asString().equalsIgnoreCase("false"); }
        @Override public boolean asBoolean(boolean defaultValue) { return asBoolean(); }
        @Override public int size() { return 0; }
        @Override public Iterable<JsonNode> elements() { return java.util.List.of(); }
        @Override public Iterable<String> fieldNames() { return java.util.List.of(); }
    }

    private static class ArrayNode extends JsonNode {
        private final JsonParserLite.JsonArray array;

        ArrayNode(JsonParserLite.JsonValue value) {
            super(value);
            this.array = value.asArray();
        }

        @Override public boolean isNull()        { return false; }
        @Override public boolean isMissingNode() { return false; }
        @Override public boolean isBoolean()     { return false; }
        @Override public boolean isNumber()      { return false; }
        @Override public boolean isString()      { return false; }
        @Override public boolean isTextual()     { return false; }
        @Override public boolean isArray()       { return true; }
        @Override public boolean isObject()      { return false; }
        @Override public boolean isEmpty()       { return array.isEmpty(); }
        @Override public JsonNode get(String key) { return NULL; }
        @Override public JsonNode get(int index) { return wrap(array.get(index)); }
        @Override public JsonNode path(String key) { return NULL; }
        @Override public boolean has(String key) { return false; }
        @Override public boolean hasNonNull(String key) { return false; }
        @Override public java.util.Iterator<java.util.Map.Entry<String, JsonNode>> fields() { return java.util.Collections.emptyIterator(); }
        @Override public String asText() { return "[array]"; }
        @Override public String asText(String defaultValue) { return defaultValue; }
        @Override public int asInt() { return 0; }
        @Override public int asInt(int defaultValue) { return defaultValue; }
        @Override public long asLong() { return 0L; }
        @Override public long asLong(long defaultValue) { return defaultValue; }
        @Override public double asDouble() { return 0.0; }
        @Override public double asDouble(double defaultValue) { return defaultValue; }
        @Override public boolean asBoolean() { return !array.isEmpty(); }
        @Override public boolean asBoolean(boolean defaultValue) { return asBoolean(); }
        @Override public int size() { return array.size(); }
        @Override public Iterable<JsonNode> elements() {
            java.util.List<JsonNode> result = new java.util.ArrayList<>();
            for (JsonParserLite.JsonValue v : array) {
                result.add(wrap(v));
            }
            return result;
        }
        @Override public Iterable<String> fieldNames() { return java.util.List.of(); }
    }

    private static class ObjectNode extends JsonNode {
        private final JsonParserLite.JsonObject obj;

        ObjectNode(JsonParserLite.JsonValue value) {
            super(value);
            this.obj = value.asObject();
        }

        @Override public boolean isNull()        { return false; }
        @Override public boolean isMissingNode() { return false; }
        @Override public boolean isBoolean()     { return false; }
        @Override public boolean isNumber()      { return false; }
        @Override public boolean isString()      { return false; }
        @Override public boolean isTextual()     { return false; }
        @Override public boolean isArray()       { return false; }
        @Override public boolean isObject()      { return true; }
        @Override public boolean isEmpty()       { return obj.isEmpty(); }
        @Override public JsonNode get(String key) { return wrap(obj.get(key)); }
        @Override public JsonNode get(int index) { return NULL; }
        @Override public JsonNode path(String key) { return wrap(obj.get(key)); }
        @Override public boolean has(String key) { return obj.has(key); }
        @Override public boolean hasNonNull(String key) { return has(key) && !obj.get(key).isNull(); }
        @Override public java.util.Iterator<java.util.Map.Entry<String, JsonNode>> fields() {
            java.util.List<java.util.Map.Entry<String, JsonNode>> entries = new java.util.ArrayList<>();
            for (String key : obj.keys()) {
                entries.add(new java.util.AbstractMap.SimpleImmutableEntry<>(key, wrap(obj.get(key))));
            }
            return entries.iterator();
        }
        @Override public String asText() { return "{object}"; }
        @Override public String asText(String defaultValue) { return defaultValue; }
        @Override public int asInt() { return 0; }
        @Override public int asInt(int defaultValue) { return defaultValue; }
        @Override public long asLong() { return 0L; }
        @Override public long asLong(long defaultValue) { return defaultValue; }
        @Override public double asDouble() { return 0.0; }
        @Override public double asDouble(double defaultValue) { return defaultValue; }
        @Override public boolean asBoolean() { return !obj.isEmpty(); }
        @Override public boolean asBoolean(boolean defaultValue) { return asBoolean(); }
        @Override public int size() { return obj.size(); }
        @Override public Iterable<JsonNode> elements() { return java.util.List.of(); }
        @Override public Iterable<String> fieldNames() { return obj.keys(); }
    }

    public static final JsonNode NULL = new NullNode();

    public static JsonNode wrap(JsonParserLite.JsonValue value) {
        if (value.isNull()) return NULL;
        if (value.isBoolean()) return new BooleanNode(value);
        if (value.isNumber()) return new NumberNode(value);
        if (value.isString()) return new StringNode(value);
        if (value.isArray()) return new ArrayNode(value);
        if (value.isObject()) return new ObjectNode(value);
        return NULL;
    }

    public static JsonNode readTree(String json) throws java.io.IOException {
        return wrap(JsonParserLite.parse(json));
    }

    public static JsonNode readTree(java.nio.file.Path path) throws java.io.IOException {
        return wrap(JsonParserLite.parseFile(path));
    }
}
