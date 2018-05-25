package main

import (
	"fmt"
	"github.com/keitaro1020/slack-emojis/cmd"
	"os"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
}
