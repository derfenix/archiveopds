package inpx

import (
	"archive/zip"
	"container/list"
	"log/slog"
	"sync"
	"sync/atomic"
)

// zipVolume holds one open *zip.ReadCloser for a .zip on disk, shared by acquire and annotation reads.
// A mutex serializes only findZipEntry + zip.File.Open; decompression runs without the lock.
// refs: one lease is taken in getOrOpenVolume before releasing zipVolMu (eviction cannot close zr while leased);
// released when the inner entry ReadCloser is closed. Only refs==0 allows LRU evict of the whole volume.
type zipVolume struct {
	zr   *zip.ReadCloser
	mu   sync.Mutex
	refs int32
}

// addLease increments refs for this volume; must be paired with releaseLease (or inner rc Close via refCountedReadCloser).
func (v *zipVolume) addLease() {
	atomic.AddInt32(&v.refs, 1)
}

// releaseLease decrements lease count after find/Open failure or after inner body Close. Safe if over-released: clamps to 0 and logs.
func (v *zipVolume) releaseLease() {
	n := atomic.AddInt32(&v.refs, -1)
	if n < 0 {
		atomic.StoreInt32(&v.refs, 0)
		slog.Default().Warn("zip cache: lease release underflow; refs clamped to 0 (caller bug?)", "ref_after", n)
	}
}

// getOrOpenVolume returns the cached volume for an absolute .zip path, opening it on first use.
// When maxOpenZip>0, opens are LRU-cached and idle entries are evicted; if all are busy, may exceed cap (warn).
func (n *Navigator) getOrOpenVolume(zpath string) (*zipVolume, error) {
	n.zipVolMu.Lock()
	defer n.zipVolMu.Unlock()
	if n.zipVol == nil {
		n.zipVol = make(map[string]*zipVolume)
	}
	if v, ok := n.zipVol[zpath]; ok {
		if n.zipMaxOpen > 0 {
			if el := n.zipLRUIdx[zpath]; el != nil && n.zipLRU != nil {
				n.zipLRU.MoveToBack(el)
			}
		}
		v.addLease() // before zipVolMu unlock; pairs with releaseLease
		return v, nil
	}

	if n.zipMaxOpen > 0 {
		for len(n.zipVol) >= n.zipMaxOpen {
			if n.tryEvictOneIdleVolumeLocked() {
				continue
			}
			if len(n.zipVol) < n.zipMaxOpen {
				break
			}
			slog.Default().Warn("open zip cache: all volumes in use, temporarily exceeding cap",
				"max", n.zipMaxOpen, "zpath", zpath)
			break
		}
	} // maxOpen 0: unbounded, no list

	zr, err := zip.OpenReader(zpath)
	if err != nil {
		return nil, err
	}
	v := &zipVolume{zr: zr}
	n.zipVol[zpath] = v
	if n.zipMaxOpen > 0 {
		if n.zipLRU == nil {
			n.zipLRU = list.New()
			n.zipLRUIdx = make(map[string]*list.Element)
		}
		el := n.zipLRU.PushBack(zpath)
		n.zipLRUIdx[zpath] = el
	}
	v.addLease() // before zipVolMu unlock; pairs with releaseLease
	return v, nil
}

func (n *Navigator) tryEvictOneIdleVolumeLocked() bool {
	if n.zipLRU == nil || n.zipLRU.Len() == 0 {
		return false
	}
	nList := n.zipLRU.Len()
	visited := 0
	for visited < nList {
		el := n.zipLRU.Front()
		if el == nil {
			return false
		}
		zp := el.Value.(string)
		vol := n.zipVol[zp]
		if vol == nil {
			n.zipLRU.Remove(el)
			delete(n.zipLRUIdx, zp)
			return true
		}
		if atomic.LoadInt32(&vol.refs) == 0 {
			_ = vol.zr.Close()
			n.zipLRU.Remove(el)
			delete(n.zipLRUIdx, zp)
			delete(n.zipVol, zp)
			return true
		}
		n.zipLRU.MoveToBack(el)
		visited++
	}
	return false
}
