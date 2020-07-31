/*
Copyright © 2020 Ken'ichiro Oyama <k1lowxb@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/k1LoW/ghput/gh"
	"github.com/k1LoW/tbls-push/version"
	"github.com/k1LoW/tbls/datasource"
	"github.com/spf13/cobra"
)

var (
	owner     string
	repo      string
	branch    string
	namespace string
)

var rootCmd = &cobra.Command{
	Use:   "tbls-push",
	Short: "tbls-push is an external subcommand of tbls for pushing schema data (schema.json) to target GitHub repository",
	Long:  `tbls-push is an external subcommand of tbls for pushing schema data (schema.json) to target GitHub repository.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if owner == "" || repo == "" || branch == "" {
			return errors.New("`tbls push` need `--owner` AND `--repo` AND `--branch` flag")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		schema := os.Getenv("TBLS_SCHEMA")
		if schema == "" {
			printFatalln(cmd, errors.New("env `TBLS_SCHEMA` not found"))
		}
		s, err := datasource.AnalyzeJSONStringOrFile(schema)
		if err != nil {
			printFatalln(cmd, err)
		}
		b, err := ioutil.ReadFile(filepath.Clean(schema))
		if err != nil {
			printFatalln(cmd, err)
		}
		key := "tbls-push"
		message := "tbls-push"
		path := filepath.Join("tbls.d", namespace, fmt.Sprintf("%s.yml", s.Name))

		ctx := context.Background()
		g, err := gh.New(owner, repo, key)
		if err != nil {
			printFatalln(cmd, err)
		}
		if err := g.CommitAndPush(ctx, branch, string(b), path, message); err != nil {
			printFatalln(cmd, err)
		}
	},
}

func Execute() {
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)

	log.SetOutput(ioutil.Discard)
	if env := os.Getenv("DEBUG"); env != "" {
		debug, err := os.Create(fmt.Sprintf("%s.debug", version.Name))
		if err != nil {
			printFatalln(rootCmd, err)
		}
		log.SetOutput(debug)
	}

	if err := rootCmd.Execute(); err != nil {
		printFatalln(rootCmd, err)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&owner, "owner", "", "", "repository owner")
	rootCmd.Flags().StringVarP(&repo, "repo", "", "", "repository name")
	rootCmd.Flags().StringVarP(&branch, "branch", "", "master", "target branch")
	rootCmd.Flags().StringVarP(&namespace, "namespace", "", "", "namespace")
}

// https://github.com/spf13/cobra/pull/894
func printErrln(c *cobra.Command, i ...interface{}) {
	c.PrintErr(fmt.Sprintln(i...))
}

func printErrf(c *cobra.Command, format string, i ...interface{}) {
	c.PrintErr(fmt.Sprintf(format, i...))
}

func printFatalln(c *cobra.Command, i ...interface{}) {
	printErrln(c, i...)
	os.Exit(1)
}

func printFatalf(c *cobra.Command, format string, i ...interface{}) {
	printErrf(c, format, i...)
	os.Exit(1)
}
