package apperror

import (
	"runtime"
	"strings"
)

func captureFrames(skip int) []Frame {
	frames := []Frame{}
	for i := skip; i < skip+10; i++ { // Capture up to 10 frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		frames = append(frames, Frame{
			File: file,
			Line: line,
			Func: fn.Name(),
		})

		// Stop at our package boundary
		if strings.Contains(file, "apperror") && i > skip {
			break
		}
	}
	return frames
}
