package main

import (
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/chandlerswift/femtotorrent/libfemtotorrent/peer"
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
	localPeer := peer.Peer{
		IPAddress: net.ParseIP("127.0.0.1"),
		Port:      51413,
	}

	log.Printf("Writing to %v", tf.Info.Name)
	f, err := os.OpenFile(tf.Info.Name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = localPeer.Handle(tf, f)
	if err != nil {
		panic(err)
	}

}
