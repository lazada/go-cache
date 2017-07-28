package cache

import (
	"bytes"
	"fmt"
	"strings"
)

const sep byte = '_'

// Key defines cache key for some record
type Key struct {
	Set  string
	Pk   string
	Tags []string
}

// String implements fmt.Stringer interface for readable string representation
func (key Key) String() string {
	return fmt.Sprintf("Set: '%s', PK: '%s', Tags: '%v'", key.Set, key.Pk, key.Tags)
}

// ID returns string that identifies given key
func (key Key) ID() string {
	var buf bytes.Buffer
	buf.WriteString(key.Set)
	buf.WriteByte(sep)
	buf.WriteString(key.Pk)

	if len(key.Tags) != 0 {
		buf.WriteByte(sep)
		buf.WriteString(strings.Join(key.Tags, "_"))
	}

	return buf.String()
}
