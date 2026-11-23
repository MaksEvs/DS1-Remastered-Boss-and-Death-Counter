//go:build windows

package main

import (
	"errors"
	"fmt"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

const (
	processVMRead           = 0x0010
	processQueryInformation = 0x0400
	th32csSnapModule        = 0x00000008
	th32csSnapModule32      = 0x00000010
)

var (
	kernel32               = syscall.NewLazyDLL("kernel32.dll")
	openProcess            = kernel32.NewProc("OpenProcess")
	closeHandle            = kernel32.NewProc("CloseHandle")
	readProcessMemory      = kernel32.NewProc("ReadProcessMemory")
	createToolhelpSnapshot = kernel32.NewProc("CreateToolhelp32Snapshot")
	module32First          = kernel32.NewProc("Module32FirstW")
	process32First         = kernel32.NewProc("Process32FirstW")
	process32Next          = kernel32.NewProc("Process32NextW")
)

type moduleEntry32 struct {
	Size         uint32
	ModuleID     uint32
	ProcessID    uint32
	GlobalUsage  uint32
	ProcessUsage uint32
	BaseAddress  uintptr
	BaseSize     uint32
	Module       uintptr
	Name         [256]uint16
	Path         [260]uint16
}

type processEntry32 struct {
	Size            uint32
	Usage           uint32
	ProcessID       uint32
	DefaultHeapID   uintptr
	ModuleID        uint32
	Threads         uint32
	ParentProcessID uint32
	Priority        int32
	Flags           uint32
	ExecutableName  [260]uint16
}

type processMemoryReader struct {
	handle     uintptr
	moduleBase uintptr
	moduleSize uint32
}

func findProcessIDByName(processName string) (uint32, error) {
	snapshot, _, callErr := createToolhelpSnapshot.Call(0x00000002, 0)
	if snapshot == ^uintptr(0) {
		return 0, fmt.Errorf("CreateToolhelp32Snapshot failed: %w", callErr)
	}
	defer closeHandle.Call(snapshot)

	entry := processEntry32{Size: uint32(unsafe.Sizeof(processEntry32{}))}
	result, _, callErr := process32First.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
	if result == 0 {
		return 0, fmt.Errorf("Process32FirstW failed: %w", callErr)
	}

	for {
		name := utf16.Decode(entry.ExecutableName[:])
		for index, character := range name {
			if character == 0 {
				name = name[:index]
				break
			}
		}
		if equalWindowsProcessName(string(name), processName) {
			return entry.ProcessID, nil
		}

		result, _, _ = process32Next.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
		if result == 0 {
			break
		}
	}

	return 0, fmt.Errorf("%s was not found", processName)
}

func equalWindowsProcessName(actualName string, expectedName string) bool {
	return strings.EqualFold(actualName, expectedName)
}

func connectToGame(processID uint32) (*processMemoryReader, error) {
	handle, _, callErr := openProcess.Call(processVMRead|processQueryInformation, 0, uintptr(processID))
	if handle == 0 {
		return nil, fmt.Errorf("OpenProcess failed: %w", callErr)
	}

	snapshot, _, callErr := createToolhelpSnapshot.Call(th32csSnapModule|th32csSnapModule32, uintptr(processID))
	if snapshot == ^uintptr(0) {
		closeHandle.Call(handle)
		return nil, fmt.Errorf("CreateToolhelp32Snapshot failed: %w", callErr)
	}
	defer closeHandle.Call(snapshot)

	module := moduleEntry32{Size: uint32(unsafe.Sizeof(moduleEntry32{}))}
	result, _, callErr := module32First.Call(snapshot, uintptr(unsafe.Pointer(&module)))
	if result == 0 {
		closeHandle.Call(handle)
		return nil, fmt.Errorf("Module32FirstW failed: %w", callErr)
	}

	return &processMemoryReader{handle: handle, moduleBase: module.BaseAddress, moduleSize: module.BaseSize}, nil
}

func (reader *processMemoryReader) close() {
	if reader != nil && reader.handle != 0 {
		closeHandle.Call(reader.handle)
		reader.handle = 0
	}
}

func (reader *processMemoryReader) read(address uintptr, size uint32) ([]byte, error) {
	buffer := make([]byte, size)
	var bytesRead uintptr
	result, _, callErr := readProcessMemory.Call(reader.handle, address, uintptr(unsafe.Pointer(&buffer[0])), uintptr(size), uintptr(unsafe.Pointer(&bytesRead)))
	if result == 0 || bytesRead != uintptr(size) {
		return nil, fmt.Errorf("ReadProcessMemory at 0x%x failed: %w", address, callErr)
	}
	return buffer, nil
}

func (reader *processMemoryReader) scan(pattern []byte, wildcards map[int]bool) (uintptr, error) {
	scanSize := reader.moduleSize
	if scanSize > 0x1000000 {
		scanSize = 0x1000000
	}
	const chunkSize = 0x10000
	for offset := uint32(0); offset < scanSize; offset += chunkSize {
		remaining := scanSize - offset
		readSize := uint32(chunkSize)
		if remaining < readSize {
			readSize = remaining
		}
		data, err := reader.read(reader.moduleBase+uintptr(offset), readSize)
		if err != nil {
			continue
		}
		for index := 0; index+len(pattern) <= len(data); index++ {
			matches := true
			for patternIndex, expected := range pattern {
				if wildcards[patternIndex] || data[index+patternIndex] == expected {
					continue
				}
				matches = false
				break
			}
			if matches {
				return reader.moduleBase + uintptr(offset) + uintptr(index), nil
			}
		}
	}
	return 0, errors.New("event flag signature was not found")
}

func moduleName(entry moduleEntry32) string {
	name := utf16.Decode(entry.Name[:])
	for index, character := range name {
		if character == 0 {
			return string(name[:index])
		}
	}
	return string(name)
}
