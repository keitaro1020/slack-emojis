package main

import (
	"github.com/keitaro1020/slack-emojis/cmd"
	"os"
	"fmt"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
}
