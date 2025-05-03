package broccli

type paramOptions struct {
	onTrue func(c *Command)
}

type paramOption func(opts *paramOptions)

func OnTrue(fn func(c *Command)) paramOption {
	return func(opts *paramOptions) {
		opts.onTrue = fn
	}
}
