package dev.novastep.core.json;

import com.fasterxml.jackson.core.type.TypeReference;
import java.io.IOException;
import java.lang.reflect.Array;
import java.lang.reflect.Constructor;
import java.lang.reflect.Field;
import java.lang.reflect.GenericArrayType;
import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.ParameterizedType;
import java.lang.reflect.Type;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Collection;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

public final class Json {

    private Json() {
    }

    public static JacksonCompatibilityAdapter.JsonNode readTree(String json) throws IOException {
        return JacksonCompatibilityAdapter.readTree(json);
    }

    public static JacksonCompatibilityAdapter.JsonNode readTree(Path path) throws IOException {
        return JacksonCompatibilityAdapter.readTree(path);
    }

    public static <T> T read(String json, Class<T> type) throws IOException {
        return fromJsonValue(JsonParserLite.parse(json), type);
    }

    public static <T> T read(Path path, Class<T> type) throws IOException {
        return read(Files.readString(path, StandardCharsets.UTF_8), type);
    }

    public static <T> T read(String json, TypeReference<T> typeRef) throws IOException {
        return fromJsonValue(JsonParserLite.parse(json), typeRef.getType());
    }

    public static <T> T read(Path path, TypeReference<T> typeRef) throws IOException {
        return read(Files.readString(path, StandardCharsets.UTF_8), typeRef);
    }

    public static String write(JsonParserLite.JsonValue value) throws IOException {
        return JsonParserLite.stringify(value, false);
    }

    public static String writePretty(JsonParserLite.JsonValue value) throws IOException {
        return JsonParserLite.stringify(value, true);
    }

    public static String write(Object value) throws IOException {
        return JsonParserLite.stringify(toJsonValue(value), false);
    }

    public static String writePretty(Object value) throws IOException {
        return JsonParserLite.stringify(toJsonValue(value), true);
    }

    public static void write(Path path, JsonParserLite.JsonValue value, boolean pretty) throws IOException {
        Files.createDirectories(path.getParent());
        String json = pretty ? writePretty(value) : write(value);
        Files.writeString(path, json, StandardCharsets.UTF_8);
    }

    public static void write(Path path, Object value, boolean pretty) throws IOException {
        Files.createDirectories(path.getParent());
        String json = pretty ? writePretty(value) : write(value);
        Files.writeString(path, json, StandardCharsets.UTF_8);
    }

    public static JsonParserLite.JsonObject object() {
        return new JsonParserLite.JsonObject();
    }

    public static JsonParserLite.JsonArray array() {
        return new JsonParserLite.JsonArray();
    }

    public static JacksonCompatibilityAdapter.JsonNode emptyArrayNode() {
        return JacksonCompatibilityAdapter.wrap(array());
    }

    public static JacksonCompatibilityAdapter.JsonNode emptyObjectNode() {
        return JacksonCompatibilityAdapter.wrap(object());
    }

    private static JsonParserLite.JsonValue toJsonValue(Object value) throws IOException {
        if (value == null) {
            return new JsonParserLite.JsonNull();
        }
        if (value instanceof JsonParserLite.JsonValue jsonValue) {
            return jsonValue;
        }
        if (value instanceof Boolean bool) {
            return new JsonParserLite.JsonBoolean(bool);
        }
        if (value instanceof Number number) {
            return new JsonParserLite.JsonNumber(number.toString());
        }
        if (value instanceof Character ch) {
            return new JsonParserLite.JsonString(String.valueOf(ch));
        }
        if (value instanceof String str) {
            return new JsonParserLite.JsonString(str);
        }
        if (value instanceof Map<?, ?> map) {
            JsonParserLite.JsonObject object = object();
            for (Map.Entry<?, ?> entry : map.entrySet()) {
                object.set(String.valueOf(entry.getKey()), toJsonValue(entry.getValue()));
            }
            return object;
        }
        if (value instanceof Iterable<?> iterable) {
            JsonParserLite.JsonArray array = array();
            for (Object item : iterable) {
                array.add(toJsonValue(item));
            }
            return array;
        }
        if (value.getClass().isArray()) {
            JsonParserLite.JsonArray array = array();
            int length = Array.getLength(value);
            for (int i = 0; i < length; i++) {
                array.add(toJsonValue(Array.get(value, i)));
            }
            return array;
        }
        if (value instanceof Enum<?> en) {
            return new JsonParserLite.JsonString(en.name());
        }

        JsonParserLite.JsonObject object = object();
        for (Field field : getAllFields(value.getClass())) {
            if (java.lang.reflect.Modifier.isStatic(field.getModifiers())) {
                continue;
            }
            field.setAccessible(true);
            try {
                object.set(field.getName(), toJsonValue(field.get(value)));
            } catch (IllegalAccessException e) {
                throw new IOException("Cannot serialize field " + field.getName(), e);
            }
        }
        return object;
    }

    @SuppressWarnings("unchecked")
    private static <T> T fromJsonValue(JsonParserLite.JsonValue json, Class<T> targetType) throws IOException {
        Object value = fromJsonValue(json, (Type) targetType);
        return (T) value;
    }

    @SuppressWarnings("unchecked")
    private static <T> T fromJsonValue(JsonParserLite.JsonValue json, Type targetType) throws IOException {
        if (targetType instanceof Class<?> cls) {
            if (json.isNull()) {
                return null;
            }
            if (cls == Object.class) {
                return (T) toUntyped(json);
            }
            if (cls == String.class) {
                return (T) json.asString();
            }
            if (cls == Boolean.class || cls == boolean.class) {
                return (T) Boolean.valueOf(json.asBoolean());
            }
            if (cls == Integer.class || cls == int.class) {
                return (T) Integer.valueOf(json.asInt());
            }
            if (cls == Long.class || cls == long.class) {
                return (T) Long.valueOf(json.asLong());
            }
            if (cls == Double.class || cls == double.class) {
                return (T) Double.valueOf(json.asDouble());
            }
            if (cls.isEnum()) {
                @SuppressWarnings("rawtypes")
                Class<? extends Enum> enumClass = (Class<? extends Enum>) cls;
                return (T) Enum.valueOf(enumClass, json.asString());
            }
            if (Map.class.isAssignableFrom(cls) && json.isObject()) {
                return (T) toUntyped(json);
            }
            if (Collection.class.isAssignableFrom(cls) && json.isArray()) {
                return (T) toUntyped(json);
            }
            if (json.isObject()) {
                try {
                    Constructor<?> ctor = cls.getDeclaredConstructor();
                    ctor.setAccessible(true);
                    T instance = (T) ctor.newInstance();
                    JsonParserLite.JsonObject object = json.asObject();
                    for (Field field : getAllFields(cls)) {
                        if (java.lang.reflect.Modifier.isStatic(field.getModifiers())) {
                            continue;
                        }
                        field.setAccessible(true);
                        JsonParserLite.JsonValue fieldJson = object.get(field.getName());
                        if (fieldJson.isNull()) {
                            if (!field.getType().isPrimitive()) {
                                field.set(instance, null);
                            }
                            continue;
                        }
                        field.set(instance, fromJsonValue(fieldJson, field.getGenericType()));
                    }
                    return instance;
                } catch (NoSuchMethodException | InstantiationException | IllegalAccessException | InvocationTargetException e) {
                    throw new IOException("Cannot instantiate " + cls.getName(), e);
                }
            }
            throw new IOException("Unsupported target type: " + cls.getName());
        }
        if (targetType instanceof ParameterizedType parameterized) {
            Type rawType = parameterized.getRawType();
            if (rawType instanceof Class<?> rawClass) {
                if (List.class.isAssignableFrom(rawClass) && json.isArray()) {
                    Type elementType = parameterized.getActualTypeArguments()[0];
                    List<Object> list = new ArrayList<>();
                    for (JsonParserLite.JsonValue element : json.asArray().toList()) {
                        list.add(fromJsonValue(element, elementType));
                    }
                    return (T) list;
                }
                if (Map.class.isAssignableFrom(rawClass) && json.isObject()) {
                    Type valueType = parameterized.getActualTypeArguments()[1];
                    Map<String, Object> map = new LinkedHashMap<>();
                    for (String key : json.asObject().keys()) {
                        map.put(key, fromJsonValue(json.asObject().get(key), valueType));
                    }
                    return (T) map;
                }
                return (T) fromJsonValue(json, rawClass);
            }
        }
        if (targetType instanceof GenericArrayType genericArray) {
            Type componentType = genericArray.getGenericComponentType();
            if (json.isArray()) {
                List<Object> values = new ArrayList<>();
                for (JsonParserLite.JsonValue element : json.asArray().toList()) {
                    values.add(fromJsonValue(element, componentType));
                }
                Object array = Array.newInstance((Class<?>) componentType, values.size());
                for (int i = 0; i < values.size(); i++) {
                    Array.set(array, i, values.get(i));
                }
                return (T) array;
            }
        }
        throw new IOException("Unsupported TypeReference target: " + targetType);
    }

    private static Object toUntyped(JsonParserLite.JsonValue json) throws IOException {
        if (json.isNull()) {
            return null;
        }
        if (json.isBoolean()) {
            return json.asBoolean();
        }
        if (json.isNumber()) {
            String text = json.asString();
            if (text.contains(".") || text.contains("e") || text.contains("E")) {
                return json.asDouble();
            }
            return json.asLong();
        }
        if (json.isString()) {
            return json.asString();
        }
        if (json.isArray()) {
            List<Object> list = new ArrayList<>();
            for (JsonParserLite.JsonValue element : json.asArray().toList()) {
                list.add(toUntyped(element));
            }
            return list;
        }
        if (json.isObject()) {
            Map<String, Object> map = new LinkedHashMap<>();
            for (String key : json.asObject().keys()) {
                map.put(key, toUntyped(json.asObject().get(key)));
            }
            return map;
        }
        throw new IOException("Unsupported JSON value for untyped conversion");
    }

    private static List<Field> getAllFields(Class<?> clazz) {
        List<Field> fields = new ArrayList<>();
        while (clazz != null && clazz != Object.class) {
            for (Field field : clazz.getDeclaredFields()) {
                fields.add(field);
            }
            clazz = clazz.getSuperclass();
        }
        return fields;
    }
}
