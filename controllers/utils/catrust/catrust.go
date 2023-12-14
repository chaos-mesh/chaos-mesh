// Copyright 2023 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package catrust

import (
	"crypto/x509"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

func loadExtraCACerts(caCerts *x509.CertPool) error {
	extraCAPath, ok := os.LookupEnv("EXTRA_CA_TRUST_PATH")
	if !ok {
		return nil
	}

	err := filepath.WalkDir(extraCAPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}

		info, err := d.Info()
		if err != nil {
			return errors.Wrapf(err, "get info for path %q", path)
		}

		// follow symlinks - kubernetes secret volume mounts contain symlinks to the secret value files and
		// a directory containing the raw data. Following symlinks allows us to process only links to regular
		// files.
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			newPath, err := filepath.EvalSymlinks(path)
			if err != nil {
				return errors.Wrapf(err, "read symlink %q", path)
			}

			path = newPath
			info, err = os.Stat(path)
			if err != nil {
				return errors.Wrapf(err, "cannot stat %q", path)
			}
		}

		// filter directories, pipes, other irregular files
		if !info.Mode().IsRegular() {
			return nil
		}

		bytes, err := os.ReadFile(path)
		if err != nil {
			return errors.Wrapf(err, "read cert file %q", path)
		}

		ok := caCerts.AppendCertsFromPEM(bytes)
		if !ok {
			return errors.Errorf("parse PEM file %q", path)
		}
		return nil
	})

	if err != nil {
		return errors.Wrap(err, "load extra CA trust certificates")
	}

	return nil
}

type CACertLoader struct{}

func (cac CACertLoader) Load() (*x509.CertPool, error) {
	caCerts, _ := x509.SystemCertPool()

	if caCerts == nil {
		caCerts = x509.NewCertPool()
	}

	err := loadExtraCACerts(caCerts)
	if err != nil {
		return nil, errors.Wrap(err, "load extra CA certificates")
	}

	return caCerts, nil
}

func New() *CACertLoader {
	return &CACertLoader{}
}
