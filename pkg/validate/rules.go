package validate

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	giturls "github.com/whilp/git-urls"
)

const (
	maxURLRuneCount = 2083
	minURLRuneCount = 3
)

var (
	// ErrGitURL is the error that returns in case of an invalid Git clone URL.
	ErrGitURL = validation.NewError("validation_is_git_url", "must be a valid Git clone URL")

	// GitURL validates if a string is a valid Git clone URL
	GitURL = validation.NewStringRuleWithError(IsGitURL, ErrGitURL)
)

func IsGitURL(str string) bool {
	u, err := giturls.Parse(str)

	if err != nil {
		return false
	}

	if !giturls.NewTransportSet(
		"ssh",
		"git",
		"git+ssh",
		"http",
		"https",
		"ftp",
		"ftps",
		"rsync",
	).Valid(u.Scheme) {
		return false
	}

	return true
}
