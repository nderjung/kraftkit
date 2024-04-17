// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2024, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.

package cli_test

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2" //nolint:stylecheck
	. "github.com/onsi/gomega"    //nolint:stylecheck

	fcmd "kraftkit.sh/test/e2e/framework/cmd"
	fcfg "kraftkit.sh/test/e2e/framework/config"
)

// urlParser fetches the url from the output
func urlParser(stdout *fcmd.IOStream) string {
	if strings.Contains(stdout.String(), "domain") {
		url := strings.SplitN(stdout.String(), "domain\":\"", 2)[1]
		url = strings.SplitN(url, "\"", 2)[0]
		return "https://" + url
	}

	return ""
}

func serviceGroupParser(stdout *fcmd.IOStream) string {
	if strings.Contains(stdout.String(), "service group\":") {
		serviceGroup := strings.SplitN(stdout.String(), "service group\":\"", 2)[1]
		serviceGroup = strings.SplitN(serviceGroup, "\"", 2)[0]
		return serviceGroup
	}

	return ""
}

var _ = Describe("kraft cloud instance create", func() {
	var cmd *fcmd.Cmd

	var stdout *fcmd.IOStream
	var stderr *fcmd.IOStream

	var cfg *fcfg.Config

	const (
		imageName       = "nginx:latest"
		instanceName    = "instance-create-test"
		instanceMemory  = "64"
		instancePortMap = "443:8080"
	)

	BeforeEach(func() {
		if os.Getenv("KRAFTCLOUD_TOKEN") == "" {
			Skip("KRAFTCLOUD_TOKEN is not set")
		}

		if os.Getenv("KRAFTCLOUD_METRO") == "" {
			Skip("KRAFTCLOUD_METRO is not set")
		}

		stdout = fcmd.NewIOStream()
		stderr = fcmd.NewIOStream()

		cfg = fcfg.NewTempConfig()

		cmd = fcmd.NewKraft(stdout, stderr, cfg.Path())
		cmd.Env = os.Environ()
		cmd.Args = append(cmd.Args, "cloud", "instance", "create", "--log-level", "info", "--log-type", "json", "-o", "raw")
	})

	// General tests
	When("invoked without flags or positional arguments", func() {
		It("should error and print an error", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			Expect(err).To(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())

			Expect(stdout.String()).To(MatchRegexp(`requires at least 1 arg`))
		})
	})

	When("invoked with standard flags and positional arguments", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap,
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				"--start",
				imageName,
			)
		})

		It("should print the instance as running and the url should work", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`"state":"running"`))
			Expect(stdout.String()).To(MatchRegexp("\"image\":\"" + strings.SplitN(imageName, ":", 2)[0]))
			Expect(stdout.String()).To(MatchRegexp("\"memory\":\"" + instanceMemory + " MiB\""))

			url := urlParser(stdout)
			Expect(url).ToNot(BeEmpty())

			// Run the "curl" command to test the url
			curlCmd := fcmd.NewCurl(stdout, stderr)
			curlCmd.Args = append(curlCmd.Args, url)

			err = curlCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(curlCmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`Welcome to nginx!`))
		})

		AfterEach(func() {
			// Remove the instance after the test
			cleanCmd := fcmd.NewKraft(stdout, stderr, cfg.Path())
			cleanCmd.Args = append(
				cleanCmd.Args, "cloud", "instance", "delete",
				"--log-level", "info",
				"--log-type", "json",
				instanceNameFull,
			)

			err := cleanCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cleanCmd.DumpError(stdout, stderr, err))
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp("removing 1 instance(s)"))
		})
	})

	// '--help' flag tests
	When("invoked with the --help flag", func() {
		BeforeEach(func() {
			cmd.Args = append(cmd.Args, "--help")
		})

		It("should print the command's help", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`Create an instance on KraftCloud from an image.\n`))
		})
	})

	// '--memory' flag tests
	When("invoked with standard flags and positional arguments, but no memory", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap,
				"--name", instanceNameFull,
				"--start",
				imageName,
			)
		})

		It("should boot up with the default memory size (128 MiB) and respond to requests", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`"state":"running"`))
			Expect(stdout.String()).To(MatchRegexp("\"image\":\"" + strings.SplitN(imageName, ":", 2)[0]))
			Expect(stdout.String()).To(MatchRegexp("\"memory\":\"" + "128" + " MiB\""))

			url := urlParser(stdout)
			Expect(url).ToNot(BeEmpty())

			// Run the "curl" command to test the url
			curlCmd := fcmd.NewCurl(stdout, stderr)
			curlCmd.Args = append(curlCmd.Args, url)

			err = curlCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(curlCmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`Welcome to nginx!`))
		})

		AfterEach(func() {
			// Remove the instance after the test
			cleanCmd := fcmd.NewKraft(stdout, stderr, cfg.Path())
			cleanCmd.Args = append(cleanCmd.Args,
				"cloud", "instance", "delete",
				"--log-level", "info",
				"--log-type", "json",
				instanceNameFull,
			)

			err := cleanCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cleanCmd.DumpError(stdout, stderr, err))
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp("removing 1 instance(s)"))
		})
	})

	When("invoked with standard flags and positional arguments, and custom memory (57 MiB)", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap,
				"--name", instanceNameFull,
				"--memory", "57",
				"--start",
				imageName,
			)
		})

		It("should boot up with the custom memory size (57 MiB) and respond to requests", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`state":"running`))
			Expect(stdout.String()).To(MatchRegexp("image\":\"" + strings.SplitN(imageName, ":", 2)[0]))
			Expect(stdout.String()).To(MatchRegexp("memory\":\"" + "57 MiB\""))

			url := urlParser(stdout)
			Expect(url).ToNot(BeEmpty())

			// Run the "curl" command to test the url
			curlCmd := fcmd.NewCurl(stdout, stderr)
			curlCmd.Args = append(curlCmd.Args, url)

			err = curlCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(curlCmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`Welcome to nginx!`))
		})

		AfterEach(func() {
			// Remove the instance after the test
			cleanCmd := fcmd.NewKraft(stdout, stderr, cfg.Path())
			cleanCmd.Args = append(cleanCmd.Args,
				"cloud", "instance", "delete",
				"--log-level", "info",
				"--log-type", "json",
				instanceNameFull,
			)

			err := cleanCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cleanCmd.DumpError(stdout, stderr, err))
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp("removing 1 instance(s)"))
		})
	})

	When("invoked with standard flags and positional arguments, and negative memory (-16 MiB)", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap,
				"--name", instanceNameFull,
				"--memory", "-16",
				"--start",
				imageName,
			)
		})

		It("should error out from the API with a bounds error", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			Expect(err).To(HaveOccurred())

			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`Value of 'memory_mb' must be in the range 16 to 2048`))
		})
	})

	When("invoked with standard flags and positional arguments, and huge memory (10,000,000,000 MiB)", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap,
				"--name", instanceNameFull,
				"--memory", "10000000000",
				"--start",
				imageName,
			)
		})

		It("should error out from the API with a bounds error", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			Expect(err).To(HaveOccurred())

			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`Value of 'memory_mb' must be in the range 16 to 2048`))
		})
	})

	When("invoked with standard flags and positional arguments, and float memory (16.56 MiB)", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap,
				"--name", instanceNameFull,
				"--memory", "16.56",
				"--start",
				imageName,
			)
		})

		It("should error out from kraftkit with a parsing error", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			Expect(err).To(HaveOccurred())

			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`strconv.ParseInt: parsing \\"16.56\\": invalid syntax`))
		})
	})

	When("invoked with standard flags and positional arguments, and text memory (seventeen MiB)", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap,
				"--name", instanceNameFull,
				"--memory", "seventeen",
				"--start",
				imageName,
			)
		})

		It("should error out from kraftkit with a parsing error", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			Expect(err).To(HaveOccurred())

			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`strconv.ParseInt: parsing \\"seventeen\\": invalid syntax`))
		})
	})

	When("invoked with standard shorthand flags and positional arguments", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"-p", instancePortMap,
				"-n", instanceNameFull,
				"-M", instanceMemory,
				"-S",
				imageName,
			)
		})

		It("should boot up with the custom memory size (57 MiB) and respond to requests", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`"state":"running"`))
			Expect(stdout.String()).To(MatchRegexp("\"image\":\"" + strings.SplitN(imageName, ":", 2)[0]))
			Expect(stdout.String()).To(MatchRegexp("\"memory\":\"" + instanceMemory + " MiB\""))

			url := urlParser(stdout)
			Expect(url).ToNot(BeEmpty())

			// Run the "curl" command to test the url
			curlCmd := fcmd.NewCurl(stdout, stderr)
			curlCmd.Args = append(curlCmd.Args, url)

			err = curlCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(curlCmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`Welcome to nginx!`))
		})

		AfterEach(func() {
			// Remove the instance after the test
			cleanCmd := fcmd.NewKraft(stdout, stderr, cfg.Path())
			cleanCmd.Args = append(cleanCmd.Args,
				"cloud", "instance", "delete",
				"--log-level", "info",
				"--log-type", "json",
				instanceNameFull,
			)

			err := cleanCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cleanCmd.DumpError(stdout, stderr, err))
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp("removing 1 instance(s)"))
		})
	})

	// '--name' flag tests
	When("invoked with standard flags and positional arguments, but a 64 character name", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			for len(instanceNameFull) < 64 {
				instanceNameFull = instanceNameFull + "a"
			}

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap,
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				"--start",
				imageName,
			)
		})

		It("should error out with an API error", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			Expect(err).To(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`Invalid name`))
		})
	})

	When("invoked with standard flags and positional arguments, but a 62 character name, and a hyphen at the end", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			for len(instanceNameFull) < 62 {
				instanceNameFull = instanceNameFull + "a"
			}

			instanceNameFull = instanceNameFull + "-"

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap,
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				"--start",
				imageName,
			)
		})

		It("should error out with an API error", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			Expect(err).To(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`Invalid name`))
		})
	})

	When("invoked with standard flags and positional arguments, but a 62 character name, and a period at the end", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			for len(instanceNameFull) < 62 {
				instanceNameFull = instanceNameFull + "a"
			}

			instanceNameFull = instanceNameFull + "."

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap,
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				"--start",
				imageName,
			)
		})

		It("should error out with an API error", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			Expect(err).To(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`Invalid name`))
		})
	})

	When("invoked with standard flags and positional arguments, but a 62 character name with two consecutive hyphens", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			for len(instanceNameFull) < 59 {
				instanceNameFull = instanceNameFull + "a"
			}

			instanceNameFull = instanceNameFull + "--a"

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap,
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				"--start",
				imageName,
			)
		})

		It("should error out with an API error", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			Expect(err).To(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`Invalid name`))
		})
	})

	When("invoked with standard flags and positional arguments, but a 62 character name with two consecutive periods", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			for len(instanceNameFull) < 59 {
				instanceNameFull = instanceNameFull + "a"
			}

			instanceNameFull = instanceNameFull + "..a"

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap,
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				"--start",
				imageName,
			)
		})

		It("should error out with an API error", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			Expect(err).To(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`Invalid name`))
		})
	})

	When("invoked with standard flags and positional arguments, but a 62 character name starting with a digit", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("0%s-%d", instanceName, id)

			for len(instanceNameFull) < 59 {
				instanceNameFull = instanceNameFull + "a"
			}

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap,
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				"--start",
				imageName,
			)
		})

		It("should error out with an API error", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			Expect(err).To(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`Invalid name`))
		})
	})

	When("invoked with standard flags and positional arguments, but a 62 character name with uppercase letters", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			for len(instanceNameFull) < 59 {
				instanceNameFull = instanceNameFull + "A"
			}

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap,
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				"--start",
				imageName,
			)
		})

		It("should start correctly and set the name with lowercase", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(strings.ToLower(instanceNameFull)))
		})

		AfterEach(func() {
			// Remove the instance after the test
			cleanCmd := fcmd.NewKraft(stdout, stderr, cfg.Path())
			cleanCmd.Args = append(cleanCmd.Args,
				"cloud", "instance", "delete",
				"--log-level", "info",
				"--log-type", "json",
				strings.ToLower(instanceNameFull),
			)

			err := cleanCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cleanCmd.DumpError(stdout, stderr, err))
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(strings.Join([]string{"removing", strings.ToLower(instanceNameFull)}, " ")))
		})
	})

	When("invoked with standard flags and positional arguments, but a 60 character name ending in all ascii characters", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			for len(instanceNameFull) < 59 {
				instanceNameFull = instanceNameFull + "a"
			}
		})

		It("should error out with an API error", func() {
			// Do not test all of them, if it fails for all other ASCII charcaters
			// and for some random Unicode characters, it will probably fail for all of them

			for _, char := range " !\"#$%&'()*+,/:;<=>?@[\\]^_`{|}~ÑÓ" {
				instanceNameFullAppended := instanceNameFull + string(char) + "a"

				stdout = fcmd.NewIOStream()
				stderr = fcmd.NewIOStream()

				cfg = fcfg.NewTempConfig()

				cmd = fcmd.NewKraft(stdout, stderr, cfg.Path())
				cmd.Env = os.Environ()
				cmd.Args = append(cmd.Args,
					"cloud", "instance", "create",
					"--log-level", "info",
					"--log-type", "json",
					"-o", "raw",
				)

				cmd.Args = append(cmd.Args,
					"--port", instancePortMap,
					"--memory", instanceMemory,
					"--name", instanceNameFullAppended,
					"--start",
					imageName,
				)

				err := cmd.Run()
				time.Sleep(1 * time.Second)
				Expect(err).To(HaveOccurred())

				Expect(stderr.String()).To(BeEmpty())
				Expect(stdout.String()).ToNot(BeEmpty())
				Expect(stdout.String()).To(MatchRegexp(`Invalid name`))
			}
		})
	})

	// '--start' flag tests
	When("invoked with standard flags and positional arguments but no start flag", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap,
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				imageName,
			)
		})

		It("should print the instance as stopped and the url should not work", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`"state":"stopped"`))

			url := urlParser(stdout)
			Expect(url).ToNot(BeEmpty())

			// Run the "curl" command to test the url
			curlCmd := fcmd.NewCurl(stdout, stderr)

			curlCmd.Args = append(curlCmd.Args, url)

			err = curlCmd.Run()
			time.Sleep(1 * time.Second)
			Expect(err).To(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`There is no service on this URL.`))
		})

		AfterEach(func() {
			// Remove the instance after the test
			cleanCmd := fcmd.NewKraft(stdout, stderr, cfg.Path())
			cleanCmd.Args = append(cleanCmd.Args,
				"cloud", "instance", "delete",
				"--log-level", "info",
				"--log-type", "json",
				instanceNameFull,
			)

			err := cleanCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cleanCmd.DumpError(stdout, stderr, err))
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp("removing 1 instance(s)"))
		})
	})

	// '--port' flag tests
	When("invoked with standard flags and positional arguments, but no port flag", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				"--start",
				imageName,
			)
		})

		It("should work, but the instance should not be accessible", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`"state":"running"`))
			Expect(stdout.String()).To(MatchRegexp("\"image\":\"" + strings.SplitN(imageName, ":", 2)[0]))
			Expect(stdout.String()).To(MatchRegexp("\"memory\":\"" + instanceMemory + " MiB\""))

			url := urlParser(stdout)
			Expect(url).ToNot(BeEmpty())

			// Run the "curl" command to test the url
			curlCmd := fcmd.NewCurl(stdout, stderr)
			curlCmd.Args = append(curlCmd.Args, url)

			err = curlCmd.Run()
			time.Sleep(1 * time.Second)
			Expect(err).To(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			Expect(stdout.String()).To(MatchRegexp(`There is no service on this URL.`))
		})

		AfterEach(func() {
			// Remove the instance after the test
			cleanCmd := fcmd.NewKraft(stdout, stderr, cfg.Path())
			cleanCmd.Args = append(cleanCmd.Args,
				"cloud", "instance", "delete",
				"--log-level", "info",
				"--log-type", "json",
				instanceNameFull,
			)

			err := cleanCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cleanCmd.DumpError(stdout, stderr, err))
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp("removing 1 instance(s)"))
		})
	})

	When("invoked with standard flags and positional arguments, and multiple port flags", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap+"/http+tls",
				"--port", "8081:8080/tls",
				"--port", "8082:8080/tls",
				"--port", "8083:8080/tls",
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				imageName,
			)
		})

		It("should show all up in the service group details", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`"state":"stopped"`))

			serviceGroup := serviceGroupParser(stdout)
			Expect(serviceGroup).ToNot(BeEmpty())

			serviceCmd := fcmd.NewKraft(stdout, stderr, cfg.Path())
			serviceCmd.Args = append(serviceCmd.Args,
				"cloud", "service", "get",
				"--log-level", "info",
				"--log-type", "json",
				"-o", "raw",
				serviceGroup,
			)

			err = serviceCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(serviceCmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())

			Expect(stdout.String()).To(MatchRegexp(`8081:8080/tls`))
			Expect(stdout.String()).To(MatchRegexp(`8082:8080/tls`))
			Expect(stdout.String()).To(MatchRegexp(`8083:8080/tls`))
		})

		AfterEach(func() {
			// Remove the instance after the test
			cleanCmd := fcmd.NewKraft(stdout, stderr, cfg.Path())
			cleanCmd.Args = append(cleanCmd.Args,
				"cloud", "instance", "delete",
				"--log-level", "info",
				"--log-type", "json",
				instanceNameFull,
			)

			err := cleanCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cleanCmd.DumpError(stdout, stderr, err))
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp("removing 1 instance(s)"))
		})
	})

	When("invoked with standard flags and positional arguments, and a fully specified port '443:8080/tls+http'", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap+"/http+tls",
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				"--start",
				imageName,
			)
		})

		It("should print the instance as running and the url should work", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`"state":"running"`))
			Expect(stdout.String()).To(MatchRegexp("\"image\":\"" + strings.SplitN(imageName, ":", 2)[0]))
			Expect(stdout.String()).To(MatchRegexp("\"memory\":\"" + instanceMemory + " MiB\""))

			url := urlParser(stdout)
			Expect(url).ToNot(BeEmpty())

			// Run the "curl" command to test the url
			curlCmd := fcmd.NewCurl(stdout, stderr)
			curlCmd.Args = append(curlCmd.Args, url)

			err = curlCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(curlCmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`Welcome to nginx!`))
		})

		AfterEach(func() {
			// Remove the instance after the test
			cleanCmd := fcmd.NewKraft(stdout, stderr, cfg.Path())
			cleanCmd.Args = append(cleanCmd.Args,
				"cloud", "instance", "delete",
				"--log-level", "info",
				"--log-type", "json",
				instanceNameFull,
			)

			err := cleanCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cleanCmd.DumpError(stdout, stderr, err))
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp("removing 1 instance(s)"))
		})
	})

	When("invoked with standard flags and positional arguments, and an out of bounds external port", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--port", "123123:8080/tls",
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				imageName,
			)
		})

		It("should error out with an API error", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			Expect(err).To(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			Expect(stdout.String()).To(MatchRegexp(`Value of 'port' in service port description must be in the range 1 to 65535`))
		})

		AfterEach(func() {
			// Remove the instance after the test
			cleanCmd := fcmd.NewKraft(stdout, stderr, cfg.Path())
			cleanCmd.Args = append(cleanCmd.Args,
				"cloud", "instance", "delete",
				"--log-level", "info",
				"--log-type", "json",
				instanceNameFull,
			)

			err := cleanCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cleanCmd.DumpError(stdout, stderr, err))
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp("removing 1 instance(s)"))
		})
	})

	When("invoked with standard flags and positional arguments, and an out of bounds interal port", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--port", "8080:123123/tls",
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				imageName,
			)
		})

		It("should error out with an API error", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			Expect(err).To(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			Expect(stdout.String()).To(MatchRegexp(`Value of 'destination_port' in service port description must be in the range 1 to 65535`))
		})

		AfterEach(func() {
			// Remove the instance after the test
			cleanCmd := fcmd.NewKraft(stdout, stderr, cfg.Path())
			cleanCmd.Args = append(cleanCmd.Args,
				"cloud", "instance", "delete",
				"--log-level", "info",
				"--log-type", "json",
				instanceNameFull,
			)

			err := cleanCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cleanCmd.DumpError(stdout, stderr, err))
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp("removing 1 instance(s)"))
		})
	})

	When("invoked with standard flags and positional arguments, and an negative port", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--port", "8080:-8081/tls",
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				imageName,
			)
		})

		It("should error out with an API error", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			Expect(err).To(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			Expect(stdout.String()).To(MatchRegexp(`Value of 'destination_port' in service port description must be in the range 1 to 65535`))
		})

		AfterEach(func() {
			// Remove the instance after the test
			cleanCmd := fcmd.NewKraft(stdout, stderr, cfg.Path())
			cleanCmd.Args = append(cleanCmd.Args,
				"cloud", "instance", "delete",
				"--log-level", "info",
				"--log-type", "json",
				instanceNameFull,
			)

			err := cleanCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cleanCmd.DumpError(stdout, stderr, err))
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp("removing 1 instance(s)"))
		})
	})

	When("invoked with standard flags and positional arguments, and two negative ports", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--port", "-8080:-8081/tls",
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				imageName,
			)
		})

		It("should error out with an API error", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			Expect(err).To(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			Expect(stdout.String()).To(MatchRegexp(`Value of 'port' in service port description must be in the range 1 to 65535`))
		})

		AfterEach(func() {
			// Remove the instance after the test
			cleanCmd := fcmd.NewKraft(stdout, stderr, cfg.Path())
			cleanCmd.Args = append(cleanCmd.Args,
				"cloud", "instance", "delete",
				"--log-level", "info",
				"--log-type", "json",
				instanceNameFull,
			)

			err := cleanCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cleanCmd.DumpError(stdout, stderr, err))
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp("removing 1 instance(s)"))
		})
	})

	When("invoked with standard flags and positional arguments, and a text port", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--port", "8080:::123/tls",
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				imageName,
			)
		})

		It("should error out with a Kraftkit error", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			Expect(err).To(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			Expect(stdout.String()).To(MatchRegexp(`invalid --port value expected --port EXTERNAL:INTERNAL`))
		})
	})

	When("invoked with standard flags and positional arguments, and a malformed port format", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--port", "8080:::123/tls",
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				imageName,
			)
		})

		It("should error out with a Kraftkit error", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			Expect(err).To(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			Expect(stdout.String()).To(MatchRegexp(`invalid --port value expected --port EXTERNAL:INTERNAL`))
		})
	})

	// '--replicas' flag tests
	When("invoked with standard flags and positional arguments, and one replica", func() {
		var instanceNameFull string
		var replicaNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap,
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				"--replicas", "1",
				"--start",
				imageName,
			)
		})

		It("should report information about a single replica, but 'instance ls' should show two", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`"state":"running"`))
			Expect(stdout.String()).To(MatchRegexp("\"image\":\"" + strings.SplitN(imageName, ":", 2)[0]))
			Expect(stdout.String()).To(MatchRegexp("\"memory\":\"" + instanceMemory + " MiB\""))

			url := urlParser(stdout)
			Expect(url).ToNot(BeEmpty())

			// Run the "curl" command to test the url
			curlCmd := fcmd.NewCurl(stdout, stderr)
			curlCmd.Args = append(curlCmd.Args, url)

			err = curlCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(curlCmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`Welcome to nginx!`))

			// Run the "instance ls" command to test the number of replicas
			lsCmd := fcmd.NewKraft(stdout, stderr, cfg.Path())
			lsCmd.Args = append(lsCmd.Args,
				"cloud", "instance", "ls",
				"--log-level", "info",
				"--log-type", "json",
				"-o", "raw",
			)

			err = lsCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(lsCmd.DumpError(stdout, stderr, err))
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			Expect(stdout.String()).To(MatchRegexp("\"name\":\"" + instanceNameFull))

			// Second instance has a suffix starting with `-`
			Expect(stdout.String()).To(MatchRegexp("\"name\":\"" + instanceNameFull + "-"))

			seps := strings.Split(stdout.String(), instanceNameFull+"-")
			Expect(len(seps)).To(Equal(2))

			replicaNameFull = instanceNameFull + "-" + strings.SplitN(seps[1], "\"", 2)[0]
		})

		AfterEach(func() {
			// Remove the instance after the test
			cleanCmd := fcmd.NewKraft(stdout, stderr, cfg.Path())
			cleanCmd.Args = append(cleanCmd.Args,
				"cloud", "instance", "delete",
				"--log-level", "info",
				"--log-type", "json",
				instanceNameFull,
			)

			err := cleanCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cleanCmd.DumpError(stdout, stderr, err))
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp("removing 1 instance(s)"))

			// Remove the replica after the test
			cleanReplicaCmd := fcmd.NewKraft(stdout, stderr, cfg.Path())
			cleanReplicaCmd.Args = append(
				cleanCmd.Args, "cloud", "instance", "delete",
				"--log-level", "info",
				"--log-type", "json",
				"-o", "raw",
				replicaNameFull,
			)

			err = cleanReplicaCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cleanCmd.DumpError(stdout, stderr, err))
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(strings.Join([]string{"removing", replicaNameFull}, " ")))
		})
	})

	When("invoked with standard flags and positional arguments, and zero replicas", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap,
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				"--replicas", "0",
				"--start",
				imageName,
			)
		})

		It("should report information about no replicas, and 'instance ls' should show only one instance", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`"state":"running"`))
			Expect(stdout.String()).To(MatchRegexp("\"image\":\"" + strings.SplitN(imageName, ":", 2)[0]))
			Expect(stdout.String()).To(MatchRegexp("\"memory\":\"" + instanceMemory + " MiB\""))

			url := urlParser(stdout)
			Expect(url).ToNot(BeEmpty())

			// Run the "curl" command to test the url
			curlCmd := fcmd.NewCurl(stdout, stderr)
			curlCmd.Args = append(curlCmd.Args, url)

			err = curlCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(curlCmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`Welcome to nginx!`))

			// Run the "instance ls" command to test the number of replicas
			lsCmd := fcmd.NewKraft(stdout, stderr, cfg.Path())
			lsCmd.Args = append(
				lsCmd.Args, "cloud", "instance", "ls",
				"--log-level", "info",
				"--log-type", "json",
				"-o", "raw",
			)

			err = lsCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(lsCmd.DumpError(stdout, stderr, err))
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			Expect(stdout.String()).To(MatchRegexp("\"name\":\"" + instanceNameFull))

			// Second instance has a suffix starting with `-`
			Expect(stdout.String()).ToNot(MatchRegexp("\"name\":\"" + instanceNameFull + "-"))
		})

		AfterEach(func() {
			// Remove the instance after the test
			cleanCmd := fcmd.NewKraft(stdout, stderr, cfg.Path())
			cleanCmd.Args = append(cleanCmd.Args,
				"cloud", "instance", "delete",
				"--log-level", "info",
				"--log-type", "json",
				instanceNameFull,
			)

			err := cleanCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cleanCmd.DumpError(stdout, stderr, err))
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp("removing 1 instance(s)"))
		})
	})

	When("invoked with standard flags and positional arguments, and -1 replicas", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				"--replicas", "-1",
				imageName,
			)
		})

		It("should error out with an API error", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			Expect(err).To(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			Expect(stdout.String()).To(MatchRegexp(`Value of 'replicas' in instance description must be in the range `))
		})
	})

	When("invoked with standard flags and positional arguments, and 'one' replica", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				"--replicas", "one",
				imageName,
			)
		})

		It("should error out with a KraftKit error", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			Expect(err).To(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			Expect(stdout.String()).To(MatchRegexp(`strconv.ParseInt: parsing \\"one\\": invalid syntax`))
		})
	})

	When("invoked with standard flags and positional arguments, and 2.65 replicas", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				"--replicas", "2.64",
				imageName,
			)
		})

		It("should error out with a KraftKit error", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			Expect(err).To(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			Expect(stdout.String()).To(MatchRegexp(`strconv.ParseInt: parsing \\"2.64\\": invalid syntax`))
		})
	})

	// '--domain' flag tests
	When("invoked with standard flags and positional arguments, and a custom domain", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap,
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				"--domain", "smth-"+instanceNameFull,
				"--start",
				imageName,
			)
		})

		It("should show the instance as running and with the custom domain", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`"state":"running"`))
			Expect(stdout.String()).To(MatchRegexp("\"image\":\"" + strings.SplitN(imageName, ":", 2)[0]))
			Expect(stdout.String()).To(MatchRegexp("\"domain\":\"" + "smth-" + instanceNameFull))
		})

		AfterEach(func() {
			// Remove the instance after the test
			cleanCmd := fcmd.NewKraft(stdout, stderr, cfg.Path())
			cleanCmd.Args = append(cleanCmd.Args,
				"cloud", "instance", "delete",
				"--log-level", "info",
				"--log-type", "json",
				instanceNameFull,
			)

			err := cleanCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cleanCmd.DumpError(stdout, stderr, err))
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp("removing 1 instance(s)"))
		})
	})

	When("invoked with standard flags and positional arguments, and a domain with spaces", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--domain", "smth else",
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				imageName,
			)
		})

		It("should error out with an API error", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			Expect(err).To(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			Expect(stdout.String()).To(MatchRegexp(`Invalid DNS name`))
		})
	})

	When("invoked with standard flags and positional arguments, and a domain with special characters", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--domain", "șmth",
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				imageName,
			)
		})

		It("should error out with an API error", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			Expect(err).To(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			Expect(stdout.String()).To(MatchRegexp(`Invalid DNS name`))
		})
	})

	When("invoked with standard flags and positional arguments, and a 128-character long domain", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--domain", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				imageName,
			)
		})

		It("should error out with an API error", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			Expect(err).To(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			Expect(stdout.String()).To(MatchRegexp(`Invalid DNS name`))
		})
	})

	// '--scale-to-zero' flag tests
	When("invoked with standard flags and positional arguments, and the scale to zero flag", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap,
				"--name", instanceNameFull,
				"--scale-to-zero",
				imageName,
			)
		})

		It("should show the instance as being in standby and it should respond to requests", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`"state":"standby"`))
			Expect(stdout.String()).To(MatchRegexp("\"image\":\"" + strings.SplitN(imageName, ":", 2)[0]))
			Expect(stdout.String()).To(MatchRegexp("\"memory\":\"" + "128" + " MiB\""))

			url := urlParser(stdout)
			Expect(url).ToNot(BeEmpty())

			// Run the "curl" command to test the url
			curlCmd := fcmd.NewCurl(stdout, stderr)
			curlCmd.Args = append(curlCmd.Args, url)

			err = curlCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(curlCmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`Welcome to nginx!`))
		})

		AfterEach(func() {
			// Remove the instance after the test
			cleanCmd := fcmd.NewKraft(stdout, stderr, cfg.Path())
			cleanCmd.Args = append(cleanCmd.Args,
				"cloud", "instance", "delete",
				"--log-level", "info",
				"--log-type", "json",
				instanceNameFull,
			)

			err := cleanCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cleanCmd.DumpError(stdout, stderr, err))
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp("removing 1 instance(s)"))
		})
	})

	// '--env' flag tests
	When("invoked with standard flags and positional arguments, and a single env flag", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap,
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				"--env", "SOME_ENV=smth",
				"--start",
				imageName,
			)
		})

		It("should enable set the env flag in the machine and run it", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`"state":"running"`))
			Expect(stdout.String()).To(MatchRegexp("\"image\":\"" + strings.SplitN(imageName, ":", 2)[0]))
			Expect(stdout.String()).To(MatchRegexp("\"memory\":\"" + instanceMemory + " MiB\""))

			url := urlParser(stdout)
			Expect(url).ToNot(BeEmpty())

			// Run the "curl" command to test the url
			curlCmd := fcmd.NewCurl(stdout, stderr)
			curlCmd.Args = append(curlCmd.Args, url)

			err = curlCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(curlCmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`Welcome to nginx!`))
		})

		AfterEach(func() {
			// Remove the instance after the test
			cleanCmd := fcmd.NewKraft(stdout, stderr, cfg.Path())
			cleanCmd.Args = append(cleanCmd.Args,
				"cloud", "instance", "delete",
				"--log-level", "info",
				"--log-type", "json",
				instanceNameFull,
			)

			err := cleanCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cleanCmd.DumpError(stdout, stderr, err))
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp("removing 1 instance(s)"))
		})
	})

	When("invoked with standard flags and positional arguments, and a two env flags", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap,
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				"--env", "SOME_ENV=smth",
				"--env", "SOME_OTHER_ENV=smth",
				"--start",
				imageName,
			)
		})

		It("should set both envs in the machine and run it", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`"state":"running"`))
			Expect(stdout.String()).To(MatchRegexp("\"image\":\"" + strings.SplitN(imageName, ":", 2)[0]))
			Expect(stdout.String()).To(MatchRegexp("\"memory\":\"" + instanceMemory + " MiB\""))

			url := urlParser(stdout)
			Expect(url).ToNot(BeEmpty())

			// Run the "curl" command to test the url
			curlCmd := fcmd.NewCurl(stdout, stderr)
			curlCmd.Args = append(curlCmd.Args, url)

			err = curlCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(curlCmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`Welcome to nginx!`))
		})

		AfterEach(func() {
			// Remove the instance after the test
			cleanCmd := fcmd.NewKraft(stdout, stderr, cfg.Path())
			cleanCmd.Args = append(cleanCmd.Args,
				"cloud", "instance", "delete",
				"--log-level", "info",
				"--log-type", "json",
				instanceNameFull,
			)

			err := cleanCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cleanCmd.DumpError(stdout, stderr, err))
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp("removing 1 instance(s)"))
		})
	})

	When("invoked with standard flags and positional arguments, and a special character env flag", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap,
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				"--env", "ȘOME_ENV_🎉=șmth🎉",
				"--start",
				imageName,
			)
		})

		It("should set the env flag in the machine and run it", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`"state":"running"`))
			Expect(stdout.String()).To(MatchRegexp("\"image\":\"" + strings.SplitN(imageName, ":", 2)[0]))
			Expect(stdout.String()).To(MatchRegexp("\"memory\":\"" + instanceMemory + " MiB\""))

			url := urlParser(stdout)
			Expect(url).ToNot(BeEmpty())

			// Run the "curl" command to test the url
			curlCmd := fcmd.NewCurl(stdout, stderr)
			curlCmd.Args = append(curlCmd.Args, url)

			err = curlCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(curlCmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`Welcome to nginx!`))
		})

		AfterEach(func() {
			// Remove the instance after the test
			cleanCmd := fcmd.NewKraft(stdout, stderr, cfg.Path())
			cleanCmd.Args = append(cleanCmd.Args,
				"cloud", "instance", "delete",
				"--log-level", "info",
				"--log-type", "json",
				instanceNameFull,
			)

			err := cleanCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cleanCmd.DumpError(stdout, stderr, err))
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp("removing 1 instance(s)"))
		})
	})

	When("invoked with standard flags and positional arguments, and a long env flag", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			twoKiBLongEnv := ""
			for i := 0; i < 2048; i++ {
				twoKiBLongEnv += "a"
			}

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap,
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				"--env", "SOME_LONG_ENV="+twoKiBLongEnv,
				"--start",
				imageName,
			)
		})

		It("should work somehow, but the image will be stopped", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`"state":"stopped"`))
			Expect(stdout.String()).To(MatchRegexp("\"image\":\"" + strings.SplitN(imageName, ":", 2)[0]))
			Expect(stdout.String()).To(MatchRegexp("\"memory\":\"" + instanceMemory + " MiB\""))
		})

		AfterEach(func() {
			// Remove the instance after the test
			cleanCmd := fcmd.NewKraft(stdout, stderr, cfg.Path())
			cleanCmd.Args = append(cleanCmd.Args,
				"cloud", "instance", "delete",
				"--log-level", "info",
				"--log-type", "json",
				instanceNameFull,
			)

			err := cleanCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cleanCmd.DumpError(stdout, stderr, err))
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp("removing 1 instance(s)"))
		})
	})

	// '--volumes' flag tests
	When("invoked with standard flags and positional arguments, and a single volume", func() {
		BeforeEach(func() {
		})

		It("should link the volume to the instance and run it", func() {
			// TODO
		})

		AfterEach(func() {
		})
	})

	When("invoked with standard flags and positional arguments, and a two volumes", func() {
		BeforeEach(func() {
		})

		It("should link the volumes to the instance and run it", func() {
			// TODO
		})

		AfterEach(func() {
		})
	})

	When("invoked with standard flags and positional arguments, and a non-existent volume", func() {
		BeforeEach(func() {
		})

		It("should error out with an API error", func() {
			// TODO
		})

		AfterEach(func() {
		})
	})

	When("invoked with standard flags and positional arguments, and a special character volume", func() {
		BeforeEach(func() {
		})

		It("should error out with an API error", func() {
			// TODO
		})

		AfterEach(func() {
		})
	})

	// '--service-group' flag tests
	When("invoked with standard flags and positional arguments, and a service group", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				"--service-group", "smth-"+instanceNameFull,
				"--start",
				imageName,
			)

			createServiceCmd := fcmd.NewKraft(stdout, stderr, cfg.Path())

			createServiceCmd.Args = append(
				createServiceCmd.Args, "cloud", "service", "create",
				"--log-level", "info",
				"--log-type", "json",
				"-o", "raw",
				"--name", "smth-"+instanceNameFull,
				"443:8080/tls+http",
			)

			err = createServiceCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(createServiceCmd.DumpError(stdout, stderr, err))
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
		})

		It("should attach the instance to that service group", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`"state":"running"`))
			Expect(stdout.String()).To(MatchRegexp("\"image\":\"" + strings.SplitN(imageName, ":", 2)[0]))
			Expect(stdout.String()).To(MatchRegexp("\"memory\":\"" + instanceMemory + " MiB\""))

			url := urlParser(stdout)
			Expect(url).ToNot(BeEmpty())

			// Run the "curl" command to test the url
			curlCmd := fcmd.NewCurl(stdout, stderr)
			curlCmd.Args = append(curlCmd.Args, url)

			err = curlCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(curlCmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`Welcome to nginx!`))
		})

		AfterEach(func() {
			// Remove the instance after the test
			cleanCmd := fcmd.NewKraft(stdout, stderr, cfg.Path())
			cleanCmd.Args = append(cleanCmd.Args,
				"cloud", "instance", "delete",
				"--log-level", "info",
				"--log-type", "json",
				instanceNameFull,
			)

			err := cleanCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cleanCmd.DumpError(stdout, stderr, err))
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp("removing 1 instance(s)"))

			// Remove the service after the test
			cleanServiceCmd := fcmd.NewKraft(stdout, stderr, cfg.Path())
			cleanServiceCmd.Args = append(cleanServiceCmd.Args, "cloud", "service", "delete",
				"--log-level", "info", "--log-type", "json",
				"smth-"+instanceNameFull,
			)

			err = cleanServiceCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cleanServiceCmd.DumpError(stdout, stderr, err))
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(strings.Join([]string{"removing", "smth-" + instanceNameFull}, " ")))
		})
	})

	When("invoked with standard flags and positional arguments, and a non-existent group", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap,
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				"--service-group", "smth",
				"--start",
				imageName,
			)
		})

		It("should error with an API error", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			Expect(err).To(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`No service group with name 'smth`))
		})
	})

	When("invoked with standard flags and positional arguments, and a special-character group", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap,
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				"--service-group", "șmth",
				"--start",
				imageName,
			)
		})

		It("should error with an API error", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			Expect(err).To(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`Invalid name 'șmth'`))
		})
	})

	// '--feature' flag tests
	When("invoked with standard flags and positional arguments, and the scale to zero feature", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap,
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				"--feature", "scale-to-zero",
				imageName,
			)
		})

		It("should enable scale to zero and work", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`"state":"standby"`))
			Expect(stdout.String()).To(MatchRegexp("\"image\":\"" + strings.SplitN(imageName, ":", 2)[0]))
			Expect(stdout.String()).To(MatchRegexp("\"memory\":\"" + instanceMemory + " MiB\""))

			url := urlParser(stdout)
			Expect(url).ToNot(BeEmpty())

			// Run the "curl" command to test the url
			curlCmd := fcmd.NewCurl(stdout, stderr)
			curlCmd.Args = append(curlCmd.Args, url)

			err = curlCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(curlCmd.DumpError(stdout, stderr, err))
			}
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`Welcome to nginx!`))
		})

		AfterEach(func() {
			// Remove the instance after the test
			cleanCmd := fcmd.NewKraft(stdout, stderr, cfg.Path())
			cleanCmd.Args = append(cleanCmd.Args,
				"cloud", "instance", "delete",
				"--log-level", "info",
				"--log-type", "json",
				instanceNameFull,
			)

			err := cleanCmd.Run()
			time.Sleep(1 * time.Second)
			if err != nil {
				fmt.Print(cleanCmd.DumpError(stdout, stderr, err))
			}

			Expect(err).ToNot(HaveOccurred())
			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp("removing 1 instance(s)"))
		})
	})

	When("invoked with standard flags and positional arguments, and a random string feature", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap,
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				"--feature", "smth",
				"--start",
				imageName,
			)
		})

		It("should error with an API error", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			Expect(err).To(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`'smth' is not a valid flag for 'features'`))
		})
	})

	When("invoked with standard flags and positional arguments, and a special character string feature", func() {
		var instanceNameFull string

		BeforeEach(func() {
			id, err := rand.Int(rand.Reader, big.NewInt(100000000000))
			if err != nil {
				panic(err)
			}
			instanceNameFull = fmt.Sprintf("%s-%d", instanceName, id)

			cmd.Args = append(cmd.Args,
				"--port", instancePortMap,
				"--memory", instanceMemory,
				"--name", instanceNameFull,
				"--feature", "șmth",
				"--start",
				imageName,
			)
		})

		It("should error with an API error", func() {
			err := cmd.Run()
			time.Sleep(1 * time.Second)
			Expect(err).To(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())
			Expect(stdout.String()).To(MatchRegexp(`'șmth' is not a valid flag for 'features'`))
		})
	})
})
