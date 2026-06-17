package robinhood

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/golang/snappy"
	"github.com/google/uuid"
	"github.com/ncruces/go-sqlite3"
)

const (
	// Path for Browser executables
	FirefoxExecutableWindows = `C:\Program Files\Mozilla Firefox\firefox.exe`
	// Requires $HOME path to be added
	FirefoxExecutableMac = "/applications/firefox.app/contents/macos/firefox"
	// Requires $HOME path to be added
	FirefoxExecutableLinux = "/usr/bin/firefox"
	// ABS path to the chrome executable
	ChromeExecutableWindows = `C:\Program Files\Google\Chrome\Application\chrome.exe`
	// Requires $HOME path to be added
	ChromeExecutableMac = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	// ABS path to the chrome executable
	ChromeExecutableLinux = "/usr/bin/google-chrome"
)

const (
	// Path to local Robinhood storage, Assuming 'Default' profile is used
	//
	// Requires $HOME path to be added
	ChromePathToProfileDirMac = "Library/Application Support/Google/Chrome/Default/IndexedDB/storage/default/"
	// Requires $HOME path to be added
	ChromePathToProfileDirLinux = ".config/google-chrome/Default/IndexedDB/storage/default/"
	// Requires %USERPROFILE% path to be added
	ChromePathToProfileDirWindows = `AppData\Local\Google\Chrome\User Data\Default\IndexedDB`

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
	// Chrome uses leveldb implementation instead of sqlite, works for all supported platforms
	ChromeDBDir = "https_robinhood.com_0.indexeddb.leveldb"
)

type Browser interface {
	String() string
	OpenAndClose(waitTime float32, headless bool)
	ExtractToken(pathToDB string) string
	GetPathToDB() string
	GetPathToExecutable()
}

type Firefox struct {
	PathToDB         string
	PathToExecutable string
}

func (f *Firefox) String() string {
	return fmt.Sprintf("Firefox{PathToDB:%q, PathToExecutable:%q", f.PathToDB, f.PathToExecutable)
}

type Chrome struct {
	PathToDB         string
	PathToExecutable string
}

func (c *Chrome) String() string {
	return fmt.Sprintf("Chrome{PathToDB:%q, PathToExecutable:%q", c.PathToDB, c.PathToExecutable)
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
		DBPath := ""
		return DBPath
	case "windows":
		DBPath := ""
		return DBPath
	default:
		panic(fmt.Errorf("platform is not supported %s", runtime.GOOS))
	}
}

func NewChrome() *Chrome {
	return &Chrome{
		PathToDB:         "",
		PathToExecutable: "",
	}
}

func NewFirefox() *Firefox {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("unable to find home dir: %v", err))
	}
	return &Firefox{
		PathToDB:         getFirefoxDBPath(homeDir),
		PathToExecutable: getFirefoxExecutablePath(homeDir),
	}
}

// ExtractToken from local firefox sqlite database from web:auth_state,
// then decode blob with snappy and return access_token string
func (f *Firefox) ExtractToken() (string, error) {
	conn, err := sqlite3.OpenFlags(f.PathToDB, sqlite3.OPEN_READONLY)
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
	return "", fmt.Errorf("token was not found from %s", f.PathToDB)
}

func ExtractTokenChrome() {
	chrome := NewChrome()
	fmt.Print(chrome, "\n")
}

// loadToken generate a new uuid from local storage
func loadToken() uuid.UUID {
	return uuid.New()
}

// def _decode_jwt(payload: str) -> dict[str, Any]:
//     # idk how this would happen but better safe than sorry
//     # and make easier debugging in the event this happens
//     if not isinstance(payload, str):
//         raise ValueError(
//             f"payload {payload} was of type {type(payload)} not str"
//         )
//     payload_b64 = payload.split(".")[1]
//     payload_b64 += "=" * (-len(payload_b64) % 4)
//     decoded_payload: dict[str, Any] = json.loads(
//         base64.urlsafe_b64decode(payload_b64).decode()
//     )
//     return decoded_payload
