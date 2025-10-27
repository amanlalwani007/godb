package kv

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

// KV is the in-memory map backed by an append-only log file.
type KV struct {
	data   map[string][]byte
	log    *os.File
	logPath string
}

// NewKV opens or creates the log file, replays it into memory and seeks to end for appends.
func NewKV(logPath string) (*KV, error) {
	f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE, 0o664)
	if err != nil {
		return nil, err
	}
	k := &KV{
		data: make(map[string][]byte),
		log:  f,
		logPath: logPath,
	}
	entries, err := readLog(f)
	if err != nil {
		f.Close()
		return nil, err
	}

	// replay entries
	for _, payload := range entries {
		if len(payload) == 0 {
			continue
		}
		switch EntryType(payload[0]) {
		case OpSet:
			off := 1
			if off+4 > len(payload) {
				return nil, fmt.Errorf("malformed set entry")
			}
			klen := int(binary.BigEndian.Uint32(payload[off : off+4])); off += 4
			if off+klen > len(payload) {
				return nil, fmt.Errorf("malformed set entry key")
			}
			key := string(payload[off : off+klen]); off += klen

			if off+4 > len(payload) {
				return nil, fmt.Errorf("malformed set entry value length")
			}
			vlen := int(binary.BigEndian.Uint32(payload[off : off+4])); off += 4
			if off+vlen > len(payload) {
				return nil, fmt.Errorf("malformed set entry value")
			}
			val := make([]byte, vlen)
			copy(val, payload[off:off+vlen])
			k.data[key] = val

		case OpDel:
			off := 1
			if off+4 > len(payload) {
				return nil, fmt.Errorf("malformed del entry")
			}
			klen := int(binary.BigEndian.Uint32(payload[off : off+4])); off += 4
			if off+klen > len(payload) {
				return nil, fmt.Errorf("malformed del entry key")
			}
			key := string(payload[off : off+klen])
			delete(k.data, key)

		default:
			return nil, fmt.Errorf("unknown entry type %d", payload[0])
		}
	}

	// seek to end for subsequent appends
	if _, err := f.Seek(0, 2); err != nil {
		f.Close()
		return nil, err
	}
	return k, nil
}

// Set writes a set entry and updates in-memory map.
func (k *KV) Set(key string, value []byte) error {
	payload := buildSetPayload([]byte(key), value)
	if err := writeLogEntry(k.log, payload); err != nil {
		return err
	}
	k.data[key] = append([]byte(nil), value...)
	return nil
}

// Del writes a delete entry and removes from in-memory map.
func (k *KV) Del(key string) error {
	payload := buildDelPayload([]byte(key))
	if err := writeLogEntry(k.log, payload); err != nil {
		return err
	}
	delete(k.data, key)
	return nil
}

// Get returns a copy of the value if present.
func (k *KV) Get(key string) ([]byte, bool) {
	v, ok := k.data[key]
	if !ok {
		return nil, false
	}
	val := append([]byte(nil), v...)
	return val, true
}

// Close closes the log file handle.
func (k *KV) Close() error {
	return k.log.Close()
}

// Compact builds a compacted log file from current in-memory state.
// Steps:
// 1) Create a temporary new log file (e.g., db.log.compact.tmp).
// 2) Write set entries for all current keys to temp log, fsync the file.
// 3) Rename temp -> db.log.rotated (atomic).
// 4) fsync the directory to make rename durable.
// 5) Reopen new log file for further appends.
func (k *KV) Compact() error {
	dir := filepath.Dir(k.logPath)
	tmpName := k.logPath + ".compact.tmp"
	tmpF, err := os.OpenFile(tmpName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC|os.O_EXCL, 0o664)
	if err != nil {
		return err
	}

	// write current state as set entries (deterministic order is not necessary, but could be sorted)
	for key, val := range k.data {
		payload := buildSetPayload([]byte(key), val)
		if err := writeLogEntry(tmpF, payload); err != nil {
			tmpF.Close()
			_ = os.Remove(tmpName)
			return err
		}
	}
	// ensure tmp file is closed
	if err := tmpF.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}

	// rename tmp -> new log file atomically
	rotatedName := k.logPath + ".compact.new"
	if err := os.Rename(tmpName, rotatedName); err != nil {
		_ = os.Remove(tmpName)
		return err
	}

	// fsync directory to make rename durable
	df, err := os.Open(dir)
	if err != nil {
		return err
	}
	if err := df.Sync(); err != nil {
		df.Close()
		return err
	}
	if err := df.Close(); err != nil {
		return err
	}

	// close current log
	if err := k.log.Close(); err != nil {
		return err
	}

	// Finally, replace the active log with rotatedName using atomic rename
	if err := os.Rename(rotatedName, k.logPath); err != nil {
		return err
	}

	// fsync dir again to ensure final rename durable
	df2, err := os.Open(dir)
	if err != nil {
		return err
	}
	if err := df2.Sync(); err != nil {
		df2.Close()
		return err
	}
	if err := df2.Close(); err != nil {
		return err
	}

	// reopen the log for appends
	newLog, err := os.OpenFile(k.logPath, os.O_RDWR|os.O_APPEND, 0o664)
	if err != nil {
		return err
	}
	k.log = newLog
	return nil
}
