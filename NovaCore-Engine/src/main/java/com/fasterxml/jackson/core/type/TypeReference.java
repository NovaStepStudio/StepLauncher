package com.fasterxml.jackson.core.type;

import java.lang.reflect.ParameterizedType;
import java.lang.reflect.Type;

public abstract class TypeReference<T> {
    private final Type type;

    protected TypeReference() {
        Type superclass = getClass().getGenericSuperclass();
        if (superclass instanceof ParameterizedType parameterized) {
            this.type = parameterized.getActualTypeArguments()[0];
        } else {
            throw new IllegalStateException("TypeReference must be created with generic type information");
        }
    }

    public Type getType() {
        return type;
    }
}
