package util

import (
	"runtime"
	"runtime/debug"
)

// FreeMemory runs FreeOSMemory() only

// FreeMemoryGC runs FreeOSMemory() and GC()
func FreeMemoryGC() {
	runtime.GC()
	debug.FreeOSMemory()
}
