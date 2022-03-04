package torrentfile

import (
	"crypto/sha1"
	"fmt"
	"time"

	"github.com/chandlerswift/femtotorrent/libfemtotorrent/bencoding"
)

type TorrentFile struct {
	Announce     string
	Comment      string
	CreationDate time.Time
	HTTPSeeds    []string
	Info         TorrentFileInfo
	InfoHash     [20]byte
}

type TorrentFileInfo struct {
	Length int    // Either this or Files is present
	Files  []File // Either this or Length is present
	Name   string
	Pieces [][]byte
}

type File struct {
	Path   string
	Length int
}

func DecodeTorrentFile(data []byte) (tf TorrentFile, err error) {
	raw, err := bencoding.Decode(data)
	if err != nil {
		return
	}
	rawDict, ok := raw.(map[string]interface{})
	if !ok {
		return tf, fmt.Errorf("Torrent (type %T) could not be decoded as map[string]interface{}", raw)
	}
	announce, ok := rawDict["announce"].([]byte)
	if !ok {
		return tf, fmt.Errorf("announce property not found in rawDict")
	}
	tf.Announce = string(announce)

	// Not mandated by the spec
	if comment, ok := rawDict["comment"].([]byte); ok {
		tf.Comment = string(comment)
	}

	info, ok := rawDict["info"].(map[string]interface{})
	if !ok {
		return tf, fmt.Errorf("info property not found in rawDict")
	}

	// TODO: this isn't entirely right. From the spec:
	//
	// The 20 byte sha1 hash of the bencoded form of the info value
	// from the metainfo file. This value will almost certainly have to be
	// escaped. Note that this is a substring of the metainfo file. The
	// info-hash must be the hash of the encoded form as found in the .torrent
	// file, which is identical to bdecoding the metainfo file, extracting the
	// info dictionary and encoding it if and only if the bdecoder fully
	// validated the input (e.g. key ordering, absence of leading zeros).
	// Conversely that means clients must either reject invalid metainfo files
	// or extract the substring directly. They must not perform a decode-encode
	// roundtrip on invalid data.
	infoBytes, err := bencoding.Encode(info)
	if err != nil {
		return
	}
	tf.InfoHash = sha1.Sum(infoBytes)
	name, ok := info["name"].([]byte)
	if !ok {
		return tf, fmt.Errorf("name property not found in info")
	}
	tf.Info.Name = string(name)

	if tf.Info.Length, ok = info["length"].(int); !ok {
		return tf, fmt.Errorf("length property not found in info")
	}

	return
}
