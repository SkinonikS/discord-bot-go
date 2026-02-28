package youtubeMusicPlayerSource

import (
	"io"
	"os/exec"
)

type Stream struct {
	rc  io.ReadCloser
	cmd *exec.Cmd
}

func (s *Stream) Read(b []byte) (int, error) {
	return s.rc.Read(b)
}

func (s *Stream) Close() error {
	if err := s.rc.Close(); err != nil {
		return err
	}
	if s.cmd.Process != nil {
		return s.cmd.Process.Kill()
	}
	return s.cmd.Wait()
}
