package player

import (
	"errors"
	"time"
)

var (
	ErrInvalidIndex = errors.New("invalid queue index")
)

type Track struct {
	Title       string
	URL         string
	Duration    time.Duration
	RequestedBy string
	Source      string
}

type Queue struct {
	tracks []*Track
}

func NewQueue() *Queue {
	return &Queue{}
}

func (q *Queue) Shift(track *Track) {
	q.tracks = append([]*Track{track}, q.tracks...)
}

func (q *Queue) Push(track *Track) {
	q.tracks = append(q.tracks, track)
}

func (q *Queue) Pop() *Track {
	if len(q.tracks) == 0 {
		return nil
	}
	track := q.tracks[0]
	q.tracks = q.tracks[1:]
	return track
}

func (q *Queue) Remove(index int) (*Track, error) {
	i := index - 1
	if i < 0 || i >= len(q.tracks) {
		return nil, ErrInvalidIndex
	}
	track := q.tracks[i]
	q.tracks = append(q.tracks[:i], q.tracks[i+1:]...)
	return track, nil
}

func (q *Queue) Snapshot() *Queue {
	tracks := make([]*Track, len(q.tracks))
	copy(tracks, q.tracks)
	return &Queue{tracks: tracks}
}

func (q *Queue) List() []*Track {
	return q.tracks
}

func (q *Queue) IsEmpty() bool {
	return len(q.tracks) == 0
}

func (q *Queue) Len() int {
	return len(q.tracks)
}

func (q *Queue) Clear() {
	q.tracks = q.tracks[:0]
}
