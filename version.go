package scanver

import (
	"bufio"
	"context"
	"io"
	"strings"

	"github.com/BurntSushi/toml"
)

// LookupPackageVersions looks up pkgName package versions from repo.
//
// It checks go.sum file, then, when there is no go.sum, it also checks Gopkg.lock file.
func (c *Client) LookupPackageVersions(ctx context.Context, repo *Repository, pkgName string) ([]string, error) {
	vs1, err := c.lookupPackageVersionsFromGoSum(ctx, repo, pkgName)
	if err == nil {
		return vs1, nil
	}

	if err.Error() != "No file named go.sum found in /" {
		return nil, err
	}

	vs2, err := c.lookupPackageVersionsFromGopkgLock(ctx, repo, pkgName)
	if err == nil {
		return vs2, nil
	}

	if err.Error() != "No file named Gopkg.lock found in /" {
		return nil, err
	}

	return nil, nil
}

func (c *Client) lookupPackageVersionsFromGoSum(ctx context.Context, repo *Repository, pkgName string) ([]string, error) {
	r, err := c.Repositories.DownloadContents(ctx, repo.Owner, repo.Name, "/go.sum", nil)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	return ReadPackageVersionsFromGoSum(r, pkgName)
}

// ReadPackageVersionsFromGoSum reads pkgName package version from r.
//
// r expects io.Reader for go.sum file.
func ReadPackageVersionsFromGoSum(r io.Reader, pkgName string) ([]string, error) {
	reader := bufio.NewReader(r)
	vset := make(map[string]struct{})
	for {
		l, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		line := string(l)
		parts := strings.Split(strings.TrimSpace(line), " ")
		if len(parts) != 3 {
			continue
		}
		if parts[0] != pkgName {
			continue
		}

		v := strings.TrimSuffix(parts[1], "/go.mod")
		vset[v] = struct{}{}
	}

	var vs []string
	for v := range vset {
		vs = append(vs, v)
	}

	return vs, nil
}

func (c *Client) lookupPackageVersionsFromGopkgLock(ctx context.Context, repo *Repository, pkgName string) ([]string, error) {
	r, err := c.Repositories.DownloadContents(ctx, repo.Owner, repo.Name, "/Gopkg.lock", nil)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	return ReadPackageVersionsFromGopkgLock(r, pkgName)
}

// ReadPackageVersionsFromGopkgLock reads pkgName package version from r.
//
// r expects io.Reader for Gopkg.lock file.
func ReadPackageVersionsFromGopkgLock(r io.Reader, pkgName string) ([]string, error) {
	type Project struct {
		Name     string `toml:"name"`
		Revision string `toml:"revision"`
		Version  string `toml:"version"`
	}
	type Lock struct {
		Projects []*Project `toml:"projects"`
	}

	var lock Lock
	_, err := toml.DecodeReader(r, &lock)
	if err != nil {
		return nil, err
	}

	var vs []string

	for _, p := range lock.Projects {
		if p.Name == pkgName {
			if len(p.Version) > 0 {
				vs = append(vs, p.Version)
			} else {
				vs = append(vs, p.Revision)
			}
		}
	}

	return vs, nil
}
