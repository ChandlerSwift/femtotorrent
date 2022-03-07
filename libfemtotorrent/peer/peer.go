package peer

import (
	"bytes"
	"fmt"
	"log"
	"net"

	"github.com/chandlerswift/femtotorrent/libfemtotorrent/torrentfile"
)

type Peer struct {
	IPAddress  net.IP
	Port       uint16
	Choked     bool
	Interested bool
	conn       net.Conn
	ID         []byte
}

func (p *Peer) Handle(tf torrentfile.TorrentFile) (err error) {
	err = p.connect(tf)
	if err != nil {
		return
	}
	p.DeclareInterested()
	for {
		p.DeclareInterested()
		buf := make([]byte, 65536)
		var n int
		n, err = p.conn.Read(buf)
		if err != nil {
			return
		}
		log.Printf("Received %v bytes: %q", n, buf[:n])
	}
}

func (p *Peer) connect(tf torrentfile.TorrentFile) (err error) {
	// Connections start out choked and not interested.
	p.Choked = true
	p.Interested = false

	// TODO: can we also do UDP?
	fullHost := net.JoinHostPort(p.IPAddress.String(), fmt.Sprint(p.Port))
	log.Println(fullHost)
	p.conn, err = net.Dial("tcp", fullHost)
	if err != nil {
		return
	}

	// The peer wire protocol consists of a handshake followed by a never-ending
	// stream of length-prefixed messages. The handshake starts with character
	// ninteen (decimal) followed by the string 'BitTorrent protocol'. The
	// leading character is a length prefix, put there in the hope that other
	// new protocols may do the same and thus be trivially distinguishable from
	// each other.
	p.conn.Write([]byte{19})
	p.conn.Write([]byte("BitTorrent protocol"))

	// After the fixed headers come eight reserved bytes, which are all zero in
	// all current implementations. If you wish to extend the protocol using
	// these bytes, please coordinate with Bram Cohen to make sure all
	// extensions are done compatibly.
	p.conn.Write([]byte{0, 0, 0, 0, 0, 0, 0, 0})

	// Next comes the 20 byte sha1 hash of the bencoded form of the info value
	// from the metainfo file. (This is the same value which is announced as
	// info_hash to the tracker, only here it's raw instead of quoted here). If
	// both sides don't send the same value, they sever the connection. The one
	// possible exception is if a downloader wants to do multiple downloads over
	// a single port, they may wait for incoming connections to give a download
	// hash first, and respond with the same one if it's in their list.
	p.conn.Write(tf.InfoHash[:])

	// After the download hash comes the 20-byte peer id which is reported in
	// tracker requests and contained in peer lists in tracker responses. If the
	// receiving side's peer id doesn't match the one the initiating side
	// expects, it severs the connection.
	p.conn.Write([]byte("chandlerswiftdebian!"))

	buf := make([]byte, 65536)
	var n int
	n, err = p.conn.Read(buf)
	if err != nil {
		return
	}
	log.Printf("Received %d bytes: %q", n, buf[:n])
	if !bytes.Equal([]byte("\x13BitTorrent protocol"), buf[:20]) {
		return fmt.Errorf("Protocol mismatch; expected \"\\x13BitTorrent protocol\", got %q", buf[:20])
	}
	if !bytes.Equal(tf.InfoHash[:], buf[28:48]) {
		return fmt.Errorf("InfoHash mismatch; expected %q, got %q", tf.InfoHash, buf[28:48])
	}
	// TODO: peer ID

	if n > 68 {
		// we got more bytes than we know what to do with
		return fmt.Errorf("Too many bytes! Expected 68, got %v", n)
	}

	return
}

func (p *Peer) Choke() {
	p.conn.Write([]byte{0})
}

func (p *Peer) Unchoke() {
	p.conn.Write([]byte{1})
}

func (p *Peer) DeclareInterested() {
	p.conn.Write([]byte{2})
}

func (p *Peer) DeclareNotInterested() {
	p.conn.Write([]byte{3})
}

func (p *Peer) Have() {
	p.conn.Write([]byte{4})
	panic("unimplemented") // TODO
}

func (p *Peer) Bitfield() {
	p.conn.Write([]byte{5})
	panic("unimplemented") // TODO
}

func (p *Peer) Request() {
	p.conn.Write([]byte{6})
	panic("unimplemented") // TODO
}

func (p *Peer) Piece() {
	p.conn.Write([]byte{7})
	panic("unimplemented") // TODO
}

func (p *Peer) Cancel() {
	p.conn.Write([]byte{8})
	panic("unimplemented") // TODO
}
