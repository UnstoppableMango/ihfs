package corfs

import "time"

type Option func(*Fs)

func WithCacheTime(cacheTime time.Duration) Option {
	return func(f *Fs) {
		f.cacheTime = cacheTime
	}
}
