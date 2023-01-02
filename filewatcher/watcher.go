package filewatcher

import (
	"fmt"
	"local/clocker"
	"log"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
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

	b, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("os.ReadFile: %w", err)
	}

	fmt.Print(lp.Summary(string(b)))
	return nil
}
