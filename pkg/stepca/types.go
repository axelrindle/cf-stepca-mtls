package stepca

import "time"

// Certificate holds a signed Certificate and its private key in-memory only,
// they are never written to disk.
type Certificate struct {
	CrtPEM   []byte
	KeyPEM   []byte
	NotAfter time.Time
}
