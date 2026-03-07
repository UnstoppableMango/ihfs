package ghfs

import (
	"encoding/json"

	"github.com/google/go-github/v84/github"
	"github.com/unstoppablemango/ihfs"
)

func OpenOwner(fsys ihfs.FS, owner string) (*github.User, error) {
	p := &Path{owner: owner}
	return openDecode[github.User](fsys, p.ownerPath())
}

func OpenRepository(fsys ihfs.FS, owner, repo string) (*github.Repository, error) {
	p := &Path{owner: owner, repo: repo}
	return openDecode[github.Repository](fsys, p.repoPath())
}

func OpenBranch(fsys ihfs.FS, owner, repo, branch string) (*github.Branch, error) {
	p := &Path{owner: owner, repo: repo, branch: branch}
	return openDecode[github.Branch](fsys, p.branchPath())
}

func OpenContent(fsys ihfs.FS, owner, repo, ref, path string) (*github.RepositoryContent, error) {
	p := &Path{owner: owner, repo: repo, branch: ref, content: path}
	return openDecode[github.RepositoryContent](fsys, p.contentPath())
}

func OpenRelease(fsys ihfs.FS, owner, repo, tag string) (*github.RepositoryRelease, error) {
	p := &Path{owner: owner, repo: repo, tag: tag}
	return openDecode[github.RepositoryRelease](fsys, p.releasePath())
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
