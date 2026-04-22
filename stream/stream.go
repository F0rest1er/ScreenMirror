package stream

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/kbinani/screenshot"
)

type Streamer struct {
	clients      map[chan []byte]bool
	mutex        sync.Mutex
	displayIndex int
}

func NewStreamer() *Streamer {
	s := &Streamer{
		clients: make(map[chan []byte]bool),
	}
	go s.captureLoop()
	return s
}

func (s *Streamer) addClient(c chan []byte) {
	s.mutex.Lock()
	s.clients[c] = true
	s.mutex.Unlock()
}

func (s *Streamer) removeClient(c chan []byte) {
	s.mutex.Lock()
	delete(s.clients, c)
	s.mutex.Unlock()
}

func (s *Streamer) SetDisplay(idx int) {
	s.mutex.Lock()
	if idx >= 0 && idx < screenshot.NumActiveDisplays() {
		s.displayIndex = idx
	}
	s.mutex.Unlock()
}

func (s *Streamer) GetDisplay() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.displayIndex
}

func (s *Streamer) GetIndexAndBounds() (int, image.Rectangle) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	idx := s.displayIndex
	if idx >= screenshot.NumActiveDisplays() {
		idx = 0
	}
	return idx, screenshot.GetDisplayBounds(idx)
}

func (s *Streamer) captureLoop() {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		s.mutex.Lock()
		clientCount := len(s.clients)
		s.mutex.Unlock()

		if clientCount == 0 {
			continue
		}

		if !HasScreenCaptureAccess() {
			log.Println("Ожидание прав на запись экрана...")
			RequestScreenCaptureAccess()
			time.Sleep(5 * time.Second)
			continue
		}

		s.mutex.Lock()
		idx := s.displayIndex
		if idx >= screenshot.NumActiveDisplays() {
			idx = 0
			s.displayIndex = 0
		}
		s.mutex.Unlock()

		bounds := screenshot.GetDisplayBounds(idx)
		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			log.Printf("ОШИБКА: Не удалось захватить экран. Проверьте, выданы ли приложению права: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		var buf bytes.Buffer
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 70})
		if err != nil {
			continue
		}

		frameData := buf.Bytes()
		
		s.mutex.Lock()
		for client := range s.clients {
			select {
			case client <- frameData:
			default:
			}
		}
		s.mutex.Unlock()
	}
}

func (s *Streamer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clientChan := make(chan []byte, 10)
	s.addClient(clientChan)
	defer s.removeClient(clientChan)

	boundary := "mjpegframe"
	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary="+boundary)
	w.Header().Set("Connection", "close")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	for {
		select {
		case <-r.Context().Done():
			return
		case frame := <-clientChan:
			w.Write([]byte("--" + boundary + "\r\n"))
			w.Write([]byte("Content-Type: image/jpeg\r\n"))
			fmt.Fprintf(w, "Content-Length: %d\r\n\r\n", len(frame))
			w.Write(frame)
			w.Write([]byte("\r\n"))
			
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}
