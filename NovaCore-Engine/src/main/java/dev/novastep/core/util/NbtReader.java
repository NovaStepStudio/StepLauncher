package dev.novastep.core.util;

import java.io.*;
import java.nio.charset.StandardCharsets;
import java.util.*;
import java.util.zip.GZIPInputStream;
import java.util.zip.InflaterInputStream;

public class NbtReader {

    private final DataInputStream in;

    public NbtReader(InputStream is) throws IOException {
        this.in = new DataInputStream(detectCompression(is));
    }

    private InputStream detectCompression(InputStream is) throws IOException {
        BufferedInputStream bis = new BufferedInputStream(is);
        bis.mark(2);
        int b1 = bis.read();
        int b2 = bis.read();
        bis.reset();

        if (b1 == 0x1f && b2 == 0x8b)
            return new GZIPInputStream(bis);
        if (b1 == 0x78)
            return new InflaterInputStream(bis);
        return bis;
    }

    public Map<String, Object> parse() throws IOException {
        try {
            byte type = in.readByte();
            if (type == 0)
                return Collections.emptyMap();
            if (type != 10)
                throw new IOException("Expected Compound Tag (10) at root, found: " + type);

            readString();
            return readCompound();
        } catch (EOFException e) {
            throw new IOException("Unexpected EOF", e);
        }
    }

    private Map<String, Object> readCompound() throws IOException {
        Map<String, Object> map = new HashMap<>();
        while (true) {
            byte type = in.readByte();
            if (type == 0)
                break;
            String name = readString();
            Object value = readTag(type);
            if (value != null)
                map.put(name, value);
        }
        return map;
    }

    private Object readTag(byte type) throws IOException {
        return switch (type) {
            case 1 -> in.readByte();
            case 2 -> in.readShort();
            case 3 -> in.readInt();
            case 4 -> in.readLong();
            case 5 -> in.readFloat();
            case 6 -> in.readDouble();
            case 7 -> {
                int len = in.readInt();
                byte[] b = new byte[len];
                in.readFully(b);
                yield b;
            }
            case 8 -> readString();
            case 9 -> {
                byte subType = in.readByte();
                int len = in.readInt();
                List<Object> list = new ArrayList<>(len);
                for (int i = 0; i < len; i++)
                    list.add(readTag(subType));
                yield list;
            }
            case 10 -> readCompound();
            case 11 -> {
                int len = in.readInt();
                int[] a = new int[len];
                for (int i = 0; i < len; i++)
                    a[i] = in.readInt();
                yield a;
            }
            case 12 -> {
                int len = in.readInt();
                long[] a = new long[len];
                for (int i = 0; i < len; i++)
                    a[i] = in.readLong();
                yield a;
            }
            default -> throw new IOException("Unknown Tag Type: " + type);
        };
    }

    private String readString() throws IOException {
        int len = in.readUnsignedShort();
        if (len == 0)
            return "";
        byte[] b = new byte[len];
        in.readFully(b);
        return new String(b, StandardCharsets.UTF_8);
    }

    public void close() throws IOException {
        in.close();
    }

    @SuppressWarnings("unchecked")
    public static <T> T getNested(Map<String, Object> map, String path) {
        if (map == null)
            return null;
        String[] parts = path.split("\\.");
        Object current = map;
        for (String part : parts) {
            if (!(current instanceof Map))
                return null;
            current = ((Map<String, Object>) current).get(part);
        }
        return (T) current;
    }

    public static long getAsLong(Object val, long def) {
        if (val instanceof Number n)
            return n.longValue();
        return def;
    }

    public static int getAsInt(Object val, int def) {
        if (val instanceof Number n)
            return n.intValue();
        return def;
    }

    public static boolean getAsBoolean(Object val, boolean def) {
        if (val instanceof Boolean b)
            return b;
        if (val instanceof Number n)
            return n.intValue() != 0;
        return def;
    }

    public static String getAsString(Object val, String def) {
        if (val instanceof String s)
            return s;
        return def;
    }
}
