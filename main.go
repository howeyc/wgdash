package main

import (
	_ "embed"
	"flag"
	"html/template"
	"log"
	"os"
	"sync"

	"github.com/pelletier/go-toml"
)

//go:embed template.html
var tmpfs string

var config struct {
	Peers []struct {
		Key     string `toml:"key"`
		Display string `toml:"display"`
		Host    string `toml:"host"`
	} `toml:"peer"`
}

func main() {
	var wgdevname string
	var tomlpath string
	var outpath string
	flag.StringVar(&wgdevname, "dev", "wg0", "wireguard device")
	flag.StringVar(&tomlpath, "config", "", "dashboard config file")
	flag.StringVar(&outpath, "o", "", "output html file path")
	flag.Parse()

	if tomlpath != "" {
		ifile, ierr := os.Open(tomlpath)
		if ierr != nil {
			log.Fatalln(ierr)
		}
		tdec := toml.NewDecoder(ifile)
		terr := tdec.Decode(&config)
		ifile.Close()
		if terr != nil {
			log.Fatalln(terr)
		}
	}

	wgdev, derr := GetDevInfo(wgdevname)
	if derr != nil {
		log.Fatalln(derr)
	}

	wginfo, wgerr := GetWGInfo(wgdevname)
	if wgerr != nil {
		log.Fatalln(wgerr)
	}

	for _, ai := range wgdev.AddrInfo {
		wginfo.IPs = append([]string{ai.Local}, wginfo.IPs...)
	}
	wginfo.TransferRx = float64(wgdev.Stats64.Rx.Bytes) / (1024.0 * 1024.0)
	wginfo.TransferTx = float64(wgdev.Stats64.Tx.Bytes) / (1024.0 * 1024.0)

	var wg sync.WaitGroup
	for pi, _ := range wginfo.Peers {
		wg.Add(1)
		go func(p *WGPeer) {
			for ii, ip := range p.AllowedIPs {
				dur, err := ping(ip)
				if err == nil {
					p.Online[ii] = true
				}
				p.Duration[ii] = dur
			}
			wg.Done()
		}(&wginfo.Peers[pi])
		for _, cp := range config.Peers {
			if cp.Key == wginfo.Peers[pi].PublicKey {
				wginfo.Peers[pi].Hostname = cp.Host
				wginfo.Peers[pi].Displayname = cp.Display
			}
		}
	}
	wg.Wait()

	ofile, oerr := os.Create(outpath)
	if oerr != nil {
		log.Fatalln(oerr)
	}

	t := template.New("html")
	t, _ = t.Parse(tmpfs)
	t.Execute(ofile, wginfo)

	ofile.Close()
}
