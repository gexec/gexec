package command

// tmplProjectEnvironmentValueShow represents a project environment value within details view.
var tmplProjectEnvironmentValueShow = "ID: \x1b[33m{{ .ID }} \x1b[0m" + `
Name: {{ .Name }}
Kind: {{ .Kind }}
Content: {{ .Content }}
`
