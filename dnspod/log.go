package dnspod

import "github.com/pkg/errors"

func (s *Solver) SetLogLevel(level string) error {
	if err := s.logLevel.UnmarshalText([]byte(level)); err != nil {
		return errors.Wrap(err, "failed to parse log level, valid values are: debug, info, warn, error")
	}
	return nil
}

func (s *Solver) Error(err error, msg string, args ...any) {
	args = append(args, "error", err)
	s.log.Error(msg, args...)
}
