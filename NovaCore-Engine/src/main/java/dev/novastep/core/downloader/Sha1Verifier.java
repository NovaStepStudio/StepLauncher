package dev.novastep.core.downloader;

import java.io.BufferedInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;

public class Sha1Verifier {

    private static final int BUFFER_SIZE = 64 * 1024; 
    private Sha1Verifier() {}

    public static MessageDigest newDigest() {
        try {
            return MessageDigest.getInstance("SHA-1");
        } catch (NoSuchAlgorithmException e) {
            throw new RuntimeException("SHA-1 not available", e);
        }
    }

    public static String toHex(byte[] bytes) {
        char[] hex = new char[bytes.length * 2];
        for (int i = 0; i < bytes.length; i++) {
            int b = bytes[i] & 0xFF;
            hex[i * 2] = "0123456789abcdef".charAt(b >>> 4);
            hex[i * 2 + 1] = "0123456789abcdef".charAt(b & 0x0F);
        }
        return new String(hex);
    }

    public static String finalize(MessageDigest digest) {
        return toHex(digest.digest());
    }

    public static boolean matches(String computed, String expected) {
        if (expected == null || expected.isBlank()) return true; 
        return computed.equalsIgnoreCase(expected);
    }

    public static String computeFile(Path file) throws IOException {
        MessageDigest digest = newDigest();
        byte[] buf = new byte[BUFFER_SIZE];

        try (InputStream is = new BufferedInputStream(Files.newInputStream(file), BUFFER_SIZE)) {
            int read;
            while ((read = is.read(buf)) != -1) {
                digest.update(buf, 0, read);
            }
        }

        return finalize(digest);
    }

    public static boolean verifyFile(Path file, String expected) throws IOException {
        if (expected == null || expected.isBlank()) return true;
        if (!Files.exists(file)) return false;

        String computed = computeFile(file);
        return matches(computed, expected);
    }
}