# femtotorrent

A (low-)quality program for obtaining Linux ISOs.

Implements the spec at http://bittorrent.org/beps/bep_0003.html

The official spec is pretty light details and description. I've found the page
at https://wiki.theory.org/BitTorrentSpecification to be much more helpful.

### Issues
- [ ] InfoHash isn't extracted correctly (consistently) -- go doesn't guarantee map order
- [ ] Only works on single file torrents
- [ ] slooooooowwww
- [ ] Probably doesn't work on large torrents (at least on some platforms), as some sizes are stored as ints and not as int64s

### Name: a small bittorrent client
 - ~~[minitorrent](https://github.com/search?q=minitorrent)~~
 - ~~[µTorrent](https://www.utorrent.com/)~~
 - ~~[nanotorrent](https://launchpad.net/nanotorrent)~~
 - ~~[PicoTorrent](https://picotorrent.org/)~~
 - [x] femtotorrent
