package logfile

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sporadisk/clocker/logentry"
)

type Subscriber struct {
	filePath string
	lastRead time.Time
	mu       sync.Mutex
	receiver logentry.Receiver
}

func NewSubscriber(filePath string) (*Subscriber, error) {
	return &Subscriber{filePath: filePath}, nil
}

func (s *Subscriber) Subscribe(receiver logentry.Receiver) error {
	s.receiver = receiver
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("fsnotify.NewWatcher: " + err.Error())
	}
	defer watcher.Close()

	go s.watchResponder(watcher)

	err = watcher.Add(s.filePath)
	if err != nil {
		log.Fatal("watcher.Add: " + err.Error())
	}

	// Block main goroutine forever.
	// TODO: implement proper shutdown handling
	<-make(chan struct{})
	return nil
}

func (s *Subscriber) watchResponder(watcher *fsnotify.Watcher) {

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				log.Println("watcher.Events is not okay.")
				return
			}
			if event.Has(fsnotify.Write) {
				err := s.reactToFileWrite(event.Name)
				if err != nil {
					log.Printf("reactToFileWrite: %s", err.Error())
					return
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				log.Println("watcher.Errors is not okay.")
				return
			}
			log.Println("watcher.Errors: ", err)
		}
	}
}

func (s *Subscriber) reactToFileWrite(filepath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	timeElapsed := time.Since(s.lastRead)
	if timeElapsed < time.Second { // react at most once per second
		return nil
	}
	s.lastRead = time.Now()

	lp := LogParser{}
	err := lp.Init()
	if err != nil {
		return fmt.Errorf("lp.Init: %w", err)
	}

	b, err := readLoop(filepath)
	if err != nil {
		return fmt.Errorf("readLoop: %w", err)
	}

	entries := lp.Parse(string(b))
	err = s.receiver.Receive(entries)

	if err != nil {
		return fmt.Errorf("error from log entry receiver: %w", err)
	}

	return nil
}

// readLoop tries to read the file a lot
func readLoop(filepath string) ([]byte, error) {
	for i := 0; i < 100; i++ {
		f, err := os.Open(filepath)
		if err != nil {
			return nil, fmt.Errorf("os.Open: %w", err)
		}
		defer f.Close()

		b, err := io.ReadAll(f)
		if err != nil {
			return nil, fmt.Errorf("io.ReadAll: %w", err)
		}

		if b == nil {
			return nil, fmt.Errorf("io.ReadAll returned nil")
		}

		if len(b) == 0 {
			// sometimes we get an empty file, probably because the file is being written to
			time.Sleep(time.Millisecond * 100)
			continue
		}

		// success!
		if i > 0 {
			// uncomment for fun debug output
			// fmt.Printf("readLoop took %d tries to succeed\n", i+1)
		}
		return b, nil
	}

	return nil, fmt.Errorf("readLoop: too many retries")
}
