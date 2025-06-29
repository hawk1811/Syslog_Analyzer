 
package syslog

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

// SharedListener manages a single listener for multiple sources
type SharedListener struct {
	tcpListener net.Listener
	udpConn     *net.UDPConn
	protocol    string
	port        int
	sources     map[string]*SyslogSource // map[sourceIP] -> source
	sourceMutex sync.RWMutex
	stopChan    chan bool
	isRunning   bool
}

// NewSharedListener creates a new shared listener
func NewSharedListener(protocol string, port int) *SharedListener {
	return &SharedListener{
		protocol:  protocol,
		port:      port,
		sources:   make(map[string]*SyslogSource),
		stopChan:  make(chan bool),
	}
}

// Start starts the shared listener
func (sl *SharedListener) Start() error {
	address := fmt.Sprintf(":%d", sl.port)
	
	if sl.protocol == "TCP" {
		listener, err := net.Listen("tcp", address)
		if err != nil {
			return err
		}
		sl.tcpListener = listener
		go sl.handleTCPConnections()
	} else {
		udpAddr, err := net.ResolveUDPAddr("udp", address)
		if err != nil {
			return err
		}
		
		udpConn, err := net.ListenUDP("udp", udpAddr)
		if err != nil {
			return err
		}
		sl.udpConn = udpConn
		go sl.handleUDPConnections()
	}
	
	sl.isRunning = true
	return nil
}

// Stop stops the shared listener
func (sl *SharedListener) Stop() {
	if !sl.isRunning {
		return
	}
	
	sl.isRunning = false
	close(sl.stopChan)
	
	if sl.tcpListener != nil {
		sl.tcpListener.Close()
	}
	if sl.udpConn != nil {
		sl.udpConn.Close()
	}
}

// AddSource adds a source to this shared listener
func (sl *SharedListener) AddSource(source *SyslogSource) {
	sl.sourceMutex.Lock()
	defer sl.sourceMutex.Unlock()
	
	sourceKey := source.config.IP
	if sourceKey == "" {
		sourceKey = "0.0.0.0"
	}
	
	sl.sources[sourceKey] = source
}

// RemoveSource removes a source from this shared listener
func (sl *SharedListener) RemoveSource(source *SyslogSource) {
	sl.sourceMutex.Lock()
	defer sl.sourceMutex.Unlock()
	
	sourceKey := source.config.IP
	if sourceKey == "" {
		sourceKey = "0.0.0.0"
	}
	
	delete(sl.sources, sourceKey)
}

// GetSourceCount returns the number of sources using this listener
func (sl *SharedListener) GetSourceCount() int {
	sl.sourceMutex.RLock()
	defer sl.sourceMutex.RUnlock()
	return len(sl.sources)
}

// handleUDPConnections processes UDP messages for all sources
func (sl *SharedListener) handleUDPConnections() {
	buffer := make([]byte, 65536)
	
	for {
		select {
		case <-sl.stopChan:
			return
		default:
			if sl.udpConn == nil {
				continue
			}
			
			n, addr, err := sl.udpConn.ReadFromUDP(buffer)
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					return
				}
				continue
			}
			
			// Route message to appropriate sources
			sl.routeMessage(buffer[:n], addr.IP.String())
		}
	}
}

// handleTCPConnections processes TCP connections for all sources  
func (sl *SharedListener) handleTCPConnections() {
	for {
		select {
		case <-sl.stopChan:
			return
		default:
			if sl.tcpListener == nil {
				continue
			}
			
			conn, err := sl.tcpListener.Accept()
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					return
				}
				continue
			}
			
			go sl.handleTCPConnection(conn)
		}
	}
}

// handleTCPConnection processes a single TCP connection
func (sl *SharedListener) handleTCPConnection(conn net.Conn) {
	defer conn.Close()
	
	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
	sourceIP := remoteAddr.IP.String()
	
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		sl.routeMessage(scanner.Bytes(), sourceIP)
	}
}

// routeMessage routes messages to appropriate sources based on IP
func (sl *SharedListener) routeMessage(data []byte, sourceIP string) {
	sl.sourceMutex.RLock()
	defer sl.sourceMutex.RUnlock()
	
	// Try to find exact IP match first
	if source, exists := sl.sources[sourceIP]; exists && source.IsRunning() {
		source.ProcessMessage(data, sourceIP)
		return
	}
	
	// If no exact match, try wildcard (0.0.0.0) sources
	if source, exists := sl.sources["0.0.0.0"]; exists && source.IsRunning() {
		source.ProcessMessage(data, sourceIP)
		return
	}
}