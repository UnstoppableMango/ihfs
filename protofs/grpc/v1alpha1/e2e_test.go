package protofsv1alpha1_test

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	protofsv1alpha1 "github.com/unstoppablemango/ihfs/protofs/grpc/v1alpha1"
	"github.com/unstoppablemango/ihfs/testfs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func newClient(fsys ihfs.FS) (*protofsv1alpha1.Fs, func()) {
	server := grpc.NewServer()
	protofsv1alpha1.RegisterFsServer(server, fsys)
	protofsv1alpha1.RegisterFileServer(server, fsys)

	tmp := GinkgoT().TempDir()
	sock := filepath.Join(tmp, "fs.sock")

	lis, err := net.Listen("unix", sock)
	Expect(err).NotTo(HaveOccurred())

	go func() { _ = server.Serve(lis) }()

	conn, err := grpc.NewClient(fmt.Sprint("unix://", sock),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	Expect(err).NotTo(HaveOccurred())

	cleanup := func() {
		conn.Close()
		server.GracefulStop()
	}

	return protofsv1alpha1.New(conn), cleanup
}

var _ = Describe("E2e", func() {
	Describe("Open", func() {
		It("should open a file and read its content", func() {
			content := []byte("hello world")
			fsys := testfs.New(
				testfs.WithOpen(func(string) (ihfs.File, error) {
					return &testfs.File{
						StatFunc: func() (ihfs.FileInfo, error) {
							return testfs.NewFileInfo("test.txt"), nil
						},
						ReadFunc: func(p []byte) (int, error) {
							n := copy(p, content)
							return n, io.EOF
						},
					}, nil
				}),
			)

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			file, err := client.Open("test.txt")
			Expect(err).NotTo(HaveOccurred())

			data, err := io.ReadAll(file)
			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(Equal(content))
		})

		It("should return an error when the file does not exist", func() {
			fsys := testfs.New()

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			_, err := client.Open("missing.txt")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("File.Read", func() {
		It("should support multiple reads", func() {
			content := []byte("hello world")
			var readCount int
			fsys := testfs.New(
				testfs.WithOpen(func(string) (ihfs.File, error) {
					return &testfs.File{
						StatFunc: func() (ihfs.FileInfo, error) {
							return testfs.NewFileInfo("test.txt"), nil
						},
						ReadFunc: func(p []byte) (int, error) {
							readCount++
							n := copy(p, content)
							return n, io.EOF
						},
					}, nil
				}),
			)

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			file, err := client.Open("test.txt")
			Expect(err).NotTo(HaveOccurred())

			// First read: loads data from server
			buf := make([]byte, 5)
			n, err := file.Read(buf)
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(5))
			Expect(string(buf[:n])).To(Equal("hello"))

			// Second read: uses cached data
			n, err = file.Read(buf)
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(5))
			Expect(string(buf[:n])).To(Equal(" worl"))

			// Third read: remaining data
			n, err = file.Read(buf)
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(1))

			// EOF
			_, err = file.Read(buf)
			Expect(err).To(Equal(io.EOF))
		})
	})

	Describe("File.Close", func() {
		It("should close successfully", func() {
			fsys := testfs.New(
				testfs.WithOpen(func(string) (ihfs.File, error) {
					return &testfs.File{
						StatFunc: func() (ihfs.FileInfo, error) {
							return testfs.NewFileInfo("test.txt"), nil
						},
						ReadFunc: func([]byte) (int, error) { return 0, io.EOF },
					}, nil
				}),
			)

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			file, err := client.Open("test.txt")
			Expect(err).NotTo(HaveOccurred())

			Expect(file.Close()).To(Succeed())
		})
	})

	Describe("File.Stat", func() {
		It("should return file info", func() {
			fi := testfs.NewFileInfo("test.txt")
			fi.SizeFunc = func() int64 { return 100 }
			fsys := testfs.New(
				testfs.WithOpen(func(string) (ihfs.File, error) {
					return &testfs.File{
						StatFunc: func() (ihfs.FileInfo, error) { return fi, nil },
						ReadFunc: func([]byte) (int, error) { return 0, io.EOF },
					}, nil
				}),
			)

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			file, err := client.Open("test.txt")
			Expect(err).NotTo(HaveOccurred())

			stat, err := file.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(stat.Name()).To(Equal("test.txt"))
			Expect(stat.Size()).To(Equal(int64(100)))
		})
	})

	Describe("Stat", func() {
		It("should return file info", func() {
			fi := testfs.NewFileInfo("test.txt")
			fi.SizeFunc = func() int64 { return 200 }
			fsys := testfs.New(testfs.WithStat(func(string) (ihfs.FileInfo, error) {
				return fi, nil
			}))

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			info, err := client.Stat("test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Name()).To(Equal("test.txt"))
			Expect(info.Size()).To(Equal(int64(200)))
		})

		It("should propagate errors", func() {
			fsys := testfs.New()

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			_, err := client.Stat("missing.txt")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ReadDir", func() {
		It("should return directory entries", func() {
			fsys := testfs.New(testfs.WithReadDir(func(string) ([]ihfs.DirEntry, error) {
				return []ihfs.DirEntry{
					testfs.NewDirEntry("file.txt", false),
					testfs.NewDirEntry("subdir", true),
				}, nil
			}))

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			entries, err := client.ReadDir(".")
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(2))
			Expect(entries[0].Name()).To(Equal("file.txt"))
			Expect(entries[1].Name()).To(Equal("subdir"))
			Expect(entries[1].IsDir()).To(BeTrue())
		})

		It("should propagate errors", func() {
			client, cleanup := newClient(testfs.New())
			DeferCleanup(cleanup)

			_, err := client.ReadDir(".")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("File.ReadDir", func() {
		It("should return directory entries from a dir file", func() {
			fsys := testfs.New(
				testfs.WithOpen(func(string) (ihfs.File, error) {
					return &testfs.File{
						StatFunc: func() (ihfs.FileInfo, error) {
							fi := testfs.NewFileInfo(".")
							fi.IsDirFunc = func() bool { return true }
							return fi, nil
						},
						ReadFunc: func([]byte) (int, error) { return 0, io.EOF },
						ReadDirFunc: func(n int) ([]ihfs.DirEntry, error) {
							return []ihfs.DirEntry{
								testfs.NewDirEntry("a.txt", false),
							}, nil
						},
					}, nil
				}),
			)

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			file, err := client.Open(".")
			Expect(err).NotTo(HaveOccurred())

			dirFile := file.(interface {
				ReadDir(int) ([]fs.DirEntry, error)
			})
			entries, err := dirFile.ReadDir(-1)
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(1))
			Expect(entries[0].Name()).To(Equal("a.txt"))
		})
	})

	Describe("ReadFile", func() {
		It("should return file contents", func() {
			content := []byte("file contents")
			fsys := testfs.New(testfs.WithReadFile(func(string) ([]byte, error) {
				return content, nil
			}))

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			data, err := client.ReadFile("test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(Equal(content))
		})

		It("should propagate errors", func() {
			client, cleanup := newClient(testfs.New())
			DeferCleanup(cleanup)

			_, err := client.ReadFile("test.txt")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Glob", func() {
		It("should return matching paths", func() {
			fsys := testfs.New(testfs.WithGlob(func(string) ([]string, error) {
				return []string{"a.txt", "b.txt"}, nil
			}))

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			matches, err := client.Glob("*.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(matches).To(ConsistOf("a.txt", "b.txt"))
		})

		It("should propagate errors", func() {
			client, cleanup := newClient(testfs.New())
			DeferCleanup(cleanup)

			_, err := client.Glob("*.txt")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Create", func() {
		It("should create a file and return a handle", func() {
			fsys := testfs.New(testfs.WithCreate(func(string) (ihfs.File, error) {
				return &testfs.File{
					WriteFunc: func(p []byte) (int, error) { return len(p), nil },
				}, nil
			}))

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			file, err := client.Create("new.txt")
			Expect(err).NotTo(HaveOccurred())

			w := file.(io.Writer)
			n, err := w.Write([]byte("data"))
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(4))
		})

		It("should propagate errors", func() {
			client, cleanup := newClient(testfs.New())
			DeferCleanup(cleanup)

			_, err := client.Create("new.txt")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("File.WriteAt", func() {
		It("should write at offset", func() {
			fsys := testfs.New(testfs.WithCreate(func(string) (ihfs.File, error) {
				return &testfs.File{
					WriteAtFunc: func(p []byte, off int64) (int, error) {
						return len(p), nil
					},
				}, nil
			}))

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			file, err := client.Create("new.txt")
			Expect(err).NotTo(HaveOccurred())

			wa := file.(io.WriterAt)
			n, err := wa.WriteAt([]byte("data"), 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(4))
		})
	})

	Describe("File.Sync", func() {
		It("should sync the file", func() {
			synced := false
			fsys := testfs.New(
				testfs.WithOpen(func(string) (ihfs.File, error) {
					return &testfs.File{
						ReadFunc:  func([]byte) (int, error) { return 0, io.EOF },
						SyncFunc:  func() error { synced = true; return nil },
						StatFunc:  func() (ihfs.FileInfo, error) { return testfs.NewFileInfo("f"), nil },
					}, nil
				}),
			)

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			file, err := client.Open("f")
			Expect(err).NotTo(HaveOccurred())

			type syncer interface{ Sync() error }
			s := file.(syncer)
			Expect(s.Sync()).To(Succeed())
			Expect(synced).To(BeTrue())
		})
	})

	Describe("File.Truncate", func() {
		It("should truncate the file", func() {
			var truncatedSize int64
			fsys := testfs.New(
				testfs.WithOpen(func(string) (ihfs.File, error) {
					return &testfs.File{
						ReadFunc:     func([]byte) (int, error) { return 0, io.EOF },
						TruncateFunc: func(size int64) error { truncatedSize = size; return nil },
						StatFunc:     func() (ihfs.FileInfo, error) { return testfs.NewFileInfo("f"), nil },
					}, nil
				}),
			)

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			file, err := client.Open("f")
			Expect(err).NotTo(HaveOccurred())

			type truncater interface{ Truncate(int64) error }
			t := file.(truncater)
			Expect(t.Truncate(128)).To(Succeed())
			Expect(truncatedSize).To(Equal(int64(128)))
		})
	})

	Describe("WriteFile", func() {
		It("should write file contents", func() {
			var written []byte
			fsys := testfs.New(testfs.WithWriteFile(func(_ string, data []byte, _ ihfs.FileMode) error {
				written = data
				return nil
			}))

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			err := client.WriteFile("test.txt", []byte("content"), 0o644)
			Expect(err).NotTo(HaveOccurred())
			Expect(written).To(Equal([]byte("content")))
		})
	})

	Describe("Mkdir", func() {
		It("should create a directory", func() {
			var created string
			fsys := testfs.New(testfs.WithMkdir(func(name string, _ ihfs.FileMode) error {
				created = name
				return nil
			}))

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			Expect(client.Mkdir("newdir", 0o755)).To(Succeed())
			Expect(created).To(Equal("newdir"))
		})
	})

	Describe("MkdirAll", func() {
		It("should create directories along a path", func() {
			var created string
			fsys := testfs.New(testfs.WithMkdirAll(func(name string, _ ihfs.FileMode) error {
				created = name
				return nil
			}))

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			Expect(client.MkdirAll("a/b/c", 0o755)).To(Succeed())
			Expect(created).To(Equal("a/b/c"))
		})
	})

	Describe("Remove", func() {
		It("should remove a file", func() {
			var removed string
			fsys := testfs.New(testfs.WithRemove(func(name string) error {
				removed = name
				return nil
			}))

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			Expect(client.Remove("old.txt")).To(Succeed())
			Expect(removed).To(Equal("old.txt"))
		})
	})

	Describe("RemoveAll", func() {
		It("should remove a path and its children", func() {
			var removed string
			fsys := testfs.New(testfs.WithRemoveAll(func(name string) error {
				removed = name
				return nil
			}))

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			Expect(client.RemoveAll("olddir")).To(Succeed())
			Expect(removed).To(Equal("olddir"))
		})
	})

	Describe("Rename", func() {
		It("should rename a file", func() {
			var oldPath, newPath string
			fsys := testfs.New(testfs.WithRename(func(old, new string) error {
				oldPath, newPath = old, new
				return nil
			}))

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			Expect(client.Rename("old.txt", "new.txt")).To(Succeed())
			Expect(oldPath).To(Equal("old.txt"))
			Expect(newPath).To(Equal("new.txt"))
		})
	})

	Describe("Chmod", func() {
		It("should change file mode", func() {
			var changedMode fs.FileMode
			fsys := testfs.New(testfs.WithChmod(func(_ string, mode ihfs.FileMode) error {
				changedMode = mode
				return nil
			}))

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			Expect(client.Chmod("test.txt", 0o600)).To(Succeed())
			Expect(changedMode).To(Equal(fs.FileMode(0o600)))
		})
	})

	Describe("Chown", func() {
		It("should change file ownership", func() {
			var gotUID, gotGID int
			fsys := testfs.New(testfs.WithChown(func(_ string, uid, gid int) error {
				gotUID, gotGID = uid, gid
				return nil
			}))

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			Expect(client.Chown("test.txt", 1000, 1000)).To(Succeed())
			Expect(gotUID).To(Equal(1000))
			Expect(gotGID).To(Equal(1000))
		})
	})

	Describe("Chtimes", func() {
		It("should change file times", func() {
			var gotAtime, gotMtime time.Time
			fsys := testfs.New(testfs.WithChtimes(func(_ string, atime, mtime time.Time) error {
				gotAtime, gotMtime = atime, mtime
				return nil
			}))

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			now := time.Now().Truncate(time.Second)
			Expect(client.Chtimes("test.txt", now, now)).To(Succeed())
			Expect(gotAtime).To(BeTemporally("~", now, time.Second))
			Expect(gotMtime).To(BeTemporally("~", now, time.Second))
		})
	})

	Describe("Symlink", func() {
		It("should create a symlink", func() {
			var gotOld, gotNew string
			fsys := testfs.New(testfs.WithSymlink(func(old, new string) error {
				gotOld, gotNew = old, new
				return nil
			}))

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			Expect(client.Symlink("target", "link")).To(Succeed())
			Expect(gotOld).To(Equal("target"))
			Expect(gotNew).To(Equal("link"))
		})
	})

	Describe("ReadLink", func() {
		It("should return the link target", func() {
			fsys := testfs.New(testfs.WithReadLink(func(string) (string, error) {
				return "target", nil
			}))

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			target, err := client.ReadLink("link")
			Expect(err).NotTo(HaveOccurred())
			Expect(target).To(Equal("target"))
		})

		It("should propagate errors", func() {
			client, cleanup := newClient(testfs.New())
			DeferCleanup(cleanup)

			_, err := client.ReadLink("link")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Lstat", func() {
		It("should return FileInfo without following symlinks", func() {
			fi := testfs.NewFileInfo("link")
			fi.ModeFunc = func() fs.FileMode { return os.ModeSymlink | 0o777 }
			fsys := testfs.New(testfs.WithLstat(func(string) (ihfs.FileInfo, error) {
				return fi, nil
			}))

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			info, err := client.Lstat("link")
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Name()).To(Equal("link"))
			Expect(info.Mode() & os.ModeSymlink).NotTo(BeZero())
		})

		It("should propagate errors", func() {
			client, cleanup := newClient(testfs.New())
			DeferCleanup(cleanup)

			_, err := client.Lstat("link")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("File.Stat", func() {
		It("should propagate errors", func() {
			fsys := testfs.New(testfs.WithOpen(func(string) (ihfs.File, error) {
				return &testfs.File{
					StatFunc: func() (ihfs.FileInfo, error) { return nil, errors.New("stat error") },
				}, nil
			}))

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			file, err := client.Open("f")
			Expect(err).NotTo(HaveOccurred())

			_, err = file.Stat()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("File.ReadDir (error)", func() {
		It("should propagate errors", func() {
			fsys := testfs.New(testfs.WithOpen(func(string) (ihfs.File, error) {
				return &testfs.File{
					ReadDirFunc: func(int) ([]ihfs.DirEntry, error) {
						return nil, errors.New("readdir error")
					},
				}, nil
			}))

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			file, err := client.Open(".")
			Expect(err).NotTo(HaveOccurred())

			dirFile := file.(interface {
				ReadDir(int) ([]fs.DirEntry, error)
			})
			_, err = dirFile.ReadDir(-1)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("File.Write (error)", func() {
		It("should propagate errors when FS lacks CreateFS", func() {
			fsys := testfs.New(testfs.WithOpen(func(string) (ihfs.File, error) {
				return &testfs.File{}, nil
			}))

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			file, err := client.Open("f")
			Expect(err).NotTo(HaveOccurred())

			w := file.(interface{ Write([]byte) (int, error) })
			_, err = w.Write([]byte("data"))
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("File.WriteAt (error)", func() {
		It("should propagate errors when FS lacks CreateFS", func() {
			fsys := testfs.New(testfs.WithOpen(func(string) (ihfs.File, error) {
				return &testfs.File{}, nil
			}))

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			file, err := client.Open("f")
			Expect(err).NotTo(HaveOccurred())

			wa := file.(interface{ WriteAt([]byte, int64) (int, error) })
			_, err = wa.WriteAt([]byte("data"), 0)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("File.Read (load error)", func() {
		It("should propagate errors from the server", func() {
			fsys := testfs.New(testfs.WithOpen(func(string) (ihfs.File, error) {
				return &testfs.File{
					ReadFunc: func([]byte) (int, error) { return 0, errors.New("read error") },
				}, nil
			}))

			client, cleanup := newClient(fsys)
			DeferCleanup(cleanup)

			file, err := client.Open("f")
			Expect(err).NotTo(HaveOccurred())

			buf := make([]byte, 10)
			_, err = file.Read(buf)
			Expect(err).To(HaveOccurred())
		})
	})
})
