package player

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"os/exec"
	"strconv"
	"sync"

	"gopkg.in/hraban/opus.v2"
)

const (
	opusChannels   = 2
	opusSampleRate = 48000
	opusFrameSize  = 960
	opusMaxBytes   = (opusFrameSize * opusChannels) * 2
	ffmpegBufSize  = 16384
)

type FfmpegEncoder struct {
	ffmpegCmd   *exec.Cmd
	ffmpegBuf   *bufio.Reader
	opusEncoder *opus.Encoder
	opusPool    *sync.Pool
	audioBuf    []int16
}

func NewFfmpegEncoder(audioStream io.ReadCloser, ffmpegPath string) (*FfmpegEncoder, error) {
	opusEncoder, err := opus.NewEncoder(opusSampleRate, opusChannels, opus.AppAudio)
	if err != nil {
		return nil, err
	}

	ffmpegCmd := exec.Command(ffmpegPath,
		"-i", "pipe:0",
		"-f", "s16le",
		"-ar", strconv.Itoa(opusSampleRate),
		"-ac", strconv.Itoa(opusChannels),
		"pipe:1",
	)
	ffmpegCmd.Stdin = audioStream

	ffmpegOut, err := ffmpegCmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	ffmpegBuf := bufio.NewReaderSize(ffmpegOut, ffmpegBufSize)

	return &FfmpegEncoder{
		ffmpegCmd:   ffmpegCmd,
		ffmpegBuf:   ffmpegBuf,
		opusEncoder: opusEncoder,
		audioBuf:    make([]int16, opusFrameSize*opusChannels),
		opusPool: &sync.Pool{
			New: func() any {
				return make([]byte, opusMaxBytes)
			},
		},
	}, nil
}

func (e *FfmpegEncoder) Start() error {
	return e.ffmpegCmd.Start()
}

func (e *FfmpegEncoder) Stop() error {
	if e.ffmpegCmd.Process != nil {
		_ = e.ffmpegCmd.Process.Kill()
	}
	return e.ffmpegCmd.Wait()
}

func (e *FfmpegEncoder) Encode() ([]byte, error) {
	if err := binary.Read(e.ffmpegBuf, binary.LittleEndian, e.audioBuf); err != nil {
		if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, nil
		}
		return nil, err
	}

	opusBuf := e.opusPool.Get().([]byte)
	n, err := e.opusEncoder.Encode(e.audioBuf, opusBuf)
	if err != nil {
		e.opusPool.Put(opusBuf)
		return nil, err
	}

	frame := make([]byte, n)
	copy(frame, opusBuf[:n])
	e.opusPool.Put(opusBuf)

	return frame, nil
}
