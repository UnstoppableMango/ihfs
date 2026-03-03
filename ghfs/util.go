package ghfs

import (
	"encoding/json"

	"github.com/google/go-github/v84/github"
	"github.com/unstoppablemango/ihfs"
)

func OpenOwner(fsys ihfs.FS, owner string) (*github.User, error) {
	return openDecode[github.User](fsys, ownerPath(owner))
}

func OpenRepository(fsys ihfs.FS, owner, repo string) (*github.Repository, error) {
	return openDecode[github.Repository](fsys, repoPath(owner, repo))
}

func OpenBranch(fsys ihfs.FS, owner, repo, branch string) (*github.Branch, error) {
	return openDecode[github.Branch](fsys, branchPath(owner, repo, branch))
}

func OpenContent(fsys ihfs.FS, owner, repo, ref, path string) (*github.RepositoryContent, error) {
	return openDecode[github.RepositoryContent](fsys, contentPath(owner, repo, ref, path))
}

func OpenRelease(fsys ihfs.FS, owner, repo, tag string) (*github.RepositoryRelease, error) {
	return openDecode[github.RepositoryRelease](fsys, releasePath(owner, repo, tag))
}

func openDecode[T any](fsys ihfs.FS, path string) (*T, error) {
	f, err := fsys.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()
	return decode[T](f)
}

func decode[T any](f ihfs.File) (*T, error) {
	var v T
	d := json.NewDecoder(f)
	if err := d.Decode(&v); err != nil {
		return nil, err
	}
	return &v, nil
}
