package utils

import "time"

func ToDuration(ttsSteps int, sampleRate uint32, step int) time.Duration {
	if sampleRate <= 1 {
		return 0
	}
	return time.Duration(int(1000*float64(ttsSteps*step)/float64(sampleRate))) * time.Millisecond
}

func ToTTSSteps(duration time.Duration, sampleRate uint32, step int) int {
	if step <= 1 {
		return 0
	}
	return int(float64(duration.Milliseconds()) * float64(sampleRate) / float64(step*1000))
}
