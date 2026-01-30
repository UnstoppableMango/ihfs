package union_test

import (
"errors"
"io"
"io/fs"
"syscall"
"time"

. "github.com/onsi/ginkgo/v2"
. "github.com/onsi/gomega"

"github.com/unstoppablemango/ihfs"
"github.com/unstoppablemango/ihfs/testfs"
"github.com/unstoppablemango/ihfs/union"
)

var _ = Describe("CopyToLayer", func() {
var testTime time.Time

BeforeEach(func() {
testTime = time.Now().Add(-1 * time.Hour)
})

Context("when file exists in base", func() {
It("should copy file to layer", func() {
content := []byte("test content")
baseFile := &testfs.File{
ReadFunc: func(p []byte) (int, error) {
n := copy(p, content)
return n, io.EOF
},
StatFunc: func() (ihfs.FileInfo, error) {
fi := testfs.NewFileInfo("test.txt")
fi.SizeFunc = func() int64 { return int64(len(content)) }
fi.ModTimeFunc = func() time.Time { return testTime }
return fi, nil
},
CloseFunc: func() error { return nil },
}

var copiedContent []byte
layerFile := &testfs.File{
WriteFunc: func(p []byte) (int, error) {
copiedContent = append(copiedContent, p...)
return len(p), nil
},
CloseFunc: func() error { return nil },
}

base := testfs.New(
testfs.WithOpen(func(name string) (ihfs.File, error) {
return baseFile, nil
}),
)

var createdFileName string
var chtimeName string
var chtimeAtime, chtimeMtime time.Time

layer := testfs.New(
testfs.WithCreate(func(name string) (ihfs.File, error) {
createdFileName = name
return layerFile, nil
}),
testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
// Directory exists
fi := testfs.NewFileInfo(name)
fi.IsDirFunc = func() bool { return true }
return fi, nil
}),
testfs.WithChtimes(func(name string, atime, mtime time.Time) error {
chtimeName = name
chtimeAtime = atime
chtimeMtime = mtime
return nil
}),
)

err := union.CopyToLayer(base, layer, "test.txt")

Expect(err).NotTo(HaveOccurred())
Expect(createdFileName).To(Equal("test.txt"))
Expect(copiedContent).To(Equal(content))
Expect(chtimeName).To(Equal("test.txt"))
Expect(chtimeAtime).To(Equal(testTime))
Expect(chtimeMtime).To(Equal(testTime))
})

It("should create parent directories", func() {
content := []byte("test")
baseFile := &testfs.File{
ReadFunc: func(p []byte) (int, error) {
n := copy(p, content)
return n, io.EOF
},
StatFunc: func() (ihfs.FileInfo, error) {
fi := testfs.NewFileInfo("test.txt")
fi.SizeFunc = func() int64 { return int64(len(content)) }
fi.ModTimeFunc = func() time.Time { return testTime }
return fi, nil
},
CloseFunc: func() error { return nil },
}

layerFile := &testfs.File{
WriteFunc: func(p []byte) (int, error) { return len(p), nil },
CloseFunc: func() error { return nil },
}

base := testfs.New(
testfs.WithOpen(func(name string) (ihfs.File, error) {
return baseFile, nil
}),
)

var mkdirAllPath string
layer := testfs.New(
testfs.WithMkdirAll(func(path string, perm fs.FileMode) error {
mkdirAllPath = path
return nil
}),
testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
// Directory doesn't exist
return nil, fs.ErrNotExist
}),
testfs.WithCreate(func(name string) (ihfs.File, error) {
return layerFile, nil
}),
testfs.WithChtimes(func(name string, atime, mtime time.Time) error {
return nil
}),
)

err := union.CopyToLayer(base, layer, "dir/subdir/test.txt")

Expect(err).NotTo(HaveOccurred())
Expect(mkdirAllPath).To(Equal("dir/subdir"))
})
})

Context("error handling", func() {
It("should return error when base file cannot be opened", func() {
expectedErr := errors.New("open failed")
base := testfs.New(
testfs.WithOpen(func(name string) (ihfs.File, error) {
return nil, expectedErr
}),
)
layer := testfs.New()

err := union.CopyToLayer(base, layer, "test.txt")

Expect(err).To(Equal(expectedErr))
})

It("should clean up on copy failure", func() {
baseFile := &testfs.File{
ReadFunc: func(p []byte) (int, error) {
return 0, errors.New("read failed")
},
CloseFunc: func() error { return nil },
}

layerFile := &testfs.File{
WriteFunc: func(p []byte) (int, error) { return 0, errors.New("write failed") },
CloseFunc: func() error { return nil },
}

base := testfs.New(
testfs.WithOpen(func(name string) (ihfs.File, error) {
return baseFile, nil
}),
)

var removedFile string
layer := testfs.New(
testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
fi := testfs.NewFileInfo(".")
fi.IsDirFunc = func() bool { return true }
return fi, nil
}),
testfs.WithCreate(func(name string) (ihfs.File, error) {
return layerFile, nil
}),
testfs.WithRemove(func(name string) error {
removedFile = name
return nil
}),
)

err := union.CopyToLayer(base, layer, "test.txt")

Expect(err).To(HaveOccurred())
Expect(removedFile).To(Equal("test.txt"))
})

It("should clean up and return EIO when size mismatch occurs", func() {
content := []byte("test content")
baseFile := &testfs.File{
ReadFunc: func(p []byte) (int, error) {
n := copy(p, content)
return n, io.EOF
},
StatFunc: func() (ihfs.FileInfo, error) {
fi := testfs.NewFileInfo("test.txt")
fi.SizeFunc = func() int64 { return int64(len(content)) + 10 } // Wrong size
return fi, nil
},
CloseFunc: func() error { return nil },
}

layerFile := &testfs.File{
WriteFunc: func(p []byte) (int, error) { return len(p), nil },
CloseFunc: func() error { return nil },
}

base := testfs.New(
testfs.WithOpen(func(name string) (ihfs.File, error) {
return baseFile, nil
}),
)

var removedFile string
layer := testfs.New(
testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
fi := testfs.NewFileInfo(".")
fi.IsDirFunc = func() bool { return true }
return fi, nil
}),
testfs.WithCreate(func(name string) (ihfs.File, error) {
return layerFile, nil
}),
testfs.WithRemove(func(name string) error {
removedFile = name
return nil
}),
)

err := union.CopyToLayer(base, layer, "test.txt")

Expect(err).To(Equal(syscall.EIO))
Expect(removedFile).To(Equal("test.txt"))
})

It("should return error when Chtimes fails", func() {
content := []byte("test")
baseFile := &testfs.File{
ReadFunc: func(p []byte) (int, error) {
n := copy(p, content)
return n, io.EOF
},
StatFunc: func() (ihfs.FileInfo, error) {
fi := testfs.NewFileInfo("test.txt")
fi.SizeFunc = func() int64 { return int64(len(content)) }
fi.ModTimeFunc = func() time.Time { return testTime }
return fi, nil
},
CloseFunc: func() error { return nil },
}

layerFile := &testfs.File{
WriteFunc: func(p []byte) (int, error) { return len(p), nil },
CloseFunc: func() error { return nil },
}

base := testfs.New(
testfs.WithOpen(func(name string) (ihfs.File, error) {
return baseFile, nil
}),
)

expectedErr := errors.New("chtimes failed")
layer := testfs.New(
testfs.WithStat(func(name string) (ihfs.FileInfo, error) {
fi := testfs.NewFileInfo(".")
fi.IsDirFunc = func() bool { return true }
return fi, nil
}),
testfs.WithCreate(func(name string) (ihfs.File, error) {
return layerFile, nil
}),
testfs.WithChtimes(func(name string, atime, mtime time.Time) error {
return expectedErr
}),
)

err := union.CopyToLayer(base, layer, "test.txt")

Expect(err).To(Equal(expectedErr))
})
})
})
