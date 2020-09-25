/*
Copyright 2019 Banzai Cloud.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"net/http"
	"os"
	"time"
)

// ZeroModTimeFileSystem is an http.FileSystem wrapper.
// It exposes a filesystem exactly like Source, except
// all file modification times are changed to zero.
type ZeroModTimeFileSystem struct {
	Source http.FileSystem
}

func (fs ZeroModTimeFileSystem) Open(name string) (http.File, error) {
	f, err := fs.Source.Open(name)

	return file{f}, err
}

type file struct {
	http.File
}

func (f file) Stat() (os.FileInfo, error) {
	fi, err := f.File.Stat()

	return fileInfo{fi}, err
}

type fileInfo struct {
	os.FileInfo
}

func (fi fileInfo) ModTime() time.Time { return time.Time{} }
