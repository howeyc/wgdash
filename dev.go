package main

import (
	"encoding/json"
	"errors"
	"os/exec"
)

type DevInfo struct {
	Ifindex   int      `json:"ifindex"`
	Ifname    string   `json:"ifname"`
	Flags     []string `json:"flags"`
	Mtu       int      `json:"mtu"`
	Qdisc     string   `json:"qdisc"`
	Operstate string   `json:"operstate"`
	Group     string   `json:"group"`
	Txqlen    int      `json:"txqlen"`
	LinkType  string   `json:"link_type"`
	Address   string   `json:"address,omitempty"`
	Broadcast string   `json:"broadcast,omitempty"`
	AddrInfo  []struct {
		Family            string `json:"family"`
		Local             string `json:"local"`
		Prefixlen         int    `json:"prefixlen"`
		Scope             string `json:"scope"`
		Label             string `json:"label,omitempty"`
		ValidLifeTime     int64  `json:"valid_life_time"`
		PreferredLifeTime int64  `json:"preferred_life_time"`
	} `json:"addr_info"`
	Stats64 struct {
		Rx struct {
			Bytes      int `json:"bytes"`
			Packets    int `json:"packets"`
			Errors     int `json:"errors"`
			Dropped    int `json:"dropped"`
			OverErrors int `json:"over_errors"`
			Multicast  int `json:"multicast"`
		} `json:"rx"`
		Tx struct {
			Bytes         int `json:"bytes"`
			Packets       int `json:"packets"`
			Errors        int `json:"errors"`
			Dropped       int `json:"dropped"`
			CarrierErrors int `json:"carrier_errors"`
			Collisions    int `json:"collisions"`
		} `json:"tx"`
	} `json:"stats64"`
}

func GetDevInfo(dev string) (DevInfo, error) {
	var devs []DevInfo
	cmd := exec.Command("ip", "-s", "-j", "address")
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()
	json.NewDecoder(stdout).Decode(&devs)
	for _, d := range devs {
		if d.Ifname == dev {
			return d, cmd.Wait()
		}
	}
	return DevInfo{}, errors.New("device not found")
}
