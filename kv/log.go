package kv

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"io"
	"os"
)

// EntryType stored in the first payload byte
type EntryType uint8

const (
	OpSet EntryType = 1
	OpDel EntryType = 2
)

// writeLogEntry writes: [4 bytes length][4 bytes crc32][payload bytes]
// It fsyncs the file after write to make the append durable.
func writeLogEntry(f *os.File, payload []byte) error {
	// build header
	var hdr [8]byte
	binary.BigEndian.PutUint32(hdr[0:4], uint32(len(payload)))
	crc := crc32.ChecksumIEEE(payload)
	binary.BigEndian.PutUint32(hdr[4:8], crc)

	// write header + payload
	if _, err := f.Write(hdr[:]); err != nil {
		return err
	}
	if _, err := f.Write(payload); err != nil {
		return err
	}
	// durable write
	if err := f.Sync(); err != nil {
		return err
	}
	return nil
}

func buildSetPayload(key, value []byte) []byte {
	buf := &bytes.Buffer{}
	buf.WriteByte(byte(OpSet))
	_ = binary.Write(buf, binary.BigEndian, uint32(len(key)))
	buf.Write(key)
	_ = binary.Write(buf, binary.BigEndian, uint32(len(value)))
	buf.Write(value)
	return buf.Bytes()
}

func buildDelPayload(key []byte) []byte {
	buf := &bytes.Buffer{}
	buf.WriteByte(byte(OpDel))
	_ = binary.Write(buf, binary.BigEndian, uint32(len(key)))
	buf.Write(key)
	return buf.Bytes()
}

// readLog reads entries until a truncated/corrupted entry is encountered.
// It returns a slice of payloads (each payload begins with the entry type byte).
func readLog(f *os.File) ([][]byte, error) {
	var results [][]byte
	for {
		var hdr [8]byte
		if _, err := io.ReadFull(f, hdr[:]); err != nil {
			// truncated header or EOF -> stop replay gracefully
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return results, nil
			}
			return results, err
		}
		size := binary.BigEndian.Uint32(hdr[0:4])
		expectedCrc := binary.BigEndian.Uint32(hdr[4:8])

		payload := make([]byte, size)
		if _, err := io.ReadFull(f, payload); err != nil {
			// truncated payload -> stop replay
			return results, nil
		}
		crc := crc32.ChecksumIEEE(payload)
		if crc != expectedCrc {
			// checksum mismatch -> stop replay
			return results, nil
		}
		results = append(results, payload)
	}
}
