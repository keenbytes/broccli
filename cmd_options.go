package broccli

type commandOptions struct {
	onPostValidation func(c *Command) error
}

type commandOption func(opts *commandOptions)

func OnPostValidation(fn func(c *Command) error) commandOption {
	return func(opts *commandOptions) {
		opts.onPostValidation = fn
	}
}
