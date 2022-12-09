package main

// Album represents data about a single album.
type Album struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Price  int    `json:"price,omitempty"` // use int cents instead of float64 for currency
}
