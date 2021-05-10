package main

import (
	"bufio"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type WGPeer struct {
	Displayname string
	Hostname    string

	PublicKey              string
	PreSharedKey           string
	EndPoint               string
	AllowedIPs             []string
	LatestHandshake        time.Time
	TransferRx, TransferTx float64
	PersistentKeepalive    bool
	Online                 []bool

	Duration []string
}

type WGInfo struct {
	privateKey string
	PublicKey  string
	ListenPort int64
	FWMark     bool
	Peers      []WGPeer
	CheckTime  time.Time

	Displayname string
	Hostname    string

	IPs                    []string
	TransferRx, TransferTx float64
}

func GetWGInfo(dev string) (WGInfo, error) {
	var wgi WGInfo
	cmd := exec.Command("/usr/bin/wg", "show", dev, "dump")
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()
	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanWords)

	scanner.Scan()
	wgi.privateKey = scanner.Text()
	scanner.Scan()
	wgi.PublicKey = scanner.Text()
	scanner.Scan()
	wgi.ListenPort, _ = strconv.ParseInt(scanner.Text(), 10, 64)
	scanner.Scan()
	if scanner.Text() == "on" {
		wgi.FWMark = true
	}
	for scanner.Scan() {
		var peer WGPeer
		peer.PublicKey = scanner.Text()
		scanner.Scan()
		peer.PreSharedKey = scanner.Text()
		scanner.Scan()
		peer.EndPoint = scanner.Text()
		scanner.Scan()
		aips := scanner.Text()
		ips := strings.Split(aips, ",")
		for _, ip := range ips {
			ipparts := strings.Split(ip, "/")
			peer.AllowedIPs = append(peer.AllowedIPs, ipparts[0])
		}
		scanner.Scan()
		utime, _ := strconv.ParseInt(scanner.Text(), 10, 64)
		peer.LatestHandshake = time.Unix(utime, 0)
		scanner.Scan()
		rnum, _ := strconv.ParseInt(scanner.Text(), 10, 64)
		peer.TransferRx = float64(rnum) / (1024.0 * 1024.0)
		scanner.Scan()
		tnum, _ := strconv.ParseInt(scanner.Text(), 10, 64)
		peer.TransferTx = float64(tnum) / (1024.0 * 1024.0)
		scanner.Scan()
		if scanner.Text() == "on" {
			peer.PersistentKeepalive = true
		}
		peer.Online = make([]bool, len(peer.AllowedIPs))
		peer.Duration = make([]string, len(peer.AllowedIPs))

		peer.Displayname = peer.PublicKey[:10]

		wgi.Peers = append(wgi.Peers, peer)
	}
	cmd.Wait()
	wgi.CheckTime = time.Now()
	return wgi, scanner.Err()
}
