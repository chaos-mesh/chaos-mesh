// Copyright 2020 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("webhook config", func() {
	Context("Test webhook config", func() {
		It("should return cfg.Injections", func() {
			configDir := "/etc/webhook/conf"
			config, err := LoadConfigDirectory(configDir)
			Expect(err).To(BeNil())
			Expect(config.AnnotationNamespace).To(Equal(annotationNamespaceDefault))
		})

		It("shoud return error on loading injection", func() {
			configFile := "fake file"
			InjectionConfig, err := LoadInjectionConfigFromFilePath(configFile)
			Expect(InjectionConfig).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("error loading injection config"))
		})

		It("shoud return yaml error on loading injection", func() {

			err := ioutil.WriteFile("/tmp/wrong.yaml", []byte(`fake yaml`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/wrong.yaml")

			configFile := "/tmp/wrong.yaml"
			InjectionConfig, err := LoadInjectionConfigFromFilePath(configFile)

			Expect(InjectionConfig).To(BeNil())
			Expect(err).ToNot(BeNil())
		})

		It("shoud return ErrMissingName on loading injection", func() {

			err := ioutil.WriteFile("/tmp/MissingName.yaml", []byte(``), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/MissingName.yaml")

			configFile := "/tmp/MissingName.yaml"
			InjectionConfig, err := LoadInjectionConfigFromFilePath(configFile)

			Expect(InjectionConfig).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(ErrMissingName))
		})

		It("shoud return not a valid name or name:version format on loading injection", func() {

			err := ioutil.WriteFile("/tmp/MissingName.yaml", []byte(`name: "testname:test:test:test"`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/MissingName.yaml")

			configFile := "/tmp/MissingName.yaml"
			InjectionConfig, err := LoadInjectionConfigFromFilePath(configFile)

			Expect(InjectionConfig).To(BeNil())
			Expect(err).ToNot(BeNil())
			fmt.Println(err)
		})

		It("shoud return nil on loading injection", func() {

			err := ioutil.WriteFile("/tmp/Name.yaml", []byte(`name: "testname"`), 0755)
			Expect(err).To(BeNil())
			defer os.Remove("/tmp/Name.yaml")

			configFile := "/tmp/Name.yaml"
			InjectionConfig, err := LoadInjectionConfigFromFilePath(configFile)

			Expect(InjectionConfig).ToNot(BeNil())
			Expect(InjectionConfig.Name).To(Equal("testname"))
			Expect(err).To(BeNil())
		})

		It("should return testname and defaultVersion on configNameFields", func() {
			shortName := "testname"
			name, version, err := configNameFields(shortName)
			Expect(name).To(Equal("testname"))
			Expect(version).To(Equal(defaultVersion))
			Expect(err).To(BeNil())
		})

		It("should return testname and defaultVersion on configNameFields", func() {
			shortName := "testname:"
			name, version, err := configNameFields(shortName)
			Expect(name).To(Equal("testname"))
			Expect(version).To(Equal(defaultVersion))
			Expect(err).To(BeNil())
		})

		It("should return testname and testversion on configNameFields", func() {
			shortName := "testname:testversion"
			name, version, err := configNameFields(shortName)
			Expect(name).To(Equal("testname"))
			Expect(version).To(Equal("testversion"))
			Expect(err).To(BeNil())
		})

		It("should return error on configNameFields", func() {
			shortName := "not:valid:name:"
			name, version, err := configNameFields(shortName)
			Expect(name).To(Equal(""))
			Expect(version).To(Equal(""))
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("not a valid name or name:version format"))
		})

		It("should return error on LoadInjectionConfig", func() {
			f, _ := os.Open("fake file")
			InjectionConfig, err := LoadInjectionConfig(f)
			Expect(InjectionConfig).To(BeNil())
			Expect(err).ToNot(BeNil())
		})

		It("should return error on GetRequestedConfig", func() {
			var cfg Config
			InjectionConfig, err := cfg.GetRequestedConfig("not:valid:name")
			Expect(InjectionConfig).To(BeNil())
			Expect(err).ToNot(BeNil())
		})

		It("should return no injection config found on GetRequestedConfig", func() {
			var cfg Config
			InjectionConfig, err := cfg.GetRequestedConfig("testname:testversion")
			Expect(InjectionConfig).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("no injection config found"))
		})

		It("should return for unit test on GetRequestedConfig", func() {
			var cfg Config
			ifg := &InjectionConfig{
				Name: "for unit test",
			}
			cfg.Injections = make(map[string]*InjectionConfig)
			cfg.Injections["testname:testversion"] = ifg
			InjectionConfig, err := cfg.GetRequestedConfig("testname:testversion")
			Expect(InjectionConfig).To(Equal(ifg))
			Expect(err).To(BeNil())
		})

		It("should return nil on ReplaceInjectionConfigs", func() {
			var cfg Config
			ifg := &InjectionConfig{
				Name:    "testname_origin",
				version: "testversion_origin",
			}
			cfg.Injections = make(map[string]*InjectionConfig)
			cfg.Injections["testname_origin:testversion_origin"] = ifg

			var replacementConfigs []*InjectionConfig = []*InjectionConfig{
				&InjectionConfig{
					Name:    "testname_after",
					version: "testversion_after",
				}}

			cfg.ReplaceInjectionConfigs(replacementConfigs)
			_, ok := cfg.Injections["testname_after:testversion_after"]
			Expect(ok).To(Equal(true))
		})

		It("should return defaultVersion on InjectionConfig.Version()", func() {
			var ifg InjectionConfig
			version := ifg.Version()
			Expect(version).To(Equal(defaultVersion))
		})

		It("should return testVersion on Version", func() {
			var ifg InjectionConfig
			ifg.version = "testVersion"
			version := ifg.Version()
			Expect(version).To(Equal("testVersion"))
		})

		It("should return request on RequestAnnotationKey", func() {
			var cfg Config
			res := cfg.RequestAnnotationKey()
			Expect(res).To(Equal("/request"))
		})

		It("should return status on StatusAnnotationKey", func() {
			var cfg Config
			res := cfg.StatusAnnotationKey()
			Expect(res).To(Equal("/status"))
		})

		It("should return init-request on RequestInitAnnotationKey", func() {
			var cfg Config
			res := cfg.RequestInitAnnotationKey()
			Expect(res).To(Equal("/init-request"))
		})

	})
})
