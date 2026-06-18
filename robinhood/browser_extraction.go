package robinhood

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/golang/snappy"
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
	// TODO: implement this later
	// OpenAndClose(waitTime float64, headless bool)
	ExtractToken() (string, error)
	PathToDB() string
	PathToExecutable() string
}

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

type Chrome struct {
	pathToDB         string
	pathToExecutable string
}

func (c *Chrome) String() string {
	return fmt.Sprintf("Chrome{PathToDB:%q, PathToExecutable:%q", c.pathToDB, c.pathToExecutable)
}

func (c *Chrome) PathToExecutable() string {
	return c.pathToExecutable
}

func (c *Chrome) PathToDB() string {
	return c.pathToDB
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
		pathToDB:         "",
		pathToExecutable: "",
	}
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

func (c *Chrome) ExtractToken() (string, error) {
	chrome := NewChrome()
	fmt.Print(chrome, "\n")
	return "", nil
}

// Decode the token extracted from local storage
func decodeJWT(encodedToken string) (*[]byte, error) {
	token := strings.Split(encodedToken, ".")[1]
	padding := len(token) % 4
	token += strings.Repeat("=", 4-padding)
	decodedToken, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}
	return &decodedToken, nil
}

// ReturnJWTExpiration returns the the JWT expiration
func ReturnJWTExpiration(encodedToken string) (*time.Time, error) {
	bytesJWT, err := decodeJWT(encodedToken)
	if err != nil {
		return nil, err
	}
	var exp struct {
		Exp int64 `json:"exp"`
	}
	err = json.Unmarshal(*bytesJWT, &exp)
	if err != nil {
		return nil, err
	}
	expDate := time.Unix(exp.Exp, 0)
	return &expDate, nil
}

// ValidateToken returns False on invalid token or on error
//
// e.g. expired, malformed-token, etc.
//
// Uses https://api.robinhood.com/accounts/ as the endpoint
func (rh *RobinhoodClient) ValidateToken(token string) (bool, error) {
	expiration, err := ReturnJWTExpiration(token)
	if err != nil {
		return false, err
	}
	if expiration.Compare(time.Now().UTC()) < 0 {
		return false, fmt.Errorf("token is expired")
	}
	const apiAcc = "/accounts/"
	request, err := rh.buildGetRequest(apiAcc, nil)
	if err != nil {
		return false, err
	}
	request.Header.Add("Authorization", "Bearer "+token)
	response, err := rh.doGetRequest(request)
	if err != nil {
		return false, err
	}
	if response.StatusCode > 300 {
		return false, fmt.Errorf("reponse returned %d", response.StatusCode)
	}
	return true, nil
}
