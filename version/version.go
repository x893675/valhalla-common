package version

import "fmt"

var (
	BuildTag     string
	BuildBranch  string
	BuildDate    string
	CommitSHA    string
	CommitAuthor string
)

func Info() string {
	return fmt.Sprintf("Version: %s, Branch: %s, Date: %s, Commit: %s, Author: %s",
		BuildTag, BuildBranch, BuildDate, CommitSHA, CommitAuthor)
}
