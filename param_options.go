package broccli

type paramOptions struct {
	onTrue func(c *Cmd)
}

type paramOption func(opts *paramOptions)

func OnTrue(fn func(c *Cmd)) paramOption {
	return func(opts *paramOptions) {
		opts.onTrue = fn
	}
}
