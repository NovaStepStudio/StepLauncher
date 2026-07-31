package logger

type broadcastFn func(Type, string)

func (l *Logger) SetBroadcastFn(fn broadcastFn) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.broadcastFn = fn
}

func (l *Logger) getBroadcastFn() broadcastFn {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.broadcastFn
}
