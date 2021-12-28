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
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

type SshTunnel struct {
	config *ssh.ClientConfig
	host   string
	port   string
	client *ssh.Client
}

func NewSshTunnel(ip, port string, user, privateKeyFile string) (*SshTunnel, error) {
	hostKey, err := getHostKey(ip)
	if err != nil {
		return nil, err
	}
	config := ssh.ClientConfig{
		Timeout: 5 * time.Minute,
		User:    user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(func() ([]ssh.Signer, error) {
				key, err := ioutil.ReadFile(privateKeyFile)
				if err != nil {
					return nil, errors.Wrap(err, "ssh key file read failed")
				}

				signer, err := ssh.ParsePrivateKey(key)
				if err != nil {
					return nil, errors.Wrap(err, "ssh key signer failed")
				}
				return []ssh.Signer{signer}, nil
			}),
			ssh.PasswordCallback(func() (secret string, err error) {
				fmt.Printf("please input the password: ")
				password, err := term.ReadPassword(int(syscall.Stdin))
				if err != nil {
					return "", errors.Wrap(err, "read ssh password failed")
				}
				return string(password), nil
			}),
		},
		HostKeyCallback: ssh.FixedHostKey(hostKey),
	}
	return &SshTunnel{
		config: &config,
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

func (s *SshTunnel) SFTP(filename string, data []byte) error {
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

	if _, err := f.Write(data); err != nil {
		return errors.Wrapf(err, "write file %s failed", filename)
	}
	return nil
}

func getHostKey(host string) (ssh.PublicKey, error) {
	file, err := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var hostKey ssh.PublicKey
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), " ")
		if len(fields) != 3 {
			continue
		}
		if strings.Contains(fields[0], host) {
			var err error
			hostKey, _, _, _, err = ssh.ParseAuthorizedKey(scanner.Bytes())
			if err != nil {
				return nil, fmt.Errorf("error parsing %q: %v", fields[2], err)
			}
			break
		}
	}

	if hostKey == nil {
		return nil, fmt.Errorf("no hostkey for %s", host)
	}
	return hostKey, nil
}
