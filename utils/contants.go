package utils

type ContextKey string

const (
	Connection ContextKey = "connection"
	Commit     ContextKey = "commit"
	Prod       ContextKey = "prod"
	PatchName  ContextKey = "patchName"
)
