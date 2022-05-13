package platform

import (
	"context"
	"os/signal"
	"os/user"
	"syscall"
)

// Returns a context which is cancelled upon receiving a SIGINT or SIGTERM signal.
func SignalNotifyContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
}

// Tries to check if program is running with privileges on Linux (the only Unix I tested).
// The program can be tricked but the purpose is to alert the user in the future that it requires privileges to run...
func IsElevated() (bool, error) {
	u, err := user.Current()
	if err != nil {
		return false, err
	}

	// Does it by checking for the current process user Uid.
	// Other possibilities could be checking for u.Name == "root", u.Username == "root"...
	// Maybe a mix of && || ?
	return u.Uid == "0", nil
}
