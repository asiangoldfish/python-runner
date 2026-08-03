package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const VERSION = "1.1.3"

func main() {
	// Help pages
	if len(os.Args) < 2 {
		usage()
		return
	}

	option := os.Args[1]

	// Parse arguments
	switch option {
	case "init":
		if !initialise() {
			os.Exit(1)
		}
	case "-h", "--help", "help":
		usage()
	case "run":
		if !run() {
			os.Exit(1)
		}
	case "-v", "--version", "version":
		fmt.Println("pyrun " + VERSION)
	default:
		fmt.Fprintln(os.Stderr, "Option '"+option+"' is not recognised.")
	}
}

func usage() {
	fmt.Println(`Usage: pyrun [OPTION]
    
Python Runner is a wrapper for typical Python workflows. Under the hood, it uses
pyenv (Linux/macOS) or Python Install Manager (Windows) to manage Python
versions, and the built-in 'venv' module to manage Python environments.

List of available options:
    init       			 initialise a new or existing project
    help       			 this page
    run [SCRIPT_NAME]    execute a Python script.
    version              show the version number
    
Option 'init' simply creates a new virtual environment in $PWD. It uses pyenv
or the Python Install Manager to find the correct Python version based on
'.python-version'. If 'requirements.txt is available, then dependencies are
installed with it.

New dependencies can be installed using pip as per usual.`)
}

func initialise() bool {
	venvDir := ".venv"
	var version string

	fmt.Println("Initialising virtual environment")

	// Abort if venv already exists
	_, err := os.Stat(venvDir)
	if err == nil {
		fmt.Println("Virtual environment '.venv/' already exists. Aborting")
		return true
	}

	// venv not found. Create it and install deps if necessary...

	// Get Python version. We have two options:
	//
	// 1. Use .python-version
	// 2. Prompt the user for the version.
	//
	// In the second option, we prompt to install the Python version if it is
	// not already installed, and create the .python-version file.
	fmt.Println("Getting local Python version from '.python-version'")
	version, err = getPythonVersion()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return false
	}

	pythonPath, err := installPythonVersion(version)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return false
	} else {
		fmt.Println("Selected Python interpreter: " + pythonPath)
	}

	err = createVenv(pythonPath, venvDir, "requirements.txt")
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return false
	}

	return true
}

func run() bool {
	var binDir string

	// Check script presence
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Option 'run' requires a script name argument.")
		return false
	}

	scriptName := os.Args[2]
	info, err := os.Stat(scriptName)
	if err == nil {
		if info.IsDir() {
			fmt.Fprintln(os.Stderr, scriptName+" is a directory. It must be a file.")
			return false
		}
	} else if errors.Is(err, os.ErrNotExist) {
		// File not found
		fmt.Fprintln(os.Stderr, "Script "+scriptName+" does not exist.")
		return false
	} else {
		// Other errors
		fmt.Fprintln(os.Stderr, "File "+scriptName+" cannot be accessed.")
		fmt.Fprintln(os.Stderr, err.Error())
		return false
	}

	// Collect the rest of the args to be forwarded to the Python script.
	args := append([]string{scriptName}, os.Args[3:]...)

	// Get the bin path
	_, err = os.Stat(".venv")
	if err == nil {
		// Venv was found. Get bin path
		_, err := os.Stat(".venv/bin")
		if err != nil {
			binDir = ".venv/Scripts"
		} else {
			binDir = ".venv/bin"
		}
	} else if errors.Is(err, os.ErrNotExist) {
		// File not found
		fmt.Fprintf(os.Stderr, "Virtual environment '.venv' was not found. Create one with 'pyrun init'.\n")
		return false
	} else {
		// Other errors
		fmt.Fprintf(os.Stderr, "Virtual environment '.venv' cannot be accessed.\n")
		fmt.Fprintln(os.Stderr, err.Error())
		return false
	}

	cmd := exec.Command(
		binDir+"/python",
		args...,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return false
	}

	return true
}

// getVenvPython gets the Python interpreter in the virtual environment
func getVenvPython() string {
	// TODO
	return ".venv/Scripts"
}

func getPythonVersion() (string, error) {
	info, err := os.Stat(".python-version")
	var version string
	if err == nil {
		// .python-version exists. It could be a directory. Check against this.
		if info.IsDir() {
			// This is a directory. Abort.
			return "", errors.New("Tried to get Python version from file '.python-version', but it is a directory.")
		}

		// File was found.
		data, err := os.ReadFile(".python-version")
		if err != nil {
			return "", errors.New("File '.python-version' was found, but it failed to be read.")
		}

		// TODO validate version
		version = string(data)
		fmt.Println("Found Python version: " + version)

	} else if errors.Is(err, os.ErrNotExist) {
		// .python-version does *NOT* exist.
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter required Python version: ")
		version, err = reader.ReadString('\n')
		if err != nil {
			return "", errors.New("Failed to read Python version input")
		}

		version = strings.TrimSuffix(version, "\n")

		// Write the version to .python-version
		err := os.WriteFile(".python-version", []byte(version), 0644)
		if err != nil {
			return "", err
		} else {
			fmt.Println("Version " + version + " written to .python-version")
		}
	} else {
		// Other errors, e.g. file permission errors.
		return "", err
	}

	// TODO verify that the version is correctly formatted.
	return version, nil
}

func createVenv(pythonPath string, venvDir string, requirementsFile string) error {
	fmt.Print("Project Python interpreter not found. ")
	fmt.Println("Creating virtual environment...")

	cmd := exec.Command(pythonPath, "-m", "venv", ".venv")
	if err := cmd.Run(); err != nil {
		// Python not installed?
		return err
	}

	// On Windows, the executables are stored in .venv/Scripts, while it is
	// .venv/bin elsewhere.
	var binDir string
	_, err := os.Stat(venvDir + "/bin")
	if err != nil {
		binDir = venvDir + "/Scripts"
	} else {
		binDir = venvDir + "/bin"
	}

	// Install deps with requirements.txt
	_, err = os.Stat(requirementsFile)
	cmd = exec.Command(binDir+"/pip", "install", "-r", requirementsFile)
	if err == nil {
		// requirements.txt was found. Install deps.
		fmt.Println("Installing dependencies...")
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

// installPythonVersion gets the appropriate Python version based on the version.
//
// The following methods are used in order (first applicable method is used):
// 1. Python+version.
// 2. Use pyenv.
// 3. Use Python Install Manager.
// 4. Return false.
//
// If pyenv or Python Install Manager is found, use it to install the desired
// version.
func installPythonVersion(version string) (string, error) {
	// Is Python already installed?
	if path, err := exec.LookPath("python" + version); err == nil {
		return path, nil
	} else if path, err = exec.LookPath("pyenv"); err == nil {
		// pyenv was found. Use it to get the version
		cmd := exec.Command("pyenv exec python" + version)

		// Check if the Python version is installed
		if err := cmd.Run(); err == nil {
			// Python version was found! Return the path
			var buffer bytes.Buffer
			cmd := exec.Command("pyenv", "which", "python"+version)
			cmd.Stdout = &buffer

			return buffer.String(), cmd.Run()
		} else {
			// Install the Python version
			inputOption := ""
			reader := bufio.NewReader(os.Stdin)

			// Prompt for confirmation
			for inputOption != "y" && inputOption != "n" {
				fmt.Print("Would you like to install Python" + version + " with pyenv? [y/n] ")

				// TODO handle error
				inputOption, err = reader.ReadString('\n')
				inputOption = strings.TrimRight(inputOption, "\n")
				inputOption = strings.TrimSpace(inputOption)
				inputOption = strings.ToLower(inputOption)

				if inputOption != "y" && inputOption != "n" {
					fmt.Println("Must be either 'y' or 'n'")
				}
			}

			if inputOption == "n" {
				return "", errors.New("Python" + version + " was rejected. Aborting")
			} else {
				cmd := exec.Command("pyenv", "install", version)
				fmt.Println("Installing Python" + version + "...")
				if err := cmd.Run(); err != nil {
					// Failed to install Python interpreter
					return "", err
				} else {
					// Success! Return the path
					var buffer bytes.Buffer
					cmd := exec.Command("pyenv", "which", "python"+version)
					cmd.Stdout = &buffer

					return buffer.String(), cmd.Run()
				}
			}
		}
	} else if path, err = exec.LookPath("py"); err == nil {
		// Python Install Manager was found.
	}

	// Failed to get system Python path
	return "", errors.New("Failed to get system Python path for version " + version)
}
