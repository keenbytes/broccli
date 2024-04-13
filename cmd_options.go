package broccli

type cmdOptions struct {
	onPostValidation func(c *Cmd) error
}

type cmdOption func(opts *cmdOptions)

func OnPostValidation(fn func(c *Cmd) error) cmdOption {
	return func(opts *cmdOptions) {
		opts.onPostValidation = fn
	}
}
