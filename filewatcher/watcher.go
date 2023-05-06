package filewatcher

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sporadisk/clocker"
)

func Watch(filePath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("fsnotify.NewWatcher: " + err.Error())
	}
	defer watcher.Close()

	su := &surveyer{}
	go su.watchResponder(watcher)

	err = watcher.Add(filePath)
	if err != nil {
		log.Fatal("watcher.Add: " + err.Error())
	}

	// Block main goroutine forever.
	<-make(chan struct{})
}

type surveyer struct {
	lastRead time.Time
	mu       sync.Mutex
}

func (su *surveyer) watchResponder(watcher *fsnotify.Watcher) {

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				log.Println("watcher.Events is not okay.")
				return
			}
			if event.Has(fsnotify.Write) {
				err := su.reactToFileWrite(event.Name)
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

func (su *surveyer) reactToFileWrite(filepath string) error {
	su.mu.Lock()
	defer su.mu.Unlock()

	timeElapsed := time.Since(su.lastRead)
	if timeElapsed < time.Second { // react at most once per second
		return nil
	}
	su.lastRead = time.Now()

	lp := clocker.LogParser{}
	err := lp.Init("")
	if err != nil {
		return fmt.Errorf("lp.Init: %w", err)
	}

	b, err := readLoop(filepath)
	if err != nil {
		return fmt.Errorf("readLoop: %w", err)
	}

	summary, err := lp.Summary(string(b))
	fmt.Print(summary)

	if err != nil {
		if err == clocker.ErrInvalidInput {
			fmt.Printf("Input: %#v\n", b)
		}
		return fmt.Errorf("lp.Summary: %w", err)
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
