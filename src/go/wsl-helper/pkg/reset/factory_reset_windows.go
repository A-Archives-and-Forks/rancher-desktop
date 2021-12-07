/*
Copyright © 2021 SUSE LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package reset implements factory reset for Windows.
package reset

import (
	"errors"
	"fmt"
	"os"
	"path"
	"unsafe"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
)

const (
	appName = "rancher-desktop"
)

// Factory reset deletes any Rancher Desktop user data.
func FactoryReset() error {
	appData, err := getKnownFolder(windows.FOLDERID_RoamingAppData)
	if err != nil {
		return fmt.Errorf("could not get AppData folder: %w", err)
	}
	localAppData, err := getKnownFolder(windows.FOLDERID_LocalAppData)
	if err != nil {
		return fmt.Errorf("could not get LocalAppData folder: %w", err)
	}
	for _, dir := range []string{
		// Ordered from least important to most, so that if delete fails we
		// still keep some of the useful data.
		path.Join(localAppData, fmt.Sprintf("%s-updater", appName)),
		path.Join(localAppData, appName),
		path.Join(appData, appName),
	} {
		logrus.WithField("path", dir).Trace("Removing directory")
		if err := os.RemoveAll(dir); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("could not remove %s: %w", dir, err)
			}
		}
	}
	return nil
}

var (
	ole32Dll   = windows.MustLoadDLL("Ole32.dll")
	shell32Dll = windows.MustLoadDLL("Shell32.dll")
)

// getKnownFolder gets a Windows known folder.  See https://git.io/JMpgD
func getKnownFolder(folder *windows.KNOWNFOLDERID) (string, error) {
	SHGetKnownFolderPath, err := shell32Dll.FindProc("SHGetKnownFolderPath")
	if err != nil {
		return "", fmt.Errorf("could not find SHGetKnownFolderPath: %w", err)
	}
	CoTaskMemFree, err := ole32Dll.FindProc("CoTaskMemFree")
	if err != nil {
		return "", fmt.Errorf("could not find CoTaskMemFree: %w", err)
	}
	var result uintptr
	hr, _, _ := SHGetKnownFolderPath.Call(
		uintptr(unsafe.Pointer(folder)),
		0,
		uintptr(unsafe.Pointer(nil)),
		uintptr(unsafe.Pointer(&result)),
	)
	// SHGetKnownFolderPath documentation says we _must_ free the result with
	// CoTaskMemFree, even if the call failed.
	defer CoTaskMemFree.Call(result)
	if hr != 0 {
		return "", windows.Errno(hr)
	}

	// result at this point containers the path, as a PWSTR
	// Note that `go vet` has a false positive here on "misuse of Pointer".
	return windows.UTF16PtrToString((*uint16)(unsafe.Pointer(result))), nil
}
