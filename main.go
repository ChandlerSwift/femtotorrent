package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/chandlerswift/femtotorrent/libfemtotorrent/torrentfile"
	"github.com/chandlerswift/femtotorrent/libfemtotorrent/tracker"
)

func main() {
	file, err := os.Open("debian-11.2.0-amd64-netinst.iso.torrent")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)

	tf, err := torrentfile.DecodeTorrentFile(b)
	if err != nil {
		panic(err)
	}
	peers, interval, err := tracker.GetPeers(tf)
	if err != nil {
		panic(err)
	}
	log.Println(peers, interval)
	for _, peer := range peers {
		log.Printf("%v:%d", string(peer.IPAddress), peer.Port)
	}

}
