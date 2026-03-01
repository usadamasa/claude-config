package main

import (
	"fmt"
	"os"

	"github.com/usadamasa/claude-config/internal/pathutil"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "使い方: realpath <パス>\n")
		os.Exit(1)
	}

	result, err := pathutil.ResolveRealpath(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "パスの解決に失敗: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(result)
}
