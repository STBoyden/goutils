package netutils

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"reflect"
)

// TCPTypedConnection is a TypedConnection that is suited for TCP Connections and
// provides TCP-specific function implementations.
type TCPTypedConnection[T Convertable] struct {
	TypedConnection[T]
}

// NewTCPTypedConnection creates a new TCPTypedConnection specialised for T.
func NewTCPTypedConnection[T Convertable](conn net.Conn) TCPTypedConnection[T] {
	return TCPTypedConnection[T]{TypedConnection[T]{conn: conn, connectionType: ConnectionTypeTCP}}
}

// ReadFrom reads from the inner connection, attempting to read a T from the connection
// On success the amount of bytes read is returned and the data parameter is populated
// with the read data from the connection. On failure, the amount of bytes read is still
// returned alongside an error. The data pointer is left untouched.
//
// This takes a variadic parameter of type ReadOptions, which can be used to set the chunk
// size and buffer size to be used. If no ReadOptions are supplied, then the defaults are
// used using the private defaultReadFromOptions function. If more than one ReadOptions
// are supplied then only the first will be used.
func (ttc *TCPTypedConnection[T]) ReadFrom(data *T, opts ...ReadOptions) (int64, error) {
	var readOpts ReadOptions
	if opts == nil {
		readOpts = defaultReadOptions()
	} else {
		readOpts = opts[0]
	}

	switch conn := ttc.conn.(type) {
	case *net.TCPConn:
		buffer := make([]byte, readOpts.BufferSize)
		reader := bytes.NewReader(buffer)

		amountRead, err := conn.ReadFrom(reader)
		if err != nil {
			return amountRead, errors.Join(errors.New("could not receive incoming buffer"), err)
		}

		resizedBuffer := buffer[:amountRead]

		var newData T
		err = newData.Unmarshal(&newData, resizedBuffer)
		if err != nil {
			return amountRead, errors.Join(fmt.Errorf("could not unmarshal incoming buffer into %s", reflect.TypeOf(data)))
		}

		*data = newData
		return amountRead, nil
	default:
		return 0, errors.New("conn is an invalid connection type for this method")
	}
}

// DialTCP attempts to connect to a given TCP socket at host:port, and creates a new
// TCPTypedConnection[T] on success. On failure, an error is returned.
func DialTCP[T Convertable](host, port string) (*TCPTypedConnection[T], error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		return nil, err
	}

	tc := NewTCPTypedConnection[T](conn)

	return &tc, nil
}

// TCPSocketListener is a type-safe wrapper over *net.TCPSocketListener
type TCPSocketListener[T Convertable] struct {
	listener *net.TCPListener
}

// NewTypedTCPSocketListener creates a *TCPSocketListener from a pre-existing
// *net.TCPListener.
func NewTypedTCPSocketListener[T Convertable](listener *net.TCPListener) *TCPSocketListener[T] {
	return &TCPSocketListener[T]{listener: listener}
}

// Accept starts listening on the inner TCPListener, and creates a *TCPTypedConnection
// from the listener. On success, the new *TCPTypedConnection is returned. On failure, an
// error is returned.
func (tsl *TCPSocketListener[T]) Accept() (*TCPTypedConnection[T], error) {
	conn, err := tsl.listener.Accept()
	if err != nil {
		return nil, err
	}

	tc := NewTCPTypedConnection[T](conn)

	return &tc, nil
}

// Addr wraps the net.TCPListener.Addr function.
func (tsl *TCPSocketListener[T]) Addr() net.Addr {
	return tsl.listener.Addr()
}

// Close wraps the net.TCPListener.Close function.
func (tsl *TCPSocketListener[T]) Close() error {
	return tsl.listener.Close()
}
