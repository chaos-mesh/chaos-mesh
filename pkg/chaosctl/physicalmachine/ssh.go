// Copyright 2021 Chaos Mesh Authors.
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

package physicalmachine

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/term"
)

type SshTunnel struct {
	config *ssh.ClientConfig
	host   string
	port   string
	client *ssh.Client
}

func readPassword() (string, error) {
	fmt.Printf("please input the password: \n")
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", errors.Wrap(err, "read ssh password failed")
	}
	return string(password), nil
}

func getSshTunnelConfig(user, privateKeyFile string, usePassword, ignoreKnowHost bool) (*ssh.ClientConfig, error) {

	config := &ssh.ClientConfig{
		Timeout: 5 * time.Minute,
		User:    user,
		Auth:    []ssh.AuthMethod{},
	}
	if ignoreKnowHost {
		config.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	} else {
		hostKeyCallback, err := knownhosts.New(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))
		if err != nil {
			return nil, err
		}
		config.HostKeyCallback = hostKeyCallback
	}

	if usePassword {
		password, err := readPassword()
		if err != nil {
			return nil, err
		}
		passwordCallBack := ssh.PasswordCallback(func() (secret string, err error) {
			return password, nil
		})
		config.Auth = append(config.Auth, passwordCallBack)
	} else {
		keyCallBack := ssh.PublicKeysCallback(func() ([]ssh.Signer, error) {
			key, err := os.ReadFile(privateKeyFile)
			if err != nil {
				return nil, errors.Wrap(err, "ssh key file read failed")
			}
			signer, err := ssh.ParsePrivateKey(key)
			if err != nil {
				return nil, errors.Wrap(err, "ssh key signer failed")
			}
			return []ssh.Signer{signer}, nil
		})
		config.Auth = append(config.Auth, keyCallBack)
	}
	return config, nil
}

func NewSshTunnel(ip, port string, config *ssh.ClientConfig) (*SshTunnel, error) {
	return &SshTunnel{
		config: config,
		host:   ip,
		port:   port,
	}, nil
}

func (s *SshTunnel) Open() error {
	conn, err := ssh.Dial("tcp", net.JoinHostPort(s.host, s.port), s.config)
	if err != nil {
		return errors.Wrap(err, "open ssh tunnel failed")
	}
	s.client = conn
	return nil
}

func (s *SshTunnel) Close() error {
	if s.client == nil {
		return nil
	}
	return s.client.Close()
}

func (s *SshTunnel) SFTP(filename string, r io.Reader) error {
	if s.client == nil {
		return errors.New("tunnel is not opened")
	}

	// open an SFTP session over an existing ssh connection.
	client, err := sftp.NewClient(s.client)
	if err != nil {
		return errors.Wrap(err, "create sftp client failed")
	}
	defer client.Close()

	if err := client.MkdirAll(filepath.Dir(filename)); err != nil {
		return errors.Wrapf(err, "make directory %s failed", filepath.Dir(filename))
	}

	f, err := client.Create(filename)
	if err != nil {
		return errors.Wrapf(err, "create file %s failed", filename)
	}
	defer f.Close()

	buf := make([]byte, 1024)
	for {
		if n, _ := r.Read(buf); n == 0 {
			break
		}
		if _, err := f.Write(buf); err != nil {
			return errors.Wrapf(err, "write file %s failed", filename)
		}
	}
	return nil
}

func (s *SshTunnel) Exec(cmd string) error {
	if s.client == nil {
		return errors.New("tunnel is not opened")
	}
	session, err := s.client.NewSession()
	if err != nil {
		return errors.Wrap(err, "create cmd  session failed")
	}
	defer session.Close()

	var output bytes.Buffer
	session.Stdout = &output
	if err := session.Run(cmd); err != nil {
		return errors.Wrapf(err, "exec cmd %s failed", cmd)
	}
	return nil
}
