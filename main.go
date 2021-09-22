package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

const path = "test"

var wait = make(chan bool)
var startRun = make(chan interface{})

func main() {
	watcher, _ := fsnotify.NewWatcher()
	defer watcher.Close()

	go run()

	go watch(watcher)

	err := watcher.Add(path)
	if err != nil {
		log.Printf("error: %v", err.Error())
	}
	<-wait
}

func shouldRun(path string, op fsnotify.Op) bool {
	base := filepath.Base(path)
	if strings.HasSuffix(base, "~") || op == fsnotify.Chmod {
		return false
	}
	return true
}

func flushEvents() {
	for {
		select {
		case eventName := <-startRun:
			log.Printf("receiving event %s", eventName)
		default:
			return
		}
	}
}

func run() {
	for {
		select {
		case <-startRun:
			cmd := exec.Command("go", "run", path+"/main.go")
			var stdOut bytes.Buffer
			cmd.Stdout = &stdOut
			err := cmd.Run()
			if err != nil {
				log.Fatalf("cmd.Run() failed with %s, cmd: %v\n", err, cmd)
			}
			fmt.Print(string(stdOut.Bytes()))
			time.Sleep(3 * time.Second)
			flushEvents()
		default:
		}
	}
}

func watch(watcher *fsnotify.Watcher) {
	defer close(wait)

	for {
		select {
		case ev, ok := <-watcher.Events:
			if !ok {
				return
			}
			if !shouldRun(ev.Name, ev.Op) {
				continue
			}
			startRun <- ev.String()
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("have a error: %v", err.Error())
		}
	}
}
