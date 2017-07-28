package cache

import (
	"time"
)

// Entry contains data
type Entry struct {
	Key        *Key
	CreateDate time.Time // In UTC
	EndDate    int64
	Data       interface{}
}

// CreateEntry returns new instance of Entry
func CreateEntry(key *Key, endDate int64, data interface{}) *Entry {
	return &Entry{
		Key:        key,
		CreateDate: time.Now().UTC(),
		EndDate:    endDate,
		Data:       data,
	}
}

// IsValid returns Entry is valid
func (entry *Entry) IsValid() bool {
	return entry.EndDate > time.Now().Unix()
}
