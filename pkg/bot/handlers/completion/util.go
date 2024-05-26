package completion

type HandlerMode int

const (
	NewCompletion HandlerMode = iota // Initialize the first value with iota
	EditCompletion
	RegenerateCompletion
)
