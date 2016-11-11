package zapwriter

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"sync"

	"github.com/uber-go/zap"
)

type closeable interface {
	Close() (err error)
}

type Output interface {
	io.Writer
	Sync() error
}

type output struct {
	sync.RWMutex
	out       zap.WriteSyncer
	closeable bool
	dsn       string
}

func New(dsn string) (Output, error) {
	o := &output{}

	err := o.apply(dsn)
	if err != nil {
		return nil, err
	}

	return o, err
}

func (o *output) apply(dsn string) error {
	if dsn == o.dsn && o.out != nil { // nothing changed
		return nil
	}

	var newOut zap.WriteSyncer
	var newCloseable bool

	if dsn == "" || dsn == "stderr" {
		newOut = os.Stderr
	} else if dsn == "stdout" {
		newOut = os.Stdout
	} else {
		u, err := url.Parse(dsn)

		if err != nil {
			return err
		}

		if u.Scheme == "" || u.Scheme == "file" {
			newOut, err = File(u.Path)
			if err != nil {
				return err
			}
			newCloseable = true
		} else {
			return fmt.Errorf("unknown scheme %#v", u.Scheme)
		}
	}

	if o.out != nil && o.closeable {
		if c, ok := o.out.(closeable); ok {
			c.Close()
		}
		o.out = nil
	}

	o.out = newOut
	o.closeable = newCloseable

	return nil
}

func (o *output) Sync() (err error) {
	o.RLock()
	if o.out != nil {
		err = o.out.Sync()
	}
	o.RUnlock()
	return
}

func (o *output) Write(p []byte) (n int, err error) {
	o.RLock()
	if o.out != nil {
		n, err = o.out.Write(p)
	}
	o.RUnlock()
	return
}
