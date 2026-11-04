//go:build darwin

package audio

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework AVFoundation -framework Foundation -framework CoreFoundation -Wl,-sectcreate,__TEXT,__info_plist,${SRCDIR}/Info.plist
#import <AVFoundation/AVFoundation.h>
#import <dispatch/dispatch.h>
#include <stdint.h>

// Authorization values match AVAuthorizationStatus.
static int hlMicAuth(void) {
	if (@available(macOS 10.14, *)) {
		return (int)[AVCaptureDevice authorizationStatusForMediaType:AVMediaTypeAudio];
	}
	return 3;
}

// The dialog is delivered on the main queue. The caller must not be the
// main thread, or this wait deadlocks. Hearline starts capture from a
// session goroutine, which is not the UI thread.
static int hlMicRequest(void) {
	if (@available(macOS 10.14, *)) {
		__block int granted = 0;
		dispatch_semaphore_t sem = dispatch_semaphore_create(0);
		dispatch_async(dispatch_get_main_queue(), ^{
			[AVCaptureDevice requestAccessForMediaType:AVMediaTypeAudio completionHandler:^(BOOL ok) {
				granted = ok ? 1 : 0;
				dispatch_semaphore_signal(sem);
			}];
		});
		dispatch_semaphore_wait(sem, dispatch_time(DISPATCH_TIME_NOW, (int64_t)60 * NSEC_PER_SEC));
		return granted;
	}
	return 1;
}
*/
import "C"
import "context"

const (
	micAuthNotDetermined = 0
	micAuthRestricted    = 1
	micAuthDenied        = 2
	micAuthAuthorized    = 3
)

func micAuth() int {
	return int(C.hlMicAuth())
}

// ensureMicPermission asks TCC before Core Audio is opened. A denied or
// restricted grant is returned as microphoneDenied so the window shows
// 「没有麦克风/系统声权限」 instead of a file error. Not-determined asks
// once; the embedded Info.plist supplies NSMicrophoneUsageDescription
// when this process is the TCC responsible process. A terminal launch
// may still attribute the prompt to Terminal.
func ensureMicPermission(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	switch micAuth() {
	case micAuthAuthorized:
		return nil
	case micAuthDenied, micAuthRestricted:
		return microphoneDenied()
	}
	// Not determined. Ask once. A false return is only fatal when TCC has
	// actually moved to denied: an unsigned binary launched from Terminal
	// can get a false negative here while Core Audio itself is allowed.
	granted := make(chan int, 1)
	go func() {
		granted <- int(C.hlMicRequest())
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case ok := <-granted:
		if ok == 0 && micPermissionBlocked() {
			return microphoneDenied()
		}
		return nil
	}
}

func micPermissionBlocked() bool {
	switch micAuth() {
	case micAuthDenied, micAuthRestricted:
		return true
	default:
		return false
	}
}
