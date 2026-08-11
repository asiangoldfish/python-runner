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

const VERSION = "1.2.3"

var VERBOSE = false

func main() {
	// Help pages
	if len(os.Args) < 2 {
		usage()
		return
	}

	// We mark the starting index for options, as we parse the flags first.
	startIndexForOptions := -1

	// Find flags. These alter the program's behaviour.
	for i, flag := range os.Args[1:] {
		if flag[0] != '-' {
			break
		}

		switch flag {
		case "-v":
			VERBOSE = true
		default:
			fmt.Fprintln(os.Stderr, "Flag "+flag+" is not recognised.")
			os.Exit(1)
		}

		startIndexForOptions = i
	}

	rest := os.Args[startIndexForOptions+2:]

	// No option was provided.
	if startIndexForOptions <= 1 && len(rest) == 0 {
		fmt.Fprintln(os.Stderr, "Flags were provided, but options are missing.")
		return
	}

	// Parse arguments
	switch rest[0] {
	case "init":
		if !initialise() {
			os.Exit(1)
		}
	case "-h", "--help", "help":
		usage()
	case "install":
		if !installPackage(rest[1:]) {
			os.Exit(1)
		}
	case "run":
		if !run(rest[1:]) {
			os.Exit(1)
		}
	case "version":
		fmt.Println("pyrun " + VERSION)
	default:
		fmt.Fprintln(os.Stderr, "Option '"+rest[0]+"' is not recognised.")
	}
}

func usage() {
	fmt.Println(`Usage: pyrun [FLAG...] OPTION

Python Runner is a wrapper for typical Python workflows. Under the hood, it uses
pyenv (Linux/macOS) or Python Install Manager (Windows) to manage Python
versions, and the built-in 'venv' module to manage Python environments.

An option is a special argument that performs a set of instructions. Pyrun
mainly is used together with such an option. A flag is an optional argument that
changes some of Pyrun's behaviour, if applicable. Flags should be provided
before the options to prevent argument conflicts.

List of flags:
    -v                   print verbose outputs

List of available options:
    init       			 initialise a new or existing project
    help       			 this page
    install              install a package using pip
    run [SCRIPT_NAME]    execute a Python script
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

	err = createVenv(pythonPath, "requirements.txt")
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return false
	} else {
		fmt.Println("Virtual environment '.venv' successfully created!")
	}

	return true
}

func run(args []string) bool {
	// Check script presence
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Option 'run' requires a script name argument.")
		return false
	}

	scriptName := args[0]
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

	// Get the bin path
	binDir, err := getBinDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
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

// installPackage installs a Python package in the virtual environment and
// writes new dependencies to requirements.txt.
func installPackage(args []string) bool {
	// Check script presence
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Option 'install' requires at least one package to install.")
		return false
	}

	binDir, err := getBinDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return false
	}

	cmdArgs := append([]string{"install"}, args...)

	pipCommand, err := findPipCommand(binDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return false
	}

	cmd := exec.Command(
	    pipCommand,
		cmdArgs...,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return false
	}

	fmt.Println("Saving dependencies to requirements.txt...")
	cmd = exec.Command(pipCommand, "freeze")

	outfile, err := os.Create("./requirements.txt")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create requirements.txt")
		fmt.Fprintln(os.Stderr, err.Error())
		return false
	}
	defer outfile.Close()
	cmd.Stdout = outfile

	if err := cmd.Run(); err != nil {
		fmt.Fprint(os.Stderr, "Failed to save dependencies to requirements.txt.")
		fmt.Fprint(os.Stderr, err.Error())
		return false
	}

	fmt.Println("Packages successfully installed!")

	return true
}

// getBinDir gets the virtual environment directory where the Python related
// binary files reside.
func getBinDir() (string, error) {
	var binDir string

	// Get the bin path
	_, err := os.Stat(".venv")
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
		return "", errors.New("Virtual environment '.venv' was not found. Create one with 'pyrun init'.")
	} else {
		// Other errors
		return "", errors.New("Virtual environment '.venv' cannot be accessed.\n" + err.Error())
	}

	return binDir, nil
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

		version = strings.TrimSpace(version)

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

// createVenv creates a new virtual environment for the Python installation, and
// installs dependencies found in 'requirements.txt'.
//
// Installing the virtual environment comes in two steps:
// 1. Copying the Python interpreter and creating an environment for it
// 2. Unpacking pip that is included in the Python installation.
//
// Although the second step is not required because 'venv' does this for us, we
// can log the steps to the user so they are prepared for a small waiting time.
func createVenv(pythonPath string, requirementsFile string) error {
	fmt.Print("Project Python interpreter not found. ")
	fmt.Println("Creating virtual environment...")

	cmd := exec.Command(pythonPath, "-m", "venv", ".venv", "--without-pip")
	if err := cmd.Run(); err != nil {
		// Python not installed?
		return err
	}

	// On Windows, the executables are stored in .venv/Scripts, while it is
	// .venv/bin elsewhere.
	binDir, err := getBinDir()
	if err != nil {
		return err
	}

	// Install pip
	fmt.Println("Installing pip. This can take a few seconds...")
	cmd = exec.Command(binDir+"/python", "-m", "ensurepip")
	if VERBOSE {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	// Some Python installations bundle pip as the `pip` command, and others as
	// `pip3`.
	pipCommand, err := findPipCommand(binDir)
	if VERBOSE {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err != nil {
		return err
	}

	// Install deps with requirements.txt
	_, err = os.Stat(requirementsFile)
	cmd = exec.Command(pipCommand, "install", "-r", requirementsFile)
	if err == nil {
		// requirements.txt was found. Install deps.
		fmt.Println("Installing dependencies...")
		if VERBOSE {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}
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

// findPipCommand tries to find the command name for the Pip package manager.
//
// Arguments
// ---------
// binDir (string): directory where the pip command is expected to be found in.
//
// Return
// ------
// string: Filepath to the pip executable file.
// error: Possible errors that may occur.
func findPipCommand(binDir string) (string, error) {
	possibleCommands := []string{"pip", "pip3"}
	pipCommand := ""

	for _, v := range possibleCommands {
		cmd := exec.Command(binDir + "/" + v)
		if err := cmd.Run(); err == nil {
			pipCommand = binDir + "/" + v
		}
	}

	// Prevent using global pip command
	_, err := os.Stat(pipCommand)
	if err != nil {
		return "", errors.New("Pip command '" + pipCommand + "' was recognised but not found as a file.")
	}

	if VERBOSE {
		fmt.Println("Pip command found at: " + pipCommand)
	}

	if pipCommand == "" {
		return "", errors.New("The filepath to command 'pip' was not found.")
	} else {
		return pipCommand, nil
	}
}
