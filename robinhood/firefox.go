package robinhood

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/golang/snappy"
	"github.com/ncruces/go-sqlite3"
)

const (
	// path for browser executables
	FirefoxExecutableWindows = `c:\program files\mozilla firefox\firefox.exe`
	// requires $home path to be added
	FirefoxExecutableMac = "/applications/firefox.app/contents/macos/firefox"
	// requires $home path to be added
	FirefoxExecutableLinux = "/usr/bin/firefox"

	// Requires $HOME path to be added
	FirefoxPathToProfileMac = "Library/Application Support/Firefox/Profiles/"
	// Requires %USERPROFILE% path to be added
	FirefoxPathToProfileWindows = `AppData\Roaming\Mozilla\Firefox\Profiles\`
	// Requires $HOME path to be added
	FirefoxPathToProfilelinux = ".mozilla/firefox/"

	// This is the folder name that holds the db on all platforms
	FirefoxDBPath = "https+++robinhood.com/ls/data.sqlite"
	// Windows version just uses backslashes
	FirefoxDBPathWindows = `https+++robinhood.com\ls\data.sqlite`
)

type Firefox struct {
	pathToDB         string
	pathToExecutable string
}

func (f *Firefox) PathToExecutable() string {
	return f.pathToExecutable
}

func (f *Firefox) PathToDB() string {
	return f.pathToDB
}

func (f *Firefox) String() string {
	return fmt.Sprintf("Firefox{PathToDB:%q, PathToExecutable:%q", f.pathToDB, f.pathToExecutable)
}

// Should only be used by newFirefox
func getFirefoxExecutablePath(homeDir string) string {
	var executePath string
	switch runtime.GOOS {
	case "darwin":
		executePath = filepath.Join(homeDir, FirefoxExecutableMac)
	case "linux":
		executePath = filepath.Join(homeDir, FirefoxExecutableLinux)
	case "windows":
		executePath = FirefoxExecutableWindows
	default:
		panic(fmt.Errorf("platform is not supported %s", runtime.GOOS))
	}
	return executePath
}

// Parse Firefox profile for the last accessed profile, this is a somewhat hacky
// way to determine the correct profile to use
func findFirefoxProfile(path string) (string, error) {
	profiles, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}
	var fileName string
	var lastModifiedTime time.Time
	for _, v := range profiles {
		fileInfo, err := v.Info()
		if err != nil {
			continue
		}
		fileTime := fileInfo.ModTime()
		if fileName == "" {
			fileName = v.Name()
			lastModifiedTime = fileTime
		} else if fileTime.After(lastModifiedTime) {
			lastModifiedTime = fileInfo.ModTime()
			fileName = v.Name()
		}
	}
	return filepath.Join(path, fileName, "storage", "default", FirefoxDBPath), nil
}

// Should only be used by newFirefox
//
// Retrieves the Robinhood db path from local storage
func getFirefoxDBPath(homeDir string) string {
	switch runtime.GOOS {
	case "darwin":
		DBPath := filepath.Join(homeDir, FirefoxPathToProfileMac)
		DBPath, err := findFirefoxProfile(DBPath)
		if err != nil {
			panic(fmt.Errorf("failed to read %s, %v", DBPath, err))
		}
		return DBPath
	case "linux":
		// TODO: implement this sometime
		DBPath := ""
		return DBPath
	case "windows":
		DBPath := filepath.Join(homeDir, FirefoxPathToProfileWindows)
		DBPath, err := findFirefoxProfile(DBPath)
		if err != nil {
			panic(fmt.Errorf("failed to read %s, %v", DBPath, err))
		}
		return DBPath
	default:
		panic(fmt.Errorf("platform is not supported %s", runtime.GOOS))
	}
}

// ExtractToken from local firefox sqlite database from web:auth_state,
// then decode blob with snappy and return access_token string
func (f *Firefox) ExtractToken() (string, error) {
	conn, err := sqlite3.OpenFlags(f.pathToDB, sqlite3.OPEN_READONLY)
	if err != nil {
		return "", fmt.Errorf("failed to connect to %v, %v", conn, err)
	}
	defer conn.Close() //nolint:errcheck
	stmt, _, err := conn.Prepare(`SELECT value FROM data WHERE key = 'web:auth_state'`)
	if err != nil {
		return "", fmt.Errorf("error %v", err)
	}
	defer stmt.Close() //nolint:errcheck
	if stmt.Step() {
		blob := stmt.ColumnBlob(0, nil)
		jtw, err := snappy.Decode(nil, blob)
		if err != nil {
			return "", fmt.Errorf("error %v", err)
		}
		var tempJWTStruct struct {
			AccessToken string `json:"access_token"`
		}
		err = json.Unmarshal(jtw, &tempJWTStruct)
		if err != nil {
			return "", fmt.Errorf("failed to unmarhsal jwt, error %v", err)
		}
		return tempJWTStruct.AccessToken, nil
	}
	return "", fmt.Errorf("token was not found from %s", f.pathToDB)
}

func NewFirefox() *Firefox {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("unable to find home dir: %v", err))
	}
	return &Firefox{
		pathToDB:         getFirefoxDBPath(homeDir),
		pathToExecutable: getFirefoxExecutablePath(homeDir),
	}
}
