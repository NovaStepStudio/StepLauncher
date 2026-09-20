package RichPresence

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

const DiscordClientID = "1438239391666405396"

const (
	opHandshake uint32 = 0
	opFrame     uint32 = 1
	opClose     uint32 = 2
	opPing      uint32 = 3
	opPong      uint32 = 4
)

const (
	waitDelay        = time.Second
	reconnectDelay   = 3 * time.Second
	dialTimeout      = 2 * time.Second
	writeTimeout     = 3 * time.Second
	handshakeTimeout = 5 * time.Second
	// maxDialAttempts es el número de fallos de conexión consecutivos tras el
	// que se suspenden los reintentos (Discord cerrado o sin cliente IPC).
	// Solo se reanuda al volver a activar la presencia (SetEnabled(true)).
	maxDialAttempts = 5
)

type Activity struct {
	State   string `json:"state,omitempty"`
	Details string `json:"details,omitempty"`
	Start   int64  `json:"start,omitempty"`
}

type frame struct {
	opcode uint32
	data   []byte
}

type Manager struct {
	mu        sync.Mutex
	enabled   bool
	started   bool
	closed    bool
	suspended bool // true tras maxDialAttempts fallos: no se reintenta hasta re-activar
	conn      net.Conn
	activity  *Activity
	logFn     func(format string, args ...interface{})

	quit chan struct{}
	wake chan struct{}
	wg   sync.WaitGroup
}

func NewManager() *Manager {
	return &Manager{
		quit: make(chan struct{}),
		wake: make(chan struct{}, 1),
	}
}

func (m *Manager) SetLogFn(fn func(format string, args ...interface{})) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logFn = fn
}

func (m *Manager) logf(format string, args ...interface{}) {
	if m.logFn != nil {
		m.logFn(format, args...)
	}
}

func (m *Manager) wakeLoop() {
	select {
	case m.wake <- struct{}{}:
	default:
	}
}

func (m *Manager) SetEnabled(v bool) {
	m.mu.Lock()
	changed := m.enabled != v
	m.enabled = v
	if v {
		// Una activación (inicio o toggle del usuario) levanta la suspensión
		// por fallos repetidos y reanuda los reintentos.
		m.suspended = false
	}
	if v && !m.started {
		m.started = true
		m.wg.Add(1)
		go m.loop()
	}
	m.mu.Unlock()
	if changed {
		if v {
			m.logf("[Discord] Presencia activada")
		} else {
			m.logf("[Discord] Presencia desactivada")
		}
	}
	if !v {
		m.clearOnWire()
	}
	m.wakeLoop()
}

func (m *Manager) SetActivity(details, state string, start int64) {
	var act *Activity
	if details != "" || state != "" {
		act = &Activity{State: state, Details: details, Start: start}
	}
	m.mu.Lock()
	m.activity = act
	conn := m.conn
	enabled := m.enabled
	m.mu.Unlock()
	if conn == nil || !enabled {
		return
	}
	if err := m.writeFrame(conn, opFrame, buildSetActivity(act)); err != nil {
		m.logf("[Discord] Error al enviar actividad: %v", err)
		conn.Close()
	}
}

func (m *Manager) loop() {
	defer m.wg.Done()
	attempts := 0
	for {
		select {
		case <-m.quit:
			return
		default:
		}
		if !m.isEnabled() {
			select {
			case <-m.quit:
				return
			case <-m.wake:
			case <-time.After(waitDelay):
			}
			continue
		}
		conn, err := dialDiscord()
		if err != nil {
			attempts++
			if attempts == 1 || attempts%10 == 0 {
				m.logf("[Discord] Cliente no disponible (intento %d): %v", attempts, err)
			}
			if attempts >= maxDialAttempts {
				m.mu.Lock()
				m.suspended = true
				m.mu.Unlock()
				m.logf("[Discord] Cliente no disponible tras %d intentos: se detienen los reintentos hasta volver a activar la presencia", attempts)
				for {
					select {
					case <-m.quit:
						return
					default:
					}
					m.mu.Lock()
					suspended := m.suspended
					enabled := m.enabled
					m.mu.Unlock()
					if !suspended || !enabled {
						break
					}
					select {
					case <-m.quit:
						return
					case <-m.wake:
					}
				}
				attempts = 0
				continue
			}
			select {
			case <-m.quit:
				return
			case <-m.wake:
			case <-time.After(reconnectDelay):
			}
			continue
		}
		attempts = 0
		m.logf("[Discord] Conectado a Discord IPC")
		if m.runConnection(conn) {
			return
		}
	}
}

func (m *Manager) isEnabled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.enabled
}

func (m *Manager) runConnection(conn net.Conn) bool {
	payload, _ := json.Marshal(map[string]interface{}{
		"v":         1,
		"client_id": DiscordClientID,
	})
	if err := conn.SetDeadline(time.Now().Add(handshakeTimeout)); err != nil {
		conn.Close()
		return false
	}
	if err := m.writeFrame(conn, opHandshake, payload); err != nil {
		conn.Close()
		return false
	}
	resp, err := readFrame(conn)
	if err != nil {
		conn.Close()
		return false
	}
	if resp.opcode != opFrame {
		m.logf("[Discord] Handshake rechazado (opcode=%d)", resp.opcode)
		conn.Close()
		return false
	}
	if err := conn.SetDeadline(time.Time{}); err != nil {
		conn.Close()
		return false
	}
	m.mu.Lock()
	m.conn = conn
	m.mu.Unlock()

	m.sendCurrent()

	connDone := make(chan struct{})
	go m.readLoop(conn, connDone)
	select {
	case <-m.quit:
		conn.Close()
		m.detachConn(conn)
		return true
	case <-connDone:
		m.detachConn(conn)
		return false
	}
}

func (m *Manager) readLoop(conn net.Conn, done chan<- struct{}) {
	defer close(done)
	for {
		f, err := readFrame(conn)
		if err != nil {
			conn.Close()
			return
		}
		switch f.opcode {
		case opPing:
			if err := m.writeFrame(conn, opPong, f.data); err != nil {
				conn.Close()
				return
			}
		case opClose:
			conn.Close()
			return
		default:
		}
	}
}

func (m *Manager) detachConn(conn net.Conn) {
	m.mu.Lock()
	if m.conn == conn {
		m.conn = nil
	}
	m.mu.Unlock()
}

func (m *Manager) sendCurrent() {
	m.mu.Lock()
	act := m.activity
	conn := m.conn
	m.mu.Unlock()
	if conn == nil {
		return
	}
	if err := m.writeFrame(conn, opFrame, buildSetActivity(act)); err != nil {
		m.logf("[Discord] Error al reenviar actividad: %v", err)
		conn.Close()
	}
}

func (m *Manager) clearOnWire() {
	m.mu.Lock()
	conn := m.conn
	m.mu.Unlock()
	if conn == nil {
		return
	}
	if err := m.writeFrame(conn, opFrame, buildSetActivity(nil)); err != nil {
		m.logf("[Discord] Error al limpiar presencia: %v", err)
	}
	conn.Close()
}

func (m *Manager) writeFrame(conn net.Conn, opcode uint32, payload []byte) error {
	if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		return err
	}
	buf := make([]byte, 8+len(payload))
	binary.LittleEndian.PutUint32(buf[0:4], opcode)
	binary.LittleEndian.PutUint32(buf[4:8], uint32(len(payload)))
	copy(buf[8:], payload)
	_, err := conn.Write(buf)
	return err
}

func readFrame(conn net.Conn) (frame, error) {
	var header [8]byte
	if _, err := io.ReadFull(conn, header[:]); err != nil {
		return frame{}, err
	}
	opcode := binary.LittleEndian.Uint32(header[0:4])
	length := binary.LittleEndian.Uint32(header[4:8])
	if length > 16*1024*1024 {
		return frame{}, fmt.Errorf("frame demasiado grande (%d bytes)", length)
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(conn, payload); err != nil {
		return frame{}, err
	}
	return frame{opcode: opcode, data: payload}, nil
}

func buildSetActivity(act *Activity) []byte {
	activity := map[string]interface{}{}
	if act != nil {
		if act.State != "" {
			activity["state"] = act.State
		}
		if act.Details != "" {
			activity["details"] = act.Details
		}
		if act.Start > 0 {
			activity["timestamps"] = map[string]int64{"start": act.Start}
		}
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"cmd": "SET_ACTIVITY",
		"args": map[string]interface{}{
			"pid":      os.Getpid(),
			"activity": activity,
		},
		"nonce": strconv.FormatInt(time.Now().UnixNano(), 36),
	})
	return payload
}

func (m *Manager) Close() {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return
	}
	m.closed = true
	conn := m.conn
	m.mu.Unlock()
	close(m.quit)
	if conn != nil {
		conn.Close()
	}
	m.wg.Wait()
}
