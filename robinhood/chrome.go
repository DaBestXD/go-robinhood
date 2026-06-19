package robinhood

import "fmt"

const (
	// ABS path to the chrome executable
	ChromeExecutableWindows = `C:\Program Files\Google\Chrome\Application\chrome.exe`
	// Requires $HOME path to be added
	ChromeExecutableMac = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	// ABS path to the chrome executable
	ChromeExecutableLinux = "/usr/bin/google-chrome"
	// Path to local Robinhood storage, Assuming 'Default' profile is used
	//
	// Requires $HOME path to be added
	ChromePathToProfileDirMac = "Library/Application Support/Google/Chrome/Default/IndexedDB/storage/default/"
	// Requires $HOME path to be added
	ChromePathToProfileDirLinux = ".config/google-chrome/Default/IndexedDB/storage/default/"
	// Requires %USERPROFILE% path to be added
	ChromePathToProfileDirWindows = `AppData\Local\Google\Chrome\User Data\Default\IndexedDB`

	// Chrome uses leveldb implementation instead of sqlite, works for all supported platforms
	ChromeDBDir = "https_robinhood.com_0.indexeddb.leveldb"
)

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

func (c *Chrome) ExtractToken() (string, error) {
	chrome := NewChrome()
	fmt.Print(chrome, "\n")
	return "", nil
}

func NewChrome() *Chrome {
	return &Chrome{
		pathToDB:         "",
		pathToExecutable: "",
	}
}
