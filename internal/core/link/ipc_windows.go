//go:build windows

package link

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procCreateEvent         = kernel32.NewProc("CreateEventW")
	procCloseHandle         = kernel32.NewProc("CloseHandle")
	procSetEvent            = kernel32.NewProc("SetEvent")
	procWaitForSingleObject = kernel32.NewProc("WaitForSingleObject")
	procCreateFileMapping   = kernel32.NewProc("CreateFileMappingW")
	procMapViewOfFile       = kernel32.NewProc("MapViewOfFile")
	procUnmapViewOfFile     = kernel32.NewProc("UnmapViewOfFile")
)

const (
	FILE_MAP_ALL_ACCESS = 0xF001F
	PAGE_READWRITE      = 0x04
)

// mapSharedMemory mapeia um arquivo em memória compartilhada
func mapSharedMemory(file *os.File, size int) ([]byte, error) {
	handle, _, err := procCreateFileMapping.Call(
		uintptr(file.Fd()),
		0,
		uintptr(PAGE_READWRITE),
		0,
		uintptr(size),
		0,
	)
	if handle == 0 {
		return nil, err
	}
	defer procCloseHandle.Call(handle)

	addr, _, err := procMapViewOfFile.Call(
		handle,
		uintptr(FILE_MAP_ALL_ACCESS),
		0,
		0,
		uintptr(size),
	)
	if addr == 0 {
		return nil, err
	}

	data := make([]byte, size)
	copy(data, (*[1 << 30]byte)(unsafe.Pointer(addr))[:size:size])

	return data, nil
}

// unmapSharedMemory desmapeia a memória compartilhada
func unmapSharedMemory(data []byte) error {
	_, _, err := procUnmapViewOfFile.Call(uintptr(unsafe.Pointer(&data[0])))
	return err
}

// createEvent cria um evento do Windows
func createEvent(name string) (uintptr, error) {
	namePtr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return 0, err
	}

	handle, _, err := procCreateEvent.Call(
		0,                                // Atributos de segurança padrão
		1,                                // Manual reset
		0,                                // Estado inicial não sinalizado
		uintptr(unsafe.Pointer(namePtr)), // Nome do evento
	)

	if handle == 0 {
		return 0, err
	}

	return handle, nil
}

// closeEvent fecha um handle de evento
func closeEvent(handle uintptr) {
	procCloseHandle.Call(handle)
}

// setEvent sinaliza um evento
func setEvent(handle uintptr) {
	procSetEvent.Call(handle)
}

// waitForEvent espera por um evento
func waitForEvent(handle uintptr) {
	procWaitForSingleObject.Call(
		handle,
		syscall.INFINITE,
	)
}
