package platform

import (
	"context"
	"encoding/binary"
	"os/signal"

	syscall "golang.org/x/sys/windows"
)

// Returns a context which is cancelled upon receiving a SIGINT or SIGTERM signal.
// Uses Windows syscalls for compatibility. Probably it could work using normal syscall package
// but better keep it clear.
func SignalNotifyContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
}

// Tries to check if program is running with privileges on Windows. Cursed, don't use it..
// I haven't tested it even...
// Based on https://stackoverflow.com/questions/8046097/how-to-check-if-a-process-has-the-administrative-rights
func IsElevated() (bool, error) {
	ret := false
	var hToken syscall.Token
	err := syscall.OpenProcessToken(syscall.CurrentProcess(), syscall.TOKEN_QUERY, &hToken)
	if err != nil {
		return false, err
	}
	info := make([]byte, 4)
	var retLen uint32
	err = syscall.GetTokenInformation(hToken, syscall.TokenElevation, &info[0], 4, &retLen)
	if err != nil {
		return false, err
	}
	r := binary.LittleEndian.Uint32(info)
	if r != 0 {
		ret = true
	}

	return ret, nil
}
