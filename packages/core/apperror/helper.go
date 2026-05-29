package apperror

import (
	"runtime"
	"strings"

	"go.uber.org/zap"
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

// ParseAppErrorIntoZapFields converts an AppError into a slice of zap fields
func ParseAppErrorIntoZapFields(appErr *AppError) []zap.Field {
	fields := []zap.Field{}

	// Add common fields
	fields = append(fields, zap.String("code", string(appErr.Code)))
	fields = append(fields, zap.String("request_id", appErr.RequestID))

	// Add details
	for _, detail := range appErr.Details {
		fields = append(fields, zap.String(detail.Field, detail.Message))
	}

	// Add stack frames
	fields = append(fields, zap.Any("frames", appErr.Frames))

	return fields
}
