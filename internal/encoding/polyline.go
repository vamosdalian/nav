package encoding

import (
	"math"
)

// EncodePolyline encodes a list of coordinates into a polyline string
// Uses Google's Polyline Algorithm Format
// https://developers.google.com/maps/documentation/utilities/polylinealgorithm
func EncodePolyline(coordinates [][2]float64) string {
	if len(coordinates) == 0 {
		return ""
	}
	
	var result []byte
	var prevLat, prevLon int32
	
	for _, coord := range coordinates {
		lon, lat := coord[0], coord[1]
		
		// Convert to integers (multiply by 1e5 and round)
		iLat := int32(math.Round(lat * 1e5))
		iLon := int32(math.Round(lon * 1e5))
		
		// Calculate differences
		dLat := iLat - prevLat
		dLon := iLon - prevLon
		
		// Encode differences
		result = append(result, encodeValue(dLat)...)
		result = append(result, encodeValue(dLon)...)
		
		prevLat = iLat
		prevLon = iLon
	}
	
	return string(result)
}

// encodeValue encodes a single value
func encodeValue(value int32) []byte {
	// Step 1: Take the signed value and shift it left by one bit
	var encoded uint32
	if value < 0 {
		encoded = uint32(^(value << 1))
	} else {
		encoded = uint32(value << 1)
	}
	
	var result []byte
	
	// Step 2: Break into 5-bit chunks and encode
	for encoded >= 0x20 {
		chunk := (encoded & 0x1f) | 0x20
		result = append(result, byte(chunk+63))
		encoded >>= 5
	}
	result = append(result, byte(encoded+63))
	
	return result
}

// DecodePolyline decodes a polyline string into coordinates
func DecodePolyline(encoded string) [][2]float64 {
	if encoded == "" {
		return nil
	}
	
	var coordinates [][2]float64
	var lat, lon int32
	index := 0
	
	for index < len(encoded) {
		// Decode latitude
		deltaLat, newIndex := decodeValue(encoded, index)
		index = newIndex
		lat += deltaLat
		
		// Decode longitude
		deltaLon, newIndex := decodeValue(encoded, index)
		index = newIndex
		lon += deltaLon
		
		coordinates = append(coordinates, [2]float64{
			float64(lon) / 1e5,
			float64(lat) / 1e5,
		})
	}
	
	return coordinates
}

// decodeValue decodes a single value from the polyline
func decodeValue(encoded string, index int) (int32, int) {
	var result int32
	var shift uint
	var b byte
	
	for {
		b = encoded[index] - 63
		index++
		result |= int32(b&0x1f) << shift
		shift += 5
		if b < 0x20 {
			break
		}
	}
	
	// Undo the bit shift
	if result&1 != 0 {
		result = ^(result >> 1)
	} else {
		result = result >> 1
	}
	
	return result, index
}

