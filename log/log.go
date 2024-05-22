package log

// Stdout is the log type for STDOUT
const Stdout = "STDOUT"

// Stderr is the log type for STDERR
const Stderr = "STDERR"

// logStruct {

// Log represents a message that was created by a process,
// LogType is either "STDOUT" or "STDERR",
// Content is the byte contents of the message itself
type Log struct {
	LogType string
	Content []byte
}

// }
