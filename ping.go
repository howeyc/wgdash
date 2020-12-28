package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func ping(ip string) (string, error) {
	durtime := "off"
	cmd := exec.Command("ping", "-c", "1", "-i", "1", "-W", "1", ip)
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()
	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		word := scanner.Text()
		if strings.HasPrefix(word, "time=") {
			tparts := strings.Split(word, "=")
			durtime = tparts[1]
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
	return durtime, cmd.Wait()
}
