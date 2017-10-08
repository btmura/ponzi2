// Code generated by go-bindata.
// sources:
// data/addButton.png
// data/refreshButton.png
// data/removeButton.png
// data/roundedCornerNE.ply
// data/roundedCornerNW.ply
// data/roundedCornerSE.ply
// data/roundedCornerSW.ply
// DO NOT EDIT!

package ponzi

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _addbuttonPng = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xea\x0c\xf0\x73\xe7\xe5\x92\xe2\x62\x60\x60\xe0\xf5\xf4\x70\x09\x62\x60\x60\x70\x00\x61\x0e\x36\x06\x06\x86\x55\x99\x85\xf7\x18\x18\x18\x0a\x3d\x5d\x1c\x43\x2a\x6e\xbd\xbd\x79\x91\xf3\x80\x02\x8f\x83\xe3\xe3\x7f\xd6\xc7\x79\x59\x1c\x27\xf5\xdf\x89\x5e\xf0\x6d\xc5\x82\x97\x51\x93\x54\x3c\x27\xa9\x89\x5d\xe5\x3f\x73\x66\x8f\x9e\xc6\xa2\x12\xdf\xea\x5f\xe2\x0c\xa8\xe0\x0c\xb3\xdd\x9d\xe7\xb7\xe3\x5f\x14\xce\x4b\xe0\x86\x88\x3c\x6f\xd8\xf5\x47\x7c\x03\x2b\xaa\xb2\x06\x15\xcf\x49\x2a\x5b\x4b\x18\x3a\x78\x79\xfc\x22\x99\x6f\xe6\x83\xc4\x3c\x5d\xfd\x5c\xd6\x39\x25\x34\x01\x02\x00\x00\xff\xff\xb2\x0c\x4b\x45\xaa\x00\x00\x00")

func addbuttonPngBytes() ([]byte, error) {
	return bindataRead(
		_addbuttonPng,
		"addButton.png",
	)
}

func addbuttonPng() (*asset, error) {
	bytes, err := addbuttonPngBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "addButton.png", size: 170, mode: os.FileMode(438), modTime: time.Unix(1507506109, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _refreshbuttonPng = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x00\xaf\x02\x50\xfd\x89\x50\x4e\x47\x0d\x0a\x1a\x0a\x00\x00\x00\x0d\x49\x48\x44\x52\x00\x00\x00\x40\x00\x00\x00\x40\x08\x06\x00\x00\x00\xaa\x69\x71\xde\x00\x00\x02\x76\x49\x44\x41\x54\x78\xda\xed\x9a\xbf\x6f\x13\x31\x14\xc7\xed\xd2\x22\x06\x60\x0a\x91\xda\x8a\xa9\x52\x3b\x32\x64\x6c\x86\xec\x0c\xfd\x1f\xb2\xf4\x8f\x28\x03\x53\xa5\x8c\x0c\x27\x66\x24\xf6\x92\x25\x1b\xa8\x03\x5d\xa8\x3a\x30\x86\x09\x65\xa8\xc4\x14\x55\x6c\xfc\xd0\x87\xa1\xae\x74\x0a\x77\xbe\xb3\xe3\x73\x9d\xcb\xfb\x48\x5d\xf2\xd2\xe7\xf7\xbe\x79\xf6\xd9\xcf\xa7\x94\x20\x08\x82\x20\x08\x42\x24\x80\x37\x80\x5e\x67\x01\x00\xde\x01\x9b\xeb\x2c\x00\xc0\x07\xe0\xd1\x3a\x0b\x00\xf0\x09\x78\xba\x2e\x89\xf7\x81\x57\xfc\xcf\x14\x78\xd6\xd6\xa4\xf7\x80\x11\x30\xc3\xce\x0c\x78\xde\xa6\xc4\xbb\x40\x86\x1b\x3f\x80\xfd\x12\x7f\x2f\x81\xee\xaa\x24\x3f\x04\xe6\xf8\x31\x07\x5e\x2c\xf8\x7b\x9d\xb3\x0d\x53\x4f\x3e\x63\x39\xbe\x03\xbb\x39\x7f\x83\x82\xef\x64\xa9\x26\x7f\xb6\x64\xf2\x53\xa0\x53\x91\xfc\x1d\x67\x6d\x4b\xfe\x6b\xfe\x71\x58\x91\x7c\x5a\x22\x38\x96\xfd\x97\x82\xcf\x3e\x02\x8f\x73\xfe\x7a\x0e\xfe\xb2\x14\x16\xbc\x3a\x4c\x80\x41\xc1\x46\x68\x5c\xb4\x1b\x34\x15\x30\xa9\xe9\x7b\x78\x5f\xc9\x77\x6b\xae\xf6\xc7\x25\x3b\xc1\xf7\x55\xe7\x01\xe0\xb8\xe6\x93\xa3\x9b\x62\xe9\xdf\xdc\xfd\xea\x05\x02\xbc\x05\x36\x6a\x8e\x33\x30\xbe\xd2\x99\x0a\x66\x87\x57\xc5\xa0\xe4\x7f\x4f\x3d\xc6\xab\xb3\x28\xee\xc5\x14\x60\xe4\x52\xf6\x81\xc6\xac\x9a\x0e\xa3\x98\x02\xd8\xf6\xf6\x93\x06\xc7\xb5\x2d\x8c\xb3\x98\xa7\x3a\xe7\xd2\x0f\x34\x76\xd5\x54\xe8\xbb\xfa\xdc\xf0\x88\xc3\x96\xe0\xa5\xd6\xfa\xbc\x29\x01\x8c\xef\x4b\xcf\xd8\x82\x09\xd0\xb3\xd8\xc6\x11\x8a\x70\xec\x19\x5b\x30\x01\xf6\x2d\xb6\x8b\x08\x02\x5c\x78\xc6\x16\x4c\x80\x1d\x8b\x6d\x1a\x41\x80\xa9\x67\x6c\xc5\xd3\xca\x63\x21\xfa\xa3\x94\x7a\x50\x62\x7e\xa8\xb5\xfe\xdd\xf0\x22\xbc\xa5\x94\xfa\x55\x62\xfe\xab\xb5\xde\x6c\xba\x02\x5a\x85\x8f\x00\x3f\x2d\xb6\x4e\x84\x98\x3b\x9e\xb1\x05\x13\xe0\xda\x62\x3b\x88\x20\xc0\x81\x67\x6c\xc1\x04\xf8\x66\xb1\x1d\x46\x10\xe0\xd0\x33\xb6\x60\x02\x5c\x59\x6c\x47\x11\x04\x38\xf2\x8c\x4d\xb6\xc2\x72\x18\x92\xe3\x70\x02\x0d\x91\x86\x4a\x3f\x6e\x43\x64\x99\x96\x98\x67\xf2\x37\xc9\x75\x87\x7d\x9b\xa2\x81\xcb\xfe\xfe\x9a\xa2\x26\x40\xe7\xb6\xb8\xc3\xaf\x9e\x76\x5b\xdc\x61\x2a\x2c\x5e\x8c\x9c\x98\x04\xb7\x81\x2d\xf3\xb7\x6d\x3e\x3b\x29\xb9\x3c\x49\xf3\x62\x24\x27\xc2\xb2\x57\x63\x3e\xb4\xee\x7e\x70\x75\x93\xf7\x9c\x0e\xbe\xa4\x79\x3d\xbe\xb0\x30\xce\x1b\x48\x3c\xfd\x17\x24\x16\x1e\x91\x21\xab\x21\x5b\x99\x57\x64\x0a\x76\x8c\x75\x5e\x92\x2a\x7b\x71\x6a\xd4\xf4\x0e\x4f\x47\x14\xa3\xaf\x6e\xfb\xf6\x3d\x75\xdb\xbd\xdd\x51\x4a\x3d\xc9\x75\x72\xae\xcd\x79\xfe\x4a\x29\x75\xae\xb5\xfe\xac\x04\x41\x10\x04\x41\x10\x04\x41\x10\x1a\xe2\x1f\xf7\x1d\xeb\x14\xce\xff\xde\xa0\x00\x00\x00\x00\x49\x45\x4e\x44\xae\x42\x60\x82\x01\x00\x00\xff\xff\xd8\x7b\x83\xa1\xaf\x02\x00\x00")

func refreshbuttonPngBytes() ([]byte, error) {
	return bindataRead(
		_refreshbuttonPng,
		"refreshButton.png",
	)
}

func refreshbuttonPng() (*asset, error) {
	bytes, err := refreshbuttonPngBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "refreshButton.png", size: 687, mode: os.FileMode(438), modTime: time.Unix(1507506722, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _removebuttonPng = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x00\x12\x02\xed\xfd\x89\x50\x4e\x47\x0d\x0a\x1a\x0a\x00\x00\x00\x0d\x49\x48\x44\x52\x00\x00\x00\x40\x00\x00\x00\x40\x08\x06\x00\x00\x00\xaa\x69\x71\xde\x00\x00\x01\xd9\x49\x44\x41\x54\x78\xda\xed\x9a\x4d\x6e\xc2\x30\x10\x85\x9f\x81\x2d\xe7\xa0\xea\xb2\x12\x08\x24\xd6\xec\xb8\x40\xa5\xde\xa0\x0b\x84\x38\x65\xab\xde\xa0\xeb\x72\x06\x88\xe0\x75\xd1\x20\x45\x08\x90\xed\x38\x63\x4f\x3a\xdf\x3a\xc2\xf3\x3e\x3b\x0e\xfe\x01\x0c\xc3\x30\x0c\xc3\x30\x0c\xc3\x88\x85\xe4\x28\x43\x9b\xc3\x14\xbf\x33\x48\x50\xc8\x18\xc0\x3b\xc9\x85\x60\xf8\x15\x80\xd7\xba\xed\xac\x3d\x3f\x26\xb9\xe5\x1f\x5f\x24\xd7\x02\x6d\xae\xeb\xb6\x7e\x48\xee\xb2\x49\xa8\xc3\xef\xea\x42\x28\x21\xa1\x11\xfe\x42\x1e\x09\x77\xc2\x77\x2a\xe1\x46\xf8\x3c\x12\xae\x86\xfd\x3d\x92\x4a\x78\x10\xbe\x29\x61\x2b\x32\xdb\x93\xdc\xd0\x8f\x24\x12\x3c\xc2\x37\x25\xbc\xa5\xfa\x3a\x3c\x2a\x68\xe1\x59\x10\x49\x1e\x49\xce\x5a\xb4\x35\x25\x79\x08\x10\xbe\x92\x7a\x0d\xd6\x01\x12\xa2\x46\x82\x44\x1b\xc5\x4a\x28\x3e\x7c\x64\xa1\x5e\xaf\x43\xc4\xb0\xcf\x13\xbe\x8b\xde\x52\xd3\xf3\x5d\x14\xae\x36\x7c\x8a\x00\xea\xc3\xb7\x09\xd2\x9b\xf0\x0d\x09\xb3\x7a\xc2\xf3\xa1\x0a\x78\xf6\x40\x72\xaa\x65\x7f\x20\xa4\x57\xfb\xd1\xf3\x1d\x4a\xd0\x17\x3e\xa1\x04\xbd\xe1\x13\x48\xd0\x1f\xfe\x4a\xc2\x77\x40\xf8\x23\xc9\x89\x44\x6d\x83\x42\x9d\xb9\xf3\xf9\xec\xfa\xd2\xfb\x21\xff\xed\x93\x2d\xa5\xb5\x87\xd7\x2f\x21\x41\x78\xbd\x12\x3a\xf8\x23\xa4\x47\x42\xc4\x1e\x41\xd5\x1b\x09\x81\xe1\x0f\x24\x27\xa7\xd3\xe9\x29\x60\x3d\x50\xae\x84\x88\xf0\xd3\xc8\x05\x54\x79\x12\x02\x27\xbc\x9b\xab\x3a\xb5\x12\x52\x84\x57\x2b\x21\x65\x78\x75\x12\xda\xbc\xf3\xea\x25\xfc\xeb\x83\x11\x92\xf3\x80\x6f\x77\xab\x6d\xac\xc0\x91\x10\x75\x34\x16\xb3\x1a\x9c\x03\xf0\xb9\x12\x73\x04\xb0\x74\xce\x7d\x46\x2f\x09\x9d\xfb\x00\xb0\x04\x50\x79\x3c\xfe\x02\xe0\x59\x62\x04\x5c\xee\x06\xec\x0b\x3a\x1e\xdf\x8b\x1c\x8f\x7b\x4a\x90\xbe\x20\xb1\xcf\x7d\x4b\x44\x6c\xeb\xfa\xce\x9c\xb0\x45\x4e\x1a\xb7\x45\x2a\x89\x7d\xfb\x5a\x42\x55\x44\xf8\x46\x51\x1b\x92\x73\xc1\xf6\x16\x24\x37\x25\xad\x05\x86\x19\xda\x1c\xc1\x30\x0c\xc3\x30\x0c\xc3\x30\x8c\x58\x7e\x01\xbd\x24\xdd\xf9\xa8\xbe\x68\xf8\x00\x00\x00\x00\x49\x45\x4e\x44\xae\x42\x60\x82\x01\x00\x00\xff\xff\xe0\xcc\x09\x88\x12\x02\x00\x00")

func removebuttonPngBytes() ([]byte, error) {
	return bindataRead(
		_removebuttonPng,
		"removeButton.png",
	)
}

func removebuttonPng() (*asset, error) {
	bytes, err := removebuttonPngBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "removeButton.png", size: 530, mode: os.FileMode(438), modTime: time.Unix(1507506105, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _roundedcornernePly = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x9c\x8f\xdd\x4a\xc4\x30\x10\x85\xef\x07\xe6\x1d\xe6\x05\x36\xe4\xaf\xae\x5e\x8b\xb7\xbe\x82\xc4\x66\x76\xb7\xd0\x26\x65\x4c\xa5\xf5\xe9\x85\xa2\xb5\x10\xbd\xa8\x85\xc2\xe1\x9b\x73\x4e\x66\xc6\x7e\x41\xb8\x64\x19\x42\xa1\xf0\xd6\x76\x1d\x19\xa5\x11\xda\x3c\x0c\x9c\x0a\x49\x9e\x52\xe4\xf8\x98\x25\xb1\x3c\x3f\xa9\xd5\xce\x3d\xaf\xc3\x77\x96\xc2\x33\x35\x08\xa3\xe4\x91\xa5\x2c\x74\xe9\x73\x28\x34\x57\x64\xa9\xc8\x47\x45\x52\x1d\x4b\x75\x2e\xed\x83\x53\x7b\x0b\x42\xc2\xb1\x62\x57\x61\x4e\x15\x7d\xed\x27\xfe\xd9\x9f\xe3\x95\xc9\xef\x4c\xdd\x76\x94\xf9\x95\x5a\x04\x4e\xf1\xe5\xc6\x21\xb2\x20\x18\xa5\xd7\x8f\x4e\x9b\xd2\x7f\x8b\xcd\x63\x9b\xe6\xfb\x47\xd0\xea\xde\x9f\xcf\xcd\x03\x9d\xb4\xb2\xce\xdf\x39\xf7\x9f\x0e\x6f\xbc\x35\x9e\x76\xe2\x60\xc5\xfe\xf5\xaf\x85\x8e\x77\x98\x6a\x7e\xfc\x14\x32\x08\x86\x2c\x82\x25\x87\xe0\xc8\x7f\x06\x00\x00\xff\xff\x0b\x4c\xbe\x91\xa1\x02\x00\x00")

func roundedcornernePlyBytes() ([]byte, error) {
	return bindataRead(
		_roundedcornernePly,
		"roundedCornerNE.ply",
	)
}

func roundedcornernePly() (*asset, error) {
	bytes, err := roundedcornernePlyBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "roundedCornerNE.ply", size: 673, mode: os.FileMode(438), modTime: time.Unix(1504845949, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _roundedcornernwPly = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x9c\x90\xcf\x4e\xc4\x20\x18\xc4\xef\x3c\xc5\xf7\x02\x25\xfc\xab\xab\x67\xef\x5e\x3d\x1a\x2c\xb3\xbb\x4d\x5a\x68\x3e\xa9\xd9\xfa\xf4\xc6\x8d\x62\x43\xf5\xd0\x25\x21\x99\xfc\x98\x19\x3e\x98\x86\x45\x1c\x13\x8f\x3e\x93\x7f\xeb\xfa\x9e\xb4\x54\xa2\x4b\xe3\x88\x98\x89\xd3\x1c\x03\xc2\x63\xe2\x08\x7e\x7a\x96\x5f\x66\x0c\xb8\x9e\xbd\x83\x33\x2e\xd4\x8a\x89\xd3\x04\xce\x0b\x1d\x87\xe4\x33\x5d\x6a\xb0\xd4\xe0\xa3\x06\x71\x93\x89\x9b\x50\x5c\xa5\xe6\xee\xec\x99\x18\xa1\x46\x27\x06\x62\x0d\x5f\x87\x19\x65\x6a\x84\x13\xc8\xfd\x5a\xfa\xf2\x12\xfd\x17\x34\x02\x31\xbc\x9c\xe1\x03\x58\x68\xa9\xae\x8b\x8a\x50\xff\x8b\xe2\x31\x6d\xfb\xb3\x85\x92\xc6\xba\x3b\x6b\x49\xc9\x7b\x77\x38\xb4\x0f\xbb\x1b\x1a\x25\x9d\x76\x46\x3b\x5a\x89\xdd\x15\xdf\x97\x37\xab\x79\x76\x76\x14\xda\xdc\xfe\x19\xa4\x85\x26\x23\x0c\x59\x61\xc9\x7d\x06\x00\x00\xff\xff\x5e\x8a\x14\x69\x89\x02\x00\x00")

func roundedcornernwPlyBytes() ([]byte, error) {
	return bindataRead(
		_roundedcornernwPly,
		"roundedCornerNW.ply",
	)
}

func roundedcornernwPly() (*asset, error) {
	bytes, err := roundedcornernwPlyBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "roundedCornerNW.ply", size: 649, mode: os.FileMode(438), modTime: time.Unix(1504845956, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _roundedcornersePly = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x9c\x8f\x6f\x4a\xc4\x30\x10\xc5\xbf\x0f\xcc\x1d\xe6\x02\x0d\xf9\x57\x57\x3f\x8b\x27\xf0\x00\x12\x9b\xd9\xdd\x42\x9b\x94\x31\x95\xad\xa7\x17\x8b\x76\x17\xa2\x1f\x6a\x20\x30\xfc\xe6\xbd\x97\x97\x69\x58\x10\x8e\x59\xc6\x50\x28\xbc\x75\x7d\x4f\x46\x69\x84\x2e\x8f\x23\xa7\x42\x92\xe7\x14\x39\x3e\x66\x49\x2c\xcf\x4f\x6a\x95\xf3\xc0\xeb\xf2\x9d\xa5\xf0\x85\x5a\x84\x49\xf2\xc4\x52\x16\x3a\x0e\x39\x14\xba\x54\x64\xa9\xc8\x47\x45\x52\x6d\x4b\xb5\x2f\xdd\x1a\xe7\xee\x1c\x84\x84\x63\xc5\x4e\xc2\x9c\x2a\xfa\x3a\xcc\x7c\xed\xcf\xf1\xc4\xe4\x6f\x44\xfd\xf6\x29\xf3\x2b\xb5\x08\x9c\xe2\xcb\x99\x43\x64\x41\x68\x8c\xd2\xeb\xa1\xeb\xa4\xff\x1e\x36\x8d\x6d\xdb\x9f\x8b\xd0\x68\x65\x9d\xbf\x73\x8e\x1a\xad\xee\xfd\xe1\xd0\x3e\xec\x0f\xd1\xca\x1b\x6f\x8d\xff\xca\x58\x27\xf7\x9f\x8c\xed\xf5\xef\x42\xbb\x23\x4c\xb5\xde\xdf\x82\x0c\x82\x21\x8b\x60\xc9\x21\x38\xf2\x9f\x01\x00\x00\xff\xff\x38\xf0\x26\xca\xa2\x02\x00\x00")

func roundedcornersePlyBytes() ([]byte, error) {
	return bindataRead(
		_roundedcornersePly,
		"roundedCornerSE.ply",
	)
}

func roundedcornersePly() (*asset, error) {
	bytes, err := roundedcornersePlyBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "roundedCornerSE.ply", size: 674, mode: os.FileMode(438), modTime: time.Unix(1504845965, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _roundedcornerswPly = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x9c\x90\x4d\x6a\xc3\x30\x10\x85\xf7\x03\x73\x87\xb9\x80\x8d\xfe\xdc\xb4\xeb\x1e\xa1\x8b\x2e\x8b\x6a\x4d\x12\x83\x2d\x99\xa9\x5c\xe2\x9e\xbe\x60\xa8\x62\x50\xb3\x48\x04\x82\xc7\xa7\xf7\x9e\x46\x9a\xc7\x15\xe1\x98\x64\xf2\x99\xfc\x57\x3f\x0c\xa4\x5b\x85\xd0\xa7\x69\xe2\x98\x49\xd2\x12\x03\x87\xd7\x24\x91\xe5\xed\xbd\xdd\xec\x3c\xf2\x76\xf8\xcd\x92\xf9\x42\x1d\xc2\x2c\x69\x66\xc9\x2b\x1d\xc7\xe4\x33\x5d\x2a\xb2\x56\xe4\xa7\x22\xb1\x8e\xc5\x3a\x17\xf7\xc1\xa5\x3f\x7b\x21\xe1\x50\xb1\x93\x30\xc7\x8a\x7e\x8e\x0b\x5f\xe7\xe7\x70\x62\x72\x3b\xd3\x50\x1e\xa5\xff\xa5\x06\x81\x63\xf8\x38\xb3\x0f\x2c\x08\x8d\x6e\xd5\xb6\xa8\x08\x75\x5b\x14\x8f\xe9\xba\xbf\x8d\xd0\xa8\xf6\xd9\x1d\x0e\xdd\x0b\xa9\xd6\x58\xf7\x64\xed\x43\x1d\x4e\x3b\xa3\x1d\x5d\xd5\xdd\x25\xe5\xfa\xfd\x44\x77\x76\x14\xda\x3c\xfe\x21\x8a\x34\x82\x26\x83\x60\xc8\x22\x58\x72\xbf\x01\x00\x00\xff\xff\x15\x20\x87\xfc\xa3\x02\x00\x00")

func roundedcornerswPlyBytes() ([]byte, error) {
	return bindataRead(
		_roundedcornerswPly,
		"roundedCornerSW.ply",
	)
}

func roundedcornerswPly() (*asset, error) {
	bytes, err := roundedcornerswPlyBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "roundedCornerSW.ply", size: 675, mode: os.FileMode(438), modTime: time.Unix(1504845973, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"addButton.png": addbuttonPng,
	"refreshButton.png": refreshbuttonPng,
	"removeButton.png": removebuttonPng,
	"roundedCornerNE.ply": roundedcornernePly,
	"roundedCornerNW.ply": roundedcornernwPly,
	"roundedCornerSE.ply": roundedcornersePly,
	"roundedCornerSW.ply": roundedcornerswPly,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}
var _bintree = &bintree{nil, map[string]*bintree{
	"addButton.png": &bintree{addbuttonPng, map[string]*bintree{}},
	"refreshButton.png": &bintree{refreshbuttonPng, map[string]*bintree{}},
	"removeButton.png": &bintree{removebuttonPng, map[string]*bintree{}},
	"roundedCornerNE.ply": &bintree{roundedcornernePly, map[string]*bintree{}},
	"roundedCornerNW.ply": &bintree{roundedcornernwPly, map[string]*bintree{}},
	"roundedCornerSE.ply": &bintree{roundedcornersePly, map[string]*bintree{}},
	"roundedCornerSW.ply": &bintree{roundedcornerswPly, map[string]*bintree{}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

