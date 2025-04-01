package dnspod

import "github.com/pkg/errors"

func (c *Solver) SetLogLevel(level string) error {
	if err := c.logLevel.UnmarshalText([]byte(level)); err != nil {
		return errors.Wrap(err, "failed to parse log level, valid values are: debug, info, warn, error")
	}
	return nil
}

func (c *Solver) Error(err error, msg string, args ...any) {
	args = append(args, "error", err)
	c.log.Error(msg, args...)
}
