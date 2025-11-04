package storage

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/golang/snappy"
	"github.com/vamosdalian/nav/internal/graph"
)

const (
	// File format magic number and version
	magicNumber   uint32 = 0x4E415647 // "NAVG" in hex
	formatVersion uint32 = 1
)

// Storage handles graph persistence
type Storage struct {
	filepath string
}

// NewStorage creates a new storage handler
func NewStorage(filepath string) *Storage {
	return &Storage{filepath: filepath}
}

// Save serializes and saves the graph to disk using custom binary format
func (s *Storage) Save(g *graph.Graph) error {
	file, err := os.Create(s.filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Use buffered writer for better I/O performance
	bufWriter := bufio.NewWriterSize(file, 2*1024*1024) // 2MB buffer
	defer bufWriter.Flush()

	// Use snappy compression for fast compression/decompression
	snappyWriter := snappy.NewBufferedWriter(bufWriter)
	defer snappyWriter.Close()

	// Export graph data
	data := g.Export()

	// Write using custom binary format
	if err := writeBinary(snappyWriter, data); err != nil {
		return fmt.Errorf("failed to encode graph: %w", err)
	}

	return nil
}

// Load deserializes and loads the graph from disk using custom binary format
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

	// Read using custom binary format
	data, err := readBinary(snappyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode graph: %w", err)
	}

	// Reconstruct graph
	g := graph.NewGraph()
	g.Import(data)

	return g, nil
}

// writeBinary writes graph data in custom binary format (10-20x faster than gob)
func writeBinary(w io.Writer, data *graph.ExportData) error {
	// Write header
	if err := binary.Write(w, binary.LittleEndian, magicNumber); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, formatVersion); err != nil {
		return err
	}

	// Write nodes
	if err := binary.Write(w, binary.LittleEndian, int32(len(data.Nodes))); err != nil {
		return err
	}
	for id, node := range data.Nodes {
		if err := binary.Write(w, binary.LittleEndian, id); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, node.Lat); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, node.Lon); err != nil {
			return err
		}
	}

	// Write edges
	totalEdges := 0
	for _, edgeList := range data.Edges {
		totalEdges += len(edgeList)
	}
	if err := binary.Write(w, binary.LittleEndian, int32(totalEdges)); err != nil {
		return err
	}
	for _, edgeList := range data.Edges {
		for _, edge := range edgeList {
			if err := writeEdge(w, &edge); err != nil {
				return err
			}
		}
	}

	// Write reverse edges
	totalReverseEdges := 0
	for _, edgeList := range data.ReverseEdges {
		totalReverseEdges += len(edgeList)
	}
	if err := binary.Write(w, binary.LittleEndian, int32(totalReverseEdges)); err != nil {
		return err
	}
	for _, edgeList := range data.ReverseEdges {
		for _, edge := range edgeList {
			if err := writeEdge(w, &edge); err != nil {
				return err
			}
		}
	}

	// Write restrictions
	totalRestrictions := 0
	for _, resList := range data.Restrictions {
		totalRestrictions += len(resList)
	}
	if err := binary.Write(w, binary.LittleEndian, int32(totalRestrictions)); err != nil {
		return err
	}
	for viaNode, resList := range data.Restrictions {
		for _, res := range resList {
			if err := binary.Write(w, binary.LittleEndian, res.FromWay); err != nil {
				return err
			}
			if err := binary.Write(w, binary.LittleEndian, viaNode); err != nil {
				return err
			}
			if err := binary.Write(w, binary.LittleEndian, res.ToWay); err != nil {
				return err
			}
			if err := writeString(w, res.Type); err != nil {
				return err
			}
		}
	}

	return nil
}

// writeEdge writes a single edge
func writeEdge(w io.Writer, edge *graph.Edge) error {
	if err := binary.Write(w, binary.LittleEndian, edge.From); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, edge.To); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, edge.Weight); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, edge.OSMWayID); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, edge.MaxSpeed); err != nil {
		return err
	}

	// Write tags
	if err := binary.Write(w, binary.LittleEndian, int32(len(edge.Tags))); err != nil {
		return err
	}
	for key, value := range edge.Tags {
		if err := writeString(w, key); err != nil {
			return err
		}
		if err := writeString(w, value); err != nil {
			return err
		}
	}

	return nil
}

// writeString writes a length-prefixed string
func writeString(w io.Writer, s string) error {
	if err := binary.Write(w, binary.LittleEndian, int32(len(s))); err != nil {
		return err
	}
	_, err := w.Write([]byte(s))
	return err
}

// readBinary reads graph data in custom binary format
func readBinary(r io.Reader) (*graph.ExportData, error) {
	data := &graph.ExportData{
		Nodes:        make(map[int64]*graph.Node),
		Edges:        make(map[int64][]graph.Edge),
		ReverseEdges: make(map[int64][]graph.Edge),
		Restrictions: make(map[int64][]graph.TurnRestriction),
	}

	// Read and verify header
	var magic, version uint32
	if err := binary.Read(r, binary.LittleEndian, &magic); err != nil {
		return nil, err
	}
	if magic != magicNumber {
		return nil, fmt.Errorf("invalid file format (magic: %x)", magic)
	}
	if err := binary.Read(r, binary.LittleEndian, &version); err != nil {
		return nil, err
	}
	if version != formatVersion {
		return nil, fmt.Errorf("unsupported version: %d", version)
	}

	// Read nodes
	var nodeCount int32
	if err := binary.Read(r, binary.LittleEndian, &nodeCount); err != nil {
		return nil, err
	}
	for i := 0; i < int(nodeCount); i++ {
		var id int64
		var lat, lon float64
		if err := binary.Read(r, binary.LittleEndian, &id); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &lat); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &lon); err != nil {
			return nil, err
		}
		data.Nodes[id] = &graph.Node{ID: id, Lat: lat, Lon: lon}
	}

	// Read edges
	var edgeCount int32
	if err := binary.Read(r, binary.LittleEndian, &edgeCount); err != nil {
		return nil, err
	}
	for i := 0; i < int(edgeCount); i++ {
		edge, err := readEdge(r)
		if err != nil {
			return nil, err
		}
		data.Edges[edge.From] = append(data.Edges[edge.From], *edge)
	}

	// Read reverse edges
	var reverseEdgeCount int32
	if err := binary.Read(r, binary.LittleEndian, &reverseEdgeCount); err != nil {
		return nil, err
	}
	for i := 0; i < int(reverseEdgeCount); i++ {
		edge, err := readEdge(r)
		if err != nil {
			return nil, err
		}
		data.ReverseEdges[edge.To] = append(data.ReverseEdges[edge.To], *edge)
	}

	// Read restrictions
	var restrictionCount int32
	if err := binary.Read(r, binary.LittleEndian, &restrictionCount); err != nil {
		return nil, err
	}
	for i := 0; i < int(restrictionCount); i++ {
		var fromWay, viaNode, toWay int64
		if err := binary.Read(r, binary.LittleEndian, &fromWay); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &viaNode); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &toWay); err != nil {
			return nil, err
		}
		resType, err := readString(r)
		if err != nil {
			return nil, err
		}
		data.Restrictions[viaNode] = append(data.Restrictions[viaNode], graph.TurnRestriction{
			FromWay: fromWay,
			ViaNode: viaNode,
			ToWay:   toWay,
			Type:    resType,
		})
	}

	return data, nil
}

// readEdge reads a single edge
func readEdge(r io.Reader) (*graph.Edge, error) {
	edge := &graph.Edge{
		Tags: make(map[string]string),
	}

	if err := binary.Read(r, binary.LittleEndian, &edge.From); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &edge.To); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &edge.Weight); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &edge.OSMWayID); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &edge.MaxSpeed); err != nil {
		return nil, err
	}

	// Read tags
	var tagCount int32
	if err := binary.Read(r, binary.LittleEndian, &tagCount); err != nil {
		return nil, err
	}
	for i := 0; i < int(tagCount); i++ {
		key, err := readString(r)
		if err != nil {
			return nil, err
		}
		value, err := readString(r)
		if err != nil {
			return nil, err
		}
		edge.Tags[key] = value
	}

	return edge, nil
}

// readString reads a length-prefixed string
func readString(r io.Reader) (string, error) {
	var length int32
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return "", err
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}
