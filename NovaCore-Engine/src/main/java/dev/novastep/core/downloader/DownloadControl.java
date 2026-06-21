package dev.novastep.core.downloader;

import java.util.concurrent.locks.Condition;
import java.util.concurrent.locks.ReentrantLock;

public final class DownloadControl {

    private final ReentrantLock lock      = new ReentrantLock();
    private final Condition     resumed   = lock.newCondition();

    private volatile boolean paused    = false;
    private volatile boolean cancelled = false;

    public void pause() {
        paused = true;
    }

    public void resume() {
        lock.lock();
        try {
            paused = false;
            resumed.signalAll();
        } finally {
            lock.unlock();
        }
    }

    public void cancel() {
        lock.lock();
        try {
            cancelled = true;
            paused    = false;
            resumed.signalAll();
        } finally {
            lock.unlock();
        }
    }

    public boolean isPaused()    { return paused && !cancelled; }
    public boolean isCancelled() { return cancelled; }

    public boolean checkPoint() throws InterruptedException {
        if (!paused) return !cancelled;

        lock.lock();
        try {
            while (paused && !cancelled) {
                resumed.await();
            }
        } finally {
            lock.unlock();
        }
        return !cancelled;
    }
}
