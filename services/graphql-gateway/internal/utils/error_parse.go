package utils

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"google.golang.org/grpc/codes"
)

func parseWrappedGRPCError(message string) *AppError {
	code, desc, ok := splitWrappedGRPCMessage(message)
	if !ok {
		return nil
	}

	return &AppError{
		Message: sanitizeClientMessage(desc),
		Code:    mapGRPCCode(parseGRPCCode(code)),
	}
}

func splitWrappedGRPCMessage(message string) (code string, desc string, ok bool) {
	const marker = "rpc error: code = "
	start := strings.Index(strings.ToLower(message), marker)
	if start == -1 {
		return "", "", false
	}

	segment := message[start+len(marker):]
	codePart, descPart, found := strings.Cut(segment, " desc = ")
	if !found {
		return "", "", false
	}

	code = strings.TrimSpace(codePart)
	desc = strings.TrimSpace(descPart)
	if code == "" || desc == "" {
		return "", "", false
	}

	return code, desc, true
}

func parseGRPCCode(value string) codes.Code {
	switch strings.TrimSpace(value) {
	case codes.InvalidArgument.String():
		return codes.InvalidArgument
	case codes.FailedPrecondition.String():
		return codes.FailedPrecondition
	case codes.OutOfRange.String():
		return codes.OutOfRange
	case codes.Unauthenticated.String():
		return codes.Unauthenticated
	case codes.PermissionDenied.String():
		return codes.PermissionDenied
	case codes.NotFound.String():
		return codes.NotFound
	case codes.AlreadyExists.String():
		return codes.AlreadyExists
	case codes.Aborted.String():
		return codes.Aborted
	case codes.ResourceExhausted.String():
		return codes.ResourceExhausted
	case codes.DeadlineExceeded.String():
		return codes.DeadlineExceeded
	case codes.Canceled.String():
		return codes.Canceled
	case codes.Unimplemented.String():
		return codes.Unimplemented
	case codes.Unavailable.String():
		return codes.Unavailable
	case codes.Unknown.String():
		return codes.Unknown
	default:
		return codes.Unknown
	}
}

func sanitizeClientMessage(message string) string {
	message = strings.TrimSpace(message)
	if message == "" {
		return "Internal server error"
	}

	if parsed := parseWrappedGRPCError(message); parsed != nil {
		return parsed.Message
	}

	message = trimTransportPrefix(message)
	return formatClientMessage(message)
}

func trimTransportPrefix(message string) string {
	normalized := strings.TrimSpace(message)

	for _, prefix := range []string{"input:", "request:"} {
		if !strings.HasPrefix(strings.ToLower(normalized), prefix) {
			continue
		}

		normalized = strings.TrimSpace(normalized[len(prefix):])
		return trimOperationPrefix(normalized)
	}

	return normalized
}

func trimOperationPrefix(message string) string {
	parts := strings.SplitN(message, " ", 2)
	if len(parts) < 2 {
		return message
	}

	op := strings.TrimSpace(parts[0])
	if op == "" || strings.Contains(op, ":") {
		return message
	}

	return strings.TrimSpace(parts[1])
}

func formatClientMessage(message string) string {
	message = strings.Join(strings.Fields(message), " ")
	if message == "" {
		return "Internal server error"
	}

	r, size := utf8.DecodeRuneInString(message)
	if r == utf8.RuneError && size == 0 {
		return "Internal server error"
	}

	return string(unicode.ToUpper(r)) + message[size:]
}
