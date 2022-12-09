package main

import (
	"sort"
	"sync"
)

// Database is the interface used by the server to load and store albums.
type Database interface {
	// GetAlbums returns a copy of all albums, sorted by ID.
	GetAlbums() ([]Album, error)

	// GetAlbumByID returns a single album by ID, or ErrDoesNotExist if
	// an album with that ID does not exist.
	GetAlbumByID(id string) (Album, error)

	// AddAlbum adds a single album, or ErrAlreadyExists if an album with
	// the given ID already exists.
	AddAlbum(album Album) error
}

// MemoryDatabase is a Database implementation that uses a simple
// in-memory map to store the albums.
type MemoryDatabase struct {
	lock   sync.RWMutex
	albums map[string]Album
}

// NewMemoryDatabase creates a new in-memory database.
func NewMemoryDatabase() *MemoryDatabase {
	return &MemoryDatabase{albums: make(map[string]Album)}
}

func (d *MemoryDatabase) GetAlbums() ([]Album, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()

	// Make a copy of the albums map (as a slice)
	albums := make([]Album, 0, len(d.albums))
	for _, album := range d.albums {
		albums = append(albums, album)
	}

	// Sort by ID so we return them in a defined order
	sort.Slice(albums, func(i, j int) bool {
		return albums[i].ID < albums[j].ID
	})
	return albums, nil
}

func (d *MemoryDatabase) GetAlbumByID(id string) (Album, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()

	album, ok := d.albums[id]
	if !ok {
		return Album{}, ErrDoesNotExist
	}
	return album, nil
}

func (d *MemoryDatabase) AddAlbum(album Album) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	if _, ok := d.albums[album.ID]; ok {
		return ErrAlreadyExists
	}
	d.albums[album.ID] = album
	return nil
}
