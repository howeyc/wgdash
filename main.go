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
	Device struct {
		Dev     string `toml:"dev"`
		Display string `toml:"display"`
		Host    string `toml:"host"`
	} `toml:"device"`
	Peers []struct {
		Key     string `toml:"key"`
		Display string `toml:"display"`
		Host    string `toml:"host"`
	} `toml:"peer"`
}

func main() {
	var tomlpath string
	var outpath string
	flag.StringVar(&tomlpath, "config", "", "dashboard config file (*required)")
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
	} else {
		log.Fatalln("must specify config file")
	}

	wgdev, derr := GetDevInfo(config.Device.Dev)
	if derr != nil {
		log.Fatalln(derr)
	}

	wginfo, wgerr := GetWGInfo(config.Device.Dev)
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

	wginfo.Displayname = config.Device.Display
	wginfo.Hostname = config.Device.Host

	ofile, oerr := os.Create(outpath)
	if oerr != nil {
		log.Fatalln(oerr)
	}

	t := template.New("html")
	t, _ = t.Parse(tmpfs)
	t.Execute(ofile, wginfo)

	ofile.Close()
}
