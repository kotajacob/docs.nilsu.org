package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	extension "git.sr.ht/~kota/goldmark-extension"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
)

// Doc represents a document at a certain point in time and a list of the other
// times.
type Doc struct {
	Name     *string
	HashName *string
	Data     *string
	Time     *time.Time
	Times    *TimeSlice
}

// TimeSlice is a sortable slice of time.Time
type TimeSlice []time.Time

func (ts *TimeSlice) Append(t time.Time) {
	*ts = append(*ts, t)
}

func (ts TimeSlice) Len() int           { return len(ts) }
func (ts TimeSlice) Less(i, j int) bool { return ts[i].After(ts[j]) }
func (ts TimeSlice) Swap(i, j int)      { ts[i], ts[j] = ts[j], ts[i] }

// WriteFile writes data to the named file, creating it if necessary.  If the
// file does not exist, WriteFile creates it with permissions perm (before
// umask); otherwise WriteFile truncates it before writing, without changing
// permissions.
func WriteFile(name string, data []byte, perm os.FileMode) error {
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}
	return err
}

// ReadFile reads the named file and returns the contents.
// A successful call returns err == nil, not err == EOF.
// Because ReadFile reads the whole file, it does not treat an EOF from Read
// as an error to be reported.
func ReadFile(name string) ([]byte, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var size int
	if info, err := f.Stat(); err == nil {
		size64 := info.Size()
		if int64(int(size64)) == size64 {
			size = int(size64)
		}
	}
	size++ // one byte for final read at EOF

	// If a file claims a small size, read at least 512 bytes.
	// In particular, files in Linux's /proc claim size 0 but
	// then do not work right if read in small pieces,
	// so an initial read of 1 byte would not work correctly.
	if size < 512 {
		size = 512
	}

	data := make([]byte, 0, size)
	for {
		if len(data) >= cap(data) {
			d := append(data[:cap(data)], 0)
			data = d[:len(data)]
		}
		n, err := f.Read(data[len(data):cap(data)])
		data = data[:len(data)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return data, err
		}
	}
}

// Write Doc to name/unixtime in the Config DataDir and Config HashDir,
// creating any directories needed.
func (d *Doc) Write() error {
	now := time.Now()
	// write Doc to DataDir
	if err := os.MkdirAll(filepath.Join(config.DataDir, *d.Name), 0700); err != nil {
		return err
	}
	path := filepath.Join(config.DataDir, *d.Name, fmt.Sprintf("%v", now.Unix()))
	if err := WriteFile(path, []byte(*d.Data), 0600); err != nil {
		return err
	}
	// write HashDoc to HashDir
	if err := os.MkdirAll(filepath.Join(config.HashDir, *d.HashName), 0700); err != nil {
		return err
	}
	path = filepath.Join(config.HashDir, *d.HashName, fmt.Sprintf("%v", now.Unix()))
	if err := WriteFile(path, []byte(*d.Data), 0600); err != nil {
		return err
	}
	return nil
}

// ReadTimes populates a Doc.Times by walking the file system and parsing the
// unix time file names as time.Time values.
func (d *Doc) ReadTimes() error {
	d.Times = new(TimeSlice)
	dir := path.Join(config.DataDir, *d.Name)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		u, err := strconv.ParseInt(info.Name(), 10, 64)
		if err != nil {
			// log, but skip files that can't be parsed
			log.Println("failed to parse file name as time.Time")
			return nil
		}
		t := time.Unix(u, 0)
		d.Times.Append(t)
		return nil
	})
	if err != nil {
		return err
	}
	sort.Sort(d.Times)
	return nil
}

// ReadHashTimes populates a Doc.Times by walking the file system and parsing the
// unix time file names as time.Time values.
func (d *Doc) ReadHashTimes() error {
	d.Times = new(TimeSlice)
	dir := path.Join(config.HashDir, *d.HashName)
	log.Println(dir)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		log.Println("path:", path)
		log.Println("entry:", info.Name())
		if info.IsDir() {
			return nil
		}
		u, err := strconv.ParseInt(info.Name(), 10, 64)
		if err != nil {
			// log, but skip files that can't be parsed
			log.Println("failed to parse file name as time.Time")
			return nil
		}
		t := time.Unix(u, 0)
		d.Times.Append(t)
		return nil
	})
	if err != nil {
		return err
	}
	sort.Sort(d.Times)
	return nil
}

// ReadData populates the data stored in a Doc at it's current set time. This
// should be used AFTER setting the Doc's Time.
func (d *Doc) ReadData() error {
	dir := path.Join(config.DataDir, *d.Name, fmt.Sprintf("%v", d.Time.Unix()))
	data, err := ReadFile(dir)
	if err != nil {
		return err
	}
	s := string(data)
	d.Data = &s
	return nil
}

// Hash populates the HashName from the Name fied in a Doc using sha256.
func (d *Doc) Hash() {
	sum := sha256.Sum256([]byte(*d.Name))
	hash := base64.RawURLEncoding.EncodeToString(sum[:])
	d.HashName = &hash
}

// ReadHashData populates the data stored in a Hash of a Doc at it's current
// set time. This should be used AFTER setting the Doc's Time.
func (d *Doc) ReadHashData() error {
	dir := path.Join(config.HashDir, *d.HashName, fmt.Sprintf("%v", d.Time.Unix()))
	data, err := ReadFile(dir)
	if err != nil {
		return err
	}
	s := string(data)
	d.Data = &s
	return nil
}

// HTML converts the data stored in a Doc to HTML using goldmark.
func (d *Doc) HTML() error {
	in := bytes.NewBufferString(*d.Data)
	var out bytes.Buffer
	md := goldmark.New(goldmark.WithExtensions(extension.Table, extension.Strikethrough, extension.Linkify, extension.TaskList, extension.Typographer, emoji.Emoji))
	if err := md.Convert(in.Bytes(), &out); err != nil {
		return err
	}
	data := out.String()
	d.Data = &data
	return nil
}

// ReadDoc returns an instance of a Doc at a specific point Unix time string.
// Other Doc times are included in Doc.Times.
func ReadDoc(name string, t time.Time) (*Doc, error) {
	doc := new(Doc)
	doc.Name = &name
	doc.Hash()
	if err := doc.ReadTimes(); err != nil {
		return nil, err
	}

	// read specified file
	doc.Time = &t
	if err := doc.ReadData(); err != nil {
		return nil, err
	}
	return doc, nil
}

// ReadDocLatest returns the latest version of a Doc from it's name. Past Doc
// times are included in Doc.Times.
func ReadDocLatest(name string) (*Doc, error) {
	doc := new(Doc)
	doc.Name = &name
	doc.Hash()
	if err := doc.ReadTimes(); err != nil {
		return nil, err
	}

	// read latest file
	doc.Time = &(*doc.Times)[0]
	if err := doc.ReadData(); err != nil {
		return nil, err
	}
	return doc, nil
}

// ReadHash returns an instance of a Doc at a specific point Unix time string.
// Other Doc times are included in Doc.Times.
func ReadHash(hash string, t time.Time) (*Doc, error) {
	doc := new(Doc)
	doc.HashName = &hash
	if err := doc.ReadHashTimes(); err != nil {
		return nil, err
	}
	log.Println("got times")

	// read specified file
	doc.Time = &t
	if err := doc.ReadHashData(); err != nil {
		return nil, err
	}
	return doc, nil
}

// ReadHashLatest returns the latest version of a Doc from it's name. Past Doc
// times are included in Doc.Times.
func ReadHashLatest(hash string) (*Doc, error) {
	doc := new(Doc)
	doc.HashName = &hash
	if err := doc.ReadHashTimes(); err != nil {
		return nil, err
	}
	log.Println("got times")

	// read latest file
	doc.Time = &(*doc.Times)[0]
	if err := doc.ReadHashData(); err != nil {
		return nil, err
	}
	return doc, nil
}
