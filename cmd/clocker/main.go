package main

import (
	"flag"
	"fmt"
	"local/clocker/filewatcher"
)

const helpMsg = `
Valid flags:
  --file
    Path to a file on the local FS, which will be watched for changes and used as input.
`

func main() {
	filePath := flag.String("file", "", "Path to a local file to watch for changes")
	flag.Parse()

	if *filePath != "" {
		filewatcher.Watch(*filePath)
		return
	}
	fmt.Println("No flags detected.")
	fmt.Print(helpMsg)
}
