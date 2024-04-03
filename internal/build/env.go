// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package build

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// These flags override values in build env.
	GitCommitFlag     = flag.String("git-commit", "", `Overrides git commit hash embedded into executables`)
	GitBranchFlag     = flag.String("git-branch", "", `Overrides git branch being built`)
	GitTagFlag        = flag.String("git-tag", "", `Overrides git tag being built`)
	BuildnumFlag      = flag.String("buildnum", "", `Overrides CI build number`)
	PullRequestFlag   = flag.Bool("pull-request", false, `Overrides pull request status of the build`)
	CronJobFlag       = flag.Bool("cron-job", false, `Overrides cron job status of the build`)
	UbuntuVersionFlag = flag.String("ubuntu", "", `Sets the ubuntu version being built for`)
)

// Environment contains metadata provided by the build environment.
type Environment struct {
	CI                        bool
	Name                      string // name of the environment
	Repo                      string // name of GitHub repo
	Commit, Date, Branch, Tag string // Git info
	Buildnum                  string
	UbuntuVersion             string // Ubuntu version being built for
	IsPullRequest             bool
	IsCronJob                 bool
}

func (env Environment) String() string {
	return fmt.Sprintf("%s env (commit:%s date:%s branch:%s tag:%s buildnum:%s pr:%t)",
		env.Name, env.Commit, env.Date, env.Branch, env.Tag, env.Buildnum, env.IsPullRequest)
}

// Env returns metadata about the current CI environment, falling back to LocalEnv
// if not running on CI.
func Env() Environment {
	switch {
	case os.Getenv("CI") == "true" && os.Getenv("CIRCLECI") == "true":
		commit := os.Getenv("CIRCLE_SHA1")
		return Environment{
			CI:            true,
			Name:          "circleci",
			Repo:          os.Getenv("CIRCLE_PROJECT_REPONAME"),
			Commit:        os.Getenv("CIRCLE_SHA1"),
			Date:          getDate(commit),
			Branch:        os.Getenv("CIRCLE_BRANCH"),
			Tag:           os.Getenv("CIRCLE_TAG"),
			Buildnum:      os.Getenv("CIRCLE_BUILD_NUM"),
			IsPullRequest: os.Getenv("CIRCLE_PR_NUMBER") != "",
			// NOTE(rgeraldes24): no cron jobs for now + circle ci does not have an env var for this field
			IsCronJob: false,
		}
	default:
		return LocalEnv()
	}
}

// LocalEnv returns build environment metadata gathered from git.
func LocalEnv() Environment {
	env := applyEnvFlags(Environment{Name: "local", Repo: "theQRL/go-zond"})

	head := readGitFile("HEAD")
	if fields := strings.Fields(head); len(fields) == 2 {
		head = fields[1]
	} else {
		// In this case we are in "detached head" state
		// see: https://git-scm.com/docs/git-checkout#_detached_head
		// Additional check required to verify, that file contains commit hash
		commitRe, _ := regexp.Compile("^([0-9a-f]{40})$")
		if commit := commitRe.FindString(head); commit != "" && env.Commit == "" {
			env.Commit = commit
		}
		return env
	}
	if env.Commit == "" {
		env.Commit = readGitFile(head)
	}
	env.Date = getDate(env.Commit)
	if env.Branch == "" {
		if head != "HEAD" {
			env.Branch = strings.TrimPrefix(head, "refs/heads/")
		}
	}
	if info, err := os.Stat(".git/objects"); err == nil && info.IsDir() && env.Tag == "" {
		env.Tag = firstLine(RunGit("tag", "-l", "--points-at", "HEAD"))
	}
	return env
}

func firstLine(s string) string {
	return strings.Split(s, "\n")[0]
}

func getDate(commit string) string {
	if commit == "" {
		return ""
	}
	out := RunGit("show", "-s", "--format=%ct", commit)
	if out == "" {
		return ""
	}
	date, err := strconv.ParseInt(strings.TrimSpace(out), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("failed to parse git commit date: %v", err))
	}
	return time.Unix(date, 0).Format("20060102")
}

func applyEnvFlags(env Environment) Environment {
	if !flag.Parsed() {
		panic("you need to call flag.Parse before Env or LocalEnv")
	}
	if *GitCommitFlag != "" {
		env.Commit = *GitCommitFlag
	}
	if *GitBranchFlag != "" {
		env.Branch = *GitBranchFlag
	}
	if *GitTagFlag != "" {
		env.Tag = *GitTagFlag
	}
	if *BuildnumFlag != "" {
		env.Buildnum = *BuildnumFlag
	}
	if *PullRequestFlag {
		env.IsPullRequest = true
	}
	if *CronJobFlag {
		env.IsCronJob = true
	}
	if *UbuntuVersionFlag != "" {
		env.UbuntuVersion = *UbuntuVersionFlag
	}
	return env
}
