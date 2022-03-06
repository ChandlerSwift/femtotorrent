package tracker

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/chandlerswift/femtotorrent/libfemtotorrent/bencoding"
	"github.com/chandlerswift/femtotorrent/libfemtotorrent/peer"
	"github.com/chandlerswift/femtotorrent/libfemtotorrent/torrentfile"
)

// GetPeers retrieves a list of peers from the remote tracker
// TODO: implement http://bittorrent.org/beps/bep_0023.html
func GetPeers(tf torrentfile.TorrentFile) (peers []peer.Peer, interval int, err error) {
	q := url.Values{}
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
	q.Add("info_hash", string(tf.InfoHash[:]))

	// peer_id A string of length 20 which this downloader uses as its id. Each
	// downloader generates its own id at random at the start of a new download.
	// This value will also almost certainly have to be escaped.
	q.Add("peer_id", "chandlerswiftdebian!") // TODO

	// ip An optional parameter giving the IP (or dns name) which this peer is
	// at. Generally used for the origin if it's on the same machine as the
	// tracker.

	// port The port number this peer is listening on. Common behavior is for a
	// downloader to try to listen on port 6881 and if that port is taken try
	// 6882, then 6883, etc. and give up after 6889.
	// TODO

	// uploaded The total amount uploaded so far, encoded in base ten ascii.
	q.Add("uploaded", "0")
	// downloaded The total amount downloaded so far, encoded in base ten ascii.
	q.Add("downloaded", "0")

	// left The number of bytes this peer still has to download, encoded in base
	// ten ascii. Note that this can't be computed from downloaded and the file
	// length since it might be a resume, and there's a chance that some of the
	// downloaded data failed an integrity check and had to be re-downloaded.
	q.Add("left", fmt.Sprint(tf.Info.Length)) // TODO: multi-file case

	// event This is an optional key which maps to started, completed, or
	// stopped (or empty, which is the same as not being present). If not
	// present, this is one of the announcements done at regular intervals. An
	// announcement using started is sent when a download first begins, and one
	// using completed is sent when the download is complete. No completed is
	// sent if the file was complete when started. Downloaders send an
	// announcement using stopped when they cease downloading.

	res, err := http.Get(fmt.Sprintf("%v?%v", tf.Announce, q.Encode()))
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	rawBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	untypedBody, err := bencoding.Decode(rawBody)

	body, ok := untypedBody.(map[string]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("body of response was unexpected type '%T', not map[string]interface{}", body)
	}

	if failureReason, ok := body["failure reason"].([]byte); ok {
		return nil, 0, fmt.Errorf("Received failure from server: %+v", string(failureReason))
	}

	interval, ok = body["interval"].(int)
	if !ok {
		return nil, 0, fmt.Errorf("interval not found in response %+v", body)
	}

	rawPeers, ok := body["peers"].([]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("peers list not found in response %+v", body)
	}
	for _, rawPeer := range rawPeers {
		p, ok := rawPeer.(map[string]interface{})
		if ok {
			ip, ok := p["ip"].([]byte)
			if !ok {
				return nil, 0, fmt.Errorf("ip not found in peer %+v", p)
			}
			port, ok := p["port"].(int)
			if !ok {
				return nil, 0, fmt.Errorf("port not found in peer %+v", p)
			}
			peers = append(peers, peer.Peer{
				IPAddress: ip,
				Port:      uint16(port),
			})
		} else {
			return nil, 0, fmt.Errorf("peer was unexpected type '%T', not map[string]interface{}", p)
		}
	}
	return
}
