package compose

const maxWidth = 72

// Options controls compose-prep behavior.
type Options struct {
	InjectCcBcc bool
}

// Prepare normalizes an aerc compose buffer. On any processing error,
// the original input is returned unchanged.
func Prepare(input []byte, opts Options) []byte {
	return input
}
