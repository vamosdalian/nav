package storage

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"os"

	"github.com/golang/snappy"
	"github.com/vamosdalian/nav/internal/graph"
)

// Storage handles graph persistence
type Storage struct {
	filepath string
}

// NewStorage creates a new storage handler
func NewStorage(filepath string) *Storage {
	return &Storage{filepath: filepath}
}

// Save serializes and saves the graph to disk
func (s *Storage) Save(g *graph.Graph) error {
	file, err := os.Create(s.filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Use buffered writer for better I/O performance
	bufWriter := bufio.NewWriterSize(file, 1024*1024) // 1MB buffer
	defer bufWriter.Flush()

	// Use snappy compression for fast compression/decompression
	snappyWriter := snappy.NewBufferedWriter(bufWriter)
	defer snappyWriter.Close()

	encoder := gob.NewEncoder(snappyWriter)

	// Export and encode graph data
	data := g.Export()

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode graph: %w", err)
	}

	return nil
}

// Load deserializes and loads the graph from disk
func (s *Storage) Load() (*graph.Graph, error) {
	file, err := os.Open(s.filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Use buffered reader for better I/O performance
	bufReader := bufio.NewReader(file)

	// Use snappy decompression for fast decompression
	snappyReader := snappy.NewReader(bufReader)

	decoder := gob.NewDecoder(snappyReader)

	var data graph.ExportData
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode graph: %w", err)
	}

	// Reconstruct graph
	g := graph.NewGraph()
	g.Import(&data)

	return g, nil
}
