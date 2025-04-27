package validate

import (
	"github.com/adhocore/gronx"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	giturls "github.com/whilp/git-urls"
)

var (
	// ErrGitURL is the error that returns in case of an invalid Git clone URL.
	ErrGitURL = validation.NewError("validation_is_git_url", "must be a valid Git clone URL")

	// GitURL validates if a string is a valid Git clone URL
	GitURL = validation.NewStringRuleWithError(IsGitURL, ErrGitURL)

	// ErrCronSyntax is the error that returns in case of an invalid cron syntax.
	ErrCronSyntax = validation.NewError("validation_is_git_url", "must be a valid Git clone URL")

	// CronSyntax validates if a string is a valid cron syntax
	CronSyntax = validation.NewStringRuleWithError(gronx.IsValid, ErrGitURL)
)

// IsGitURL provides a vaöidator for Git clone URLs.
func IsGitURL(str string) bool {
	if len(str) >= 8000 {
		return false
	}

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
		"file",
	).Valid(u.Scheme) {
		return false
	}

	return true
}
