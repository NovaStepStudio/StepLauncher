package dev.novastep.core.json;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.*;

public final class JsonParserLite {

    private JsonParserLite() {}

    public static JsonValue parse(String json) throws IOException {
        if (json == null || json.isBlank()) {
            throw new IOException("Invalid JSON: null or empty string");
        }
        Parser parser = new Parser(json.strip());
        JsonValue result = parser.parseValue();
        parser.skipWhitespace();
        if (parser.pos < parser.input.length()) {
            throw new IOException("Trailing characters after JSON at position " + parser.pos);
        }
        return result;
    }

    public static JsonValue parseFile(Path path) throws IOException {
        String content = Files.readString(path, StandardCharsets.UTF_8);
        return parse(content);
    }

    public static String stringify(JsonValue value) {
        return stringify(value, false);
    }

    public static String stringify(JsonValue value, boolean pretty) {
        StringBuilder sb = new StringBuilder();
        stringify(value, sb, pretty, 0);
        return sb.toString();
    }

    private static void stringify(JsonValue value, StringBuilder sb, boolean pretty, int indent) {
        if (value.isNull()) {
            sb.append("null");
        } else if (value.isBoolean()) {
            sb.append(value.asBoolean());
        } else if (value.isNumber()) {
            String numStr = value.asString();
            sb.append(numStr);
        } else if (value.isString()) {
            sb.append('"');
            escapeJson(value.asString(), sb);
            sb.append('"');
        } else if (value.isArray()) {
            JsonArray arr = value.asArray();
            sb.append('[');
            if (pretty && !arr.isEmpty()) sb.append('\n');
            for (int i = 0; i < arr.size(); i++) {
                if (pretty) sb.append("  ".repeat(indent + 1));
                stringify(arr.get(i), sb, pretty, indent + 1);
                if (i < arr.size() - 1) sb.append(',');
                if (pretty) sb.append('\n');
            }
            if (pretty && !arr.isEmpty()) sb.append("  ".repeat(indent));
            sb.append(']');
        } else if (value.isObject()) {
            JsonObject obj = value.asObject();
            sb.append('{');
            if (pretty && !obj.isEmpty()) sb.append('\n');
            List<String> keys = new ArrayList<>(obj.keys());
            for (int i = 0; i < keys.size(); i++) {
                String key = keys.get(i);
                if (pretty) sb.append("  ".repeat(indent + 1));
                sb.append('"');
                escapeJson(key, sb);
                sb.append("\":");
                if (pretty) sb.append(' ');
                stringify(obj.get(key), sb, pretty, indent + 1);
                if (i < keys.size() - 1) sb.append(',');
                if (pretty) sb.append('\n');
            }
            if (pretty && !obj.isEmpty()) sb.append("  ".repeat(indent));
            sb.append('}');
        }
    }

    private static void escapeJson(String str, StringBuilder sb) {
        for (char c : str.toCharArray()) {
            switch (c) {
                case '"' -> sb.append("\\\"");
                case '\\' -> sb.append("\\\\");
                case '\b' -> sb.append("\\b");
                case '\f' -> sb.append("\\f");
                case '\n' -> sb.append("\\n");
                case '\r' -> sb.append("\\r");
                case '\t' -> sb.append("\\t");
                default -> {
                    if (c < 32) {
                        sb.append(String.format("\\u%04x", (int) c));
                    } else {
                        sb.append(c);
                    }
                }
            }
        }
    }

    public sealed interface JsonValue {
        boolean isNull();
        boolean isBoolean();
        boolean isNumber();
        boolean isString();
        boolean isArray();
        boolean isObject();

        boolean asBoolean();
        double asDouble();
        long asLong();
        int asInt();
        String asString();
        JsonArray asArray();
        JsonObject asObject();
    }

    public static final class JsonNull implements JsonValue {
        @Override public boolean isNull()     { return true; }
        @Override public boolean isBoolean()  { return false; }
        @Override public boolean isNumber()   { return false; }
        @Override public boolean isString()   { return false; }
        @Override public boolean isArray()    { return false; }
        @Override public boolean isObject()   { return false; }
        @Override public boolean asBoolean()  { throw new UnsupportedOperationException("null is not boolean"); }
        @Override public double asDouble()    { throw new UnsupportedOperationException("null is not number"); }
        @Override public long asLong()        { throw new UnsupportedOperationException("null is not number"); }
        @Override public int asInt()          { throw new UnsupportedOperationException("null is not number"); }
        @Override public String asString()    { return "null"; }
        @Override public JsonArray asArray()  { throw new UnsupportedOperationException("null is not array"); }
        @Override public JsonObject asObject() { throw new UnsupportedOperationException("null is not object"); }
    }

    public static final class JsonBoolean implements JsonValue {
        public final boolean value;
        JsonBoolean(boolean value) { this.value = value; }
        @Override public boolean isNull()     { return false; }
        @Override public boolean isBoolean()  { return true; }
        @Override public boolean isNumber()   { return false; }
        @Override public boolean isString()   { return false; }
        @Override public boolean isArray()    { return false; }
        @Override public boolean isObject()   { return false; }
        @Override public boolean asBoolean()  { return value; }
        @Override public double asDouble()    { throw new UnsupportedOperationException("boolean is not number"); }
        @Override public long asLong()        { throw new UnsupportedOperationException("boolean is not number"); }
        @Override public int asInt()          { throw new UnsupportedOperationException("boolean is not number"); }
        @Override public String asString()    { return String.valueOf(value); }
        @Override public JsonArray asArray()  { throw new UnsupportedOperationException("boolean is not array"); }
        @Override public JsonObject asObject() { throw new UnsupportedOperationException("boolean is not object"); }
    }

    public static final class JsonNumber implements JsonValue {
        public final String numStr;
        private final double doubleValue;
        private final long longValue;
        public final boolean isFloat;

        JsonNumber(String numStr) throws IOException {
            this.numStr = numStr;
            try {
                if (numStr.contains(".") || numStr.contains("e") || numStr.contains("E")) {
                    this.doubleValue = Double.parseDouble(numStr);
                    this.longValue = (long) doubleValue;
                    this.isFloat = true;
                } else {
                    this.longValue = Long.parseLong(numStr);
                    this.doubleValue = (double) longValue;
                    this.isFloat = false;
                }
            } catch (NumberFormatException e) {
                throw new IOException("Invalid number: " + numStr, e);
            }
        }

        @Override public boolean isNull()     { return false; }
        @Override public boolean isBoolean()  { return false; }
        @Override public boolean isNumber()   { return true; }
        @Override public boolean isString()   { return false; }
        @Override public boolean isArray()    { return false; }
        @Override public boolean isObject()   { return false; }
        @Override public boolean asBoolean()  { throw new UnsupportedOperationException("number is not boolean"); }
        @Override public double asDouble()    { return doubleValue; }
        @Override public long asLong()        { return longValue; }
        @Override public int asInt()          { return (int) longValue; }
        @Override public String asString()    { return numStr; }
        @Override public JsonArray asArray()  { throw new UnsupportedOperationException("number is not array"); }
        @Override public JsonObject asObject() { throw new UnsupportedOperationException("number is not object"); }
    }

    public static final class JsonString implements JsonValue {
        public final String value;
        JsonString(String value) { this.value = value; }
        @Override public boolean isNull()     { return false; }
        @Override public boolean isBoolean()  { return false; }
        @Override public boolean isNumber()   { return false; }
        @Override public boolean isString()   { return true; }
        @Override public boolean isArray()    { return false; }
        @Override public boolean isObject()   { return false; }
        @Override public boolean asBoolean()  { throw new UnsupportedOperationException("string is not boolean"); }
        @Override public double asDouble()    { throw new UnsupportedOperationException("string is not number"); }
        @Override public long asLong()        { throw new UnsupportedOperationException("string is not number"); }
        @Override public int asInt()          { throw new UnsupportedOperationException("string is not number"); }
        @Override public String asString()    { return value; }
        @Override public JsonArray asArray()  { throw new UnsupportedOperationException("string is not array"); }
        @Override public JsonObject asObject() { throw new UnsupportedOperationException("string is not object"); }
    }

    public static final class JsonArray implements JsonValue, Iterable<JsonValue> {
        private final List<JsonValue> values = new ArrayList<>();

        public void add(JsonValue value) {
            if (value == null) values.add(JSON_NULL);
            else values.add(value);
        }

        public JsonValue get(int index) {
            if (index < 0 || index >= values.size()) return JSON_NULL;
            return values.get(index);
        }

        public int size() { return values.size(); }
        public boolean isEmpty() { return values.isEmpty(); }
        public List<JsonValue> toList() { return new ArrayList<>(values); }

        @Override public Iterator<JsonValue> iterator() { return values.iterator(); }

        @Override public boolean isNull()     { return false; }
        @Override public boolean isBoolean()  { return false; }
        @Override public boolean isNumber()   { return false; }
        @Override public boolean isString()   { return false; }
        @Override public boolean isArray()    { return true; }
        @Override public boolean isObject()   { return false; }
        @Override public boolean asBoolean()  { throw new UnsupportedOperationException("array is not boolean"); }
        @Override public double asDouble()    { throw new UnsupportedOperationException("array is not number"); }
        @Override public long asLong()        { throw new UnsupportedOperationException("array is not number"); }
        @Override public int asInt()          { throw new UnsupportedOperationException("array is not number"); }
        @Override public String asString()    { return "JsonArray(" + values.size() + " items)"; }
        @Override public JsonArray asArray()  { return this; }
        @Override public JsonObject asObject() { throw new UnsupportedOperationException("array is not object"); }
    }

    public static final class JsonObject implements JsonValue, Iterable<String> {
        private final Map<String, JsonValue> fields = new LinkedHashMap<>();

        public void set(String key, JsonValue value) {
            if (value == null) fields.put(key, JSON_NULL);
            else fields.put(key, value);
        }

        public JsonValue get(String key) {
            return fields.getOrDefault(key, JSON_NULL);
        }

        public String getString(String key) {
            JsonValue v = get(key);
            if (v.isString()) return v.asString();
            if (v.isNumber()) return v.asString();
            return null;
        }

        public int getInt(String key) {
            JsonValue v = get(key);
            if (v.isNumber()) return v.asInt();
            return 0;
        }

        public long getLong(String key) {
            JsonValue v = get(key);
            if (v.isNumber()) return v.asLong();
            return 0L;
        }

        public double getDouble(String key) {
            JsonValue v = get(key);
            if (v.isNumber()) return v.asDouble();
            return 0.0;
        }

        public boolean getBoolean(String key) {
            JsonValue v = get(key);
            if (v.isBoolean()) return v.asBoolean();
            return false;
        }

        public JsonArray getArray(String key) {
            JsonValue v = get(key);
            if (v.isArray()) return v.asArray();
            return null;
        }

        public JsonObject getObject(String key) {
            JsonValue v = get(key);
            if (v.isObject()) return v.asObject();
            return null;
        }

        public Set<String> keys() { return new LinkedHashSet<>(fields.keySet()); }
        public Collection<JsonValue> values() { return new ArrayList<>(fields.values()); }
        public int size() { return fields.size(); }
        public boolean isEmpty() { return fields.isEmpty(); }
        public boolean has(String key) { return fields.containsKey(key); }

        @Override public Iterator<String> iterator() { return fields.keySet().iterator(); }

        @Override public boolean isNull()     { return false; }
        @Override public boolean isBoolean()  { return false; }
        @Override public boolean isNumber()   { return false; }
        @Override public boolean isString()   { return false; }
        @Override public boolean isArray()    { return false; }
        @Override public boolean isObject()   { return true; }
        @Override public boolean asBoolean()  { throw new UnsupportedOperationException("object is not boolean"); }
        @Override public double asDouble()    { throw new UnsupportedOperationException("object is not number"); }
        @Override public long asLong()        { throw new UnsupportedOperationException("object is not number"); }
        @Override public int asInt()          { throw new UnsupportedOperationException("object is not number"); }
        @Override public String asString()    { return "JsonObject(" + fields.size() + " fields)"; }
        @Override public JsonArray asArray()  { throw new UnsupportedOperationException("object is not array"); }
        @Override public JsonObject asObject() { return this; }
    }

    private static final JsonNull JSON_NULL = new JsonNull();

    private static class Parser {
        private final String input;
        private int pos = 0;

        Parser(String input) {
            this.input = input;
        }

        private JsonValue parseValue() throws IOException {
            skipWhitespace();
            if (pos >= input.length()) throw new IOException("Unexpected end of JSON");

            char ch = input.charAt(pos);
            return switch (ch) {
                case '{' -> parseObject();
                case '[' -> parseArray();
                case '"' -> parseString();
                case 't', 'f' -> parseBoolean();
                case 'n' -> parseNull();
                case '-', '0','1','2','3','4','5','6','7','8','9' -> parseNumber();
                default -> throw new IOException("Unexpected character at position " + pos + ": " + ch);
            };
        }

        private JsonValue parseNull() throws IOException {
            if (input.startsWith("null", pos)) {
                pos += 4;
                return JSON_NULL;
            }
            throw new IOException("Invalid null at position " + pos);
        }

        private JsonValue parseBoolean() throws IOException {
            if (input.startsWith("true", pos)) {
                pos += 4;
                return new JsonBoolean(true);
            }
            if (input.startsWith("false", pos)) {
                pos += 5;
                return new JsonBoolean(false);
            }
            throw new IOException("Invalid boolean at position " + pos);
        }

        private JsonValue parseNumber() throws IOException {
            int start = pos;
            if (pos < input.length() && input.charAt(pos) == '-') pos++;
            
            if (pos >= input.length()) throw new IOException("Invalid number at position " + start);

            while (pos < input.length() && Character.isDigit(input.charAt(pos))) pos++;

            if (pos < input.length() && input.charAt(pos) == '.') {
                pos++;
                if (pos >= input.length() || !Character.isDigit(input.charAt(pos))) {
                    throw new IOException("Invalid number at position " + pos);
                }
                while (pos < input.length() && Character.isDigit(input.charAt(pos))) pos++;
            }

            if (pos < input.length() && (input.charAt(pos) == 'e' || input.charAt(pos) == 'E')) {
                pos++;
                if (pos < input.length() && (input.charAt(pos) == '+' || input.charAt(pos) == '-')) pos++;
                if (pos >= input.length() || !Character.isDigit(input.charAt(pos))) {
                    throw new IOException("Invalid number at position " + pos);
                }
                while (pos < input.length() && Character.isDigit(input.charAt(pos))) pos++;
            }

            String numStr = input.substring(start, pos);
            return new JsonNumber(numStr);
        }

        private JsonValue parseString() throws IOException {
            if (input.charAt(pos) != '"') throw new IOException("Expected '\"' at position " + pos);
            pos++;

            StringBuilder sb = new StringBuilder();
            while (pos < input.length()) {
                char ch = input.charAt(pos);
                if (ch == '"') {
                    pos++;
                    return new JsonString(sb.toString());
                } else if (ch == '\\') {
                    pos++;
                    if (pos >= input.length()) throw new IOException("Unterminated string escape");
                    char escaped = input.charAt(pos);
                    switch (escaped) {
                        case '"' -> sb.append('"');
                        case '\\' -> sb.append('\\');
                        case '/' -> sb.append('/');
                        case 'b' -> sb.append('\b');
                        case 'f' -> sb.append('\f');
                        case 'n' -> sb.append('\n');
                        case 'r' -> sb.append('\r');
                        case 't' -> sb.append('\t');
                        case 'u' -> {
                            pos++;
                            if (pos + 3 >= input.length()) throw new IOException("Invalid unicode escape");
                            String hex = input.substring(pos, pos + 4);
                            try {
                                sb.append((char) Integer.parseInt(hex, 16));
                                pos += 3;
                            } catch (NumberFormatException e) {
                                throw new IOException("Invalid unicode escape: " + hex);
                            }
                        }
                        default -> throw new IOException("Invalid escape character: \\" + escaped);
                    }
                    pos++;
                } else if (ch < 32) {
                    throw new IOException("Unescaped control character in string at position " + pos);
                } else {
                    sb.append(ch);
                    pos++;
                }
            }
            throw new IOException("Unterminated string");
        }

        private JsonValue parseArray() throws IOException {
            if (input.charAt(pos) != '[') throw new IOException("Expected '[' at position " + pos);
            pos++;
            JsonArray array = new JsonArray();

            skipWhitespace();
            if (pos < input.length() && input.charAt(pos) == ']') {
                pos++;
                return array;
            }

            while (true) {
                array.add(parseValue());
                skipWhitespace();

                if (pos >= input.length()) throw new IOException("Unterminated array");
                char ch = input.charAt(pos);
                if (ch == ']') {
                    pos++;
                    return array;
                } else if (ch == ',') {
                    pos++;
                    skipWhitespace();
                } else {
                    throw new IOException("Expected ',' or ']' in array at position " + pos);
                }
            }
        }

        private JsonValue parseObject() throws IOException {
            if (input.charAt(pos) != '{') throw new IOException("Expected '{' at position " + pos);
            pos++;
            JsonObject object = new JsonObject();

            skipWhitespace();
            if (pos < input.length() && input.charAt(pos) == '}') {
                pos++;
                return object;
            }

            while (true) {
                skipWhitespace();
                if (pos >= input.length()) throw new IOException("Unterminated object");

                JsonValue keyVal = parseValue();
                if (!keyVal.isString()) throw new IOException("Object key must be string at position " + pos);
                String key = keyVal.asString();

                skipWhitespace();
                if (pos >= input.length() || input.charAt(pos) != ':') {
                    throw new IOException("Expected ':' after key at position " + pos);
                }
                pos++;

                JsonValue value = parseValue();
                object.set(key, value);

                skipWhitespace();
                if (pos >= input.length()) throw new IOException("Unterminated object");
                char ch = input.charAt(pos);
                if (ch == '}') {
                    pos++;
                    return object;
                } else if (ch == ',') {
                    pos++;
                } else {
                    throw new IOException("Expected ',' or '}' in object at position " + pos);
                }
            }
        }

        private void skipWhitespace() {
            while (pos < input.length() && Character.isWhitespace(input.charAt(pos))) {
                pos++;
            }
        }
    }
}
