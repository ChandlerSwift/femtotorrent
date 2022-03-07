package peer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"

	"github.com/chandlerswift/femtotorrent/libfemtotorrent/torrentfile"
)

type Peer struct {
	IPAddress      net.IP
	Port           uint16
	OutgoingChoked bool
	IncomingChoked bool
	Interested     bool // TODO: for both sides of the connection
	conn           net.Conn
	ID             []byte
}

func (p *Peer) Handle(tf torrentfile.TorrentFile) (err error) {
	err = p.connect(tf)
	if err != nil {
		return
	}
	currentPiece := uint32(0)
	currentOffset := uint32(0)
	isOutstandingPiece := false
	currentPieceBuf := bytes.Buffer{}
	for {
		if !p.IncomingChoked && !p.Interested {
			log.Println("Declaring our interest")
			p.DeclareInterested()
		}
		if !p.IncomingChoked && !isOutstandingPiece {
			remainingSize := tf.Info.PieceLength - currentPieceBuf.Len()
			var reqLen uint32
			if remainingSize < 2<<13 {
				reqLen = uint32(remainingSize)
			} else {
				reqLen = 2 << 13
			}
			log.Printf("Requesting chunk %v@%v (%v bytes)", currentPiece, currentOffset, reqLen)
			p.Request(currentPiece, currentOffset, reqLen)
			isOutstandingPiece = true
		}
		buf := make([]byte, 65536) // TODO: don't allocate each time around
		var n int
		n, err = p.conn.Read(buf)
		if err != nil {
			return
		}
		switch buf[4] { // message type
		case 0:
			log.Println("Received an incoming choke message")
			p.IncomingChoked = true
		case 1:
			log.Println("Received an incoming unchoke message")
			p.IncomingChoked = false
		case 2:
			log.Println("Received an incoming interested message")
		case 3:
			log.Println("Received an incoming uninterested message")
		case 4:
			var piece uint64
			piece, err = binary.ReadUvarint(bytes.NewBuffer(buf[5:9]))
			if err != nil {
				return
			}
			log.Printf("Received a have message for piece %v", piece)
		case 7: // piece
			if n != 4+1+4+4+2<<13 {
				return fmt.Errorf("Unexpected piece message size %v", n)
			}
			var index, begin uint64
			index, err = binary.ReadUvarint(bytes.NewBuffer(buf[5:9]))
			if err != nil {
				return
			}
			if uint32(index) != currentPiece {
				return fmt.Errorf("Expecting piece %v, got piece %v", currentPiece, index)
			}
			begin, err = binary.ReadUvarint(bytes.NewBuffer(buf[9:13]))
			if err != nil {
				return
			}
			if uint32(begin) != currentOffset {
				return fmt.Errorf("Expecting offset %v, got offset %v", currentOffset, begin)
			}
			log.Printf("Received %v bytes of %v@%v", n-4-1-4-4, index, begin)
			currentPieceBuf.Write(buf[13:n])
			if currentPieceBuf.Len() == tf.Info.PieceLength {
				// TODO: check hashes
				p.Have(uint32(index))
				// advance to the next piece
				currentPiece++
				currentOffset = 0
				// TODO: flush to disk
				if int(currentPiece) == len(tf.Info.Pieces) {
					log.Println("Download complete!")
					return nil
				}
			}
			isOutstandingPiece = false
		default:
			log.Printf("In loop, received %v bytes: %q", n, buf[:n])
		}
	}
}

func (p *Peer) connect(tf torrentfile.TorrentFile) (err error) {
	// Connections start out choked and not interested.
	p.IncomingChoked = true
	p.OutgoingChoked = true
	p.Interested = false

	// TODO: can we also do UDP?
	p.conn, err = net.Dial("tcp", net.JoinHostPort(p.IPAddress.String(), fmt.Sprint(p.Port)))
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
	if !bytes.Equal([]byte("\x13BitTorrent protocol"), buf[:20]) {
		return fmt.Errorf("Protocol mismatch; expected \"\\x13BitTorrent protocol\", got %q", buf[:20])
	}
	if !bytes.Equal(tf.InfoHash[:], buf[28:48]) {
		return fmt.Errorf("InfoHash mismatch; expected %q, got %q", tf.InfoHash, buf[28:48])
	}
	// TODO: validate against what we expect
	p.ID = buf[48:68]

	if n > 68 {
		// we got more bytes than we know what to do with
		return fmt.Errorf("Too many bytes! Expected 68, got %v", n)
	}

	return
}

func (p *Peer) Choke() {
	p.conn.Write([]byte{0, 0, 0, 1, 0})
}

func (p *Peer) Unchoke() {
	p.conn.Write([]byte{0, 0, 0, 1, 1})
}

func (p *Peer) DeclareInterested() {
	p.conn.Write([]byte{0, 0, 0, 1, 2})
	p.Interested = true
}

func (p *Peer) DeclareNotInterested() {
	p.conn.Write([]byte{0, 0, 0, 1, 3})
}

func (p *Peer) Have(index uint32) error {
	msg := struct {
		LengthPrefix uint32
		MessageType  uint8
		Index        uint32
	}{
		LengthPrefix: 1 + 4,
		MessageType:  4,
		Index:        index,
	}

	return binary.Write(p.conn, binary.BigEndian, msg)
}

func (p *Peer) Bitfield() {
	panic("unimplemented") // TODO
	p.conn.Write([]byte{5})
}

// 'request' messages contain an index, begin, and length. The last two are byte
// offsets. Length is generally a power of two unless it gets truncated by the
// end of the file. All current implementations use 2^14 (16 kiB), and close
// connections which request an amount greater than that.
func (p *Peer) Request(index, begin, length uint32) error {
	msg := struct {
		LengthPrefix uint32
		MessageType  uint8
		Index        uint32
		Begin        uint32
		Length       uint32
	}{
		LengthPrefix: 1 + 4 + 4 + 4,
		MessageType:  6,
		Index:        index,
		Begin:        begin,
		Length:       length,
	}

	return binary.Write(p.conn, binary.BigEndian, msg)
}

func (p *Peer) Piece() {
	p.conn.Write([]byte{7})
	panic("unimplemented") // TODO
}

func (p *Peer) Cancel() {
	p.conn.Write([]byte{8})
	panic("unimplemented") // TODO
}
