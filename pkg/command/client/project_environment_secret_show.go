package command

// tmplProjectEnvironmentSecretShow represents a project environment secret within details view.
var tmplProjectEnvironmentSecretShow = "ID: \x1b[33m{{ .ID }} \x1b[0m" + `
Name: {{ .Name }}
Kind: {{ .Kind }}
Content: {{ .Content }}
`
