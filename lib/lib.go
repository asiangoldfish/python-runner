package lib

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Usage() {
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
    cmd [COMMAND]        execute an installed executable from the virtual
                         environment
    init                 initialise a new or existing project
    help                 this page
    install              install a package using pip
        --no-save        do not save installed packages to requirements.txt
    run [SCRIPT_NAME]    execute a Python script
    version              show the version number

Option 'init' simply creates a new virtual environment in $PWD. It uses pyenv
or the Python Install Manager to find the correct Python version based on
'.python-version'. If 'requirements.txt is available, then dependencies are
installed with it.

New dependencies can be installed using pip as per usual.

Option 'cmd' executes an installed executable file from the virtual environment.
Some packages provide a CLI interface, and this option can invoke them.`)
}

func Initialise(verbose bool) bool {
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
	version, err = GetPythonVersion(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return false
	}

	pythonPath, err := InstallPythonVersion(version)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return false
	} else {
		fmt.Println("Selected Python interpreter: " + pythonPath)
	}

	err = CreateVenv(".", pythonPath, "requirements.txt", verbose)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return false
	} else {
		fmt.Println("Virtual environment '.venv' successfully created!")
	}

	return true
}

func Run(args []string) error {
	// Check script presence
	if len(args) == 0 {
		return errors.New("Option 'run' requires a script name argument.")
	}

	scriptName := args[0]
	info, err := os.Stat(scriptName)
	if err == nil {
		if info.IsDir() {
			return errors.New(scriptName + " is a directory. It must be a file.")
		}
	} else if errors.Is(err, os.ErrNotExist) {
		// File not found
		return errors.New("Script " + scriptName + " does not exist.")
	} else {
		// Other errors
		return errors.New("File " + scriptName + " cannot be accessed. Error: " + err.Error())
	}

	// Get the bin path
	binDir, err := GetBinDir(".")
	if err != nil {
		return err
	}

	cmd := exec.Command(
		binDir+"/python",
		args...,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// installPackage installs a Python package in the virtual environment and
// writes new dependencies to requirements.txt.
//
// venvParentDir specifies in which directory the ".venv" directory is in.
func InstallPackage(packages []string, verbose bool, venvParentDir string) error {
	noSave := false

	// Remove --no-save
	noSaveIndex := -1
	for i, v := range packages {
		if v == "--no-save" {
			noSave = true
			noSaveIndex = i
			break
		}
	}

	// Check script presence
	if len(packages) == 0 {
		return errors.New("Option 'install' requires at least one package to install.")
	}

	if noSave {
		packages = append(packages[:noSaveIndex], packages[noSaveIndex+1:]...)
	}

	binDir, err := GetBinDir(venvParentDir)
	if err != nil {
		return err
	}

	cmdArgs := append([]string{"install"}, packages...)

	pipCommand, err := FindPipCommand(binDir, verbose)
	if err != nil {
		return err
	}

	cmd := exec.Command(
		pipCommand,
		cmdArgs...,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		return err
	}

	if !noSave {
		fmt.Println("Saving dependencies to requirements.txt...")
		cmd = exec.Command(pipCommand, "freeze")

		outfile, err := os.Create(venvParentDir + "/requirements.txt")
		if err != nil {
			return errors.New("Failed to create requirements.txt. Error: " + err.Error())
		}
		defer outfile.Close()
		cmd.Stdout = outfile

		if err := cmd.Run(); err != nil {
			return errors.New("Failed to save dependencies to requirements.txt. Error: " + err.Error())
		}
	}

	fmt.Println("Packages successfully installed!")

	return nil
}

// getBinDir gets the virtual environment directory where the Python related
// binary files reside. It assumes that the virtual environment directory is
// ".venv".
//
// parentDir specifies what parent directory of ".venv" is. This should normally
// be ".". The parameter is only useful for testing, where a caching directory
// can be passed.
func GetBinDir(parentDir string) (string, error) {
	var binDir string

	// Get the bin path
	_, err := os.Stat(parentDir + "/.venv")
	if err == nil {
		// Venv was found. Get bin path
		_, err := os.Stat(parentDir + "/.venv/bin")
		if err != nil {
			binDir = parentDir + "/.venv/Scripts"
		} else {
			binDir = parentDir + "/.venv/bin"
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

func GetPythonVersion(parentDir string) (string, error) {
	versionFilePath := parentDir + "/.python-version"

	info, err := os.Stat(versionFilePath)
	var version string

	if err == nil {
		// .python-version exists. It could be a directory. Check against this.
		if info.IsDir() {
			// This is a directory. Abort.
			return "", errors.New("Tried to get Python version from file '.python-version', but it is a directory.")
		}

		// File was found.
		data, err := os.ReadFile(versionFilePath)
		if err != nil {
			return "", errors.New("File '.python-version' was found, but it failed to be read.")
		}

		// TODO validate version
		version = string(data)
		if version == "" {
			return "", errors.New("The content of .python-version is empty. Please remove the file and re-run the command.")
		} else {
			fmt.Println("Found Python version: " + version)
		}

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
		err := os.WriteFile(versionFilePath, []byte(version), 0644)
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
//  1. Copying the Python interpreter and creating an environment for it
//  2. Unpacking pip that is included in the Python installation.
//
// Although the second step is not required because 'venv' does this for us, we
// can log the steps to the user so they are prepared for a small waiting time.
//
// createVenv allows specifying the parent directory for the new virtual
// environment directory. By default, this should be ".". This parameter is
// useful only in testing, where a cache directory can be passed.
//
// pythonPath is the command or filepath to the system Python interpreter. It
// is usually "python3", but the absolute filepath can be passed.
func CreateVenv(parentDir string, pythonPath string, requirementsFile string, verbose bool) error {
	fmt.Print("Project Python interpreter not found. ")
	fmt.Println("Creating virtual environment...")

	cmd := exec.Command(pythonPath, "-m", "venv", parentDir+"/.venv", "--without-pip")
	if err := cmd.Run(); err != nil {
		// Python not installed?
		return err
	}

	// On Windows, the executables are stored in .venv/Scripts, while it is
	// .venv/bin elsewhere.
	binDir, err := GetBinDir(parentDir)
	if err != nil {
		return err
	}

	// Install pip
	fmt.Println("Installing pip. This can take a few seconds...")
	cmd = exec.Command(binDir+"/python", "-m", "ensurepip")
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	// Some Python installations bundle pip as the `pip` command, and others as
	// `pip3`.
	pipCommand, err := FindPipCommand(binDir, verbose)
	if verbose {
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
		fmt.Println("Installing dependencies... This may fail if the HTTP connection to Pypi times out.")
		if verbose {
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
func InstallPythonVersion(version string) (string, error) {
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

// FindPipCommand tries to find the command name for the Pip package manager.
//
// Arguments
// ---------
// binDir (string): directory where the pip command is expected to be found in.
//
// Return
// ------
// string: Filepath to the pip executable file.
// error: Possible errors that may occur.
func FindPipCommand(binDir string, verbose bool) (string, error) {
	possibleCommands := []string{"pip", "pip3"}
	pipCommand := ""

	for _, v := range possibleCommands {
		cmd := exec.Command(binDir + "/" + v)
		if err := cmd.Run(); err == nil {
			pipCommand = binDir + "/" + v
		}
	}

	// Prevent using global pip command
	if _, err := exec.LookPath(pipCommand); err != nil {
		return "", errors.New("Pip command '" + pipCommand + "' was recognised but not found as a file.")
	}

	if verbose {
		fmt.Println("Pip command found at: " + pipCommand)
	}

	if pipCommand == "" {
		return "", errors.New("The filepath to command 'pip' was not found.")
	} else {
		return pipCommand, nil
	}
}

func ExecuteCmd(name string, args []string, venvParentDir string) error {
	if name == "" {
		return errors.New("Command name is missing")
	}

	binDir, err := GetBinDir(venvParentDir)
	if err != nil {
		return err
	}

	if _, err := exec.LookPath(binDir + "/" + name); err != nil {
		return errors.New("Command " + name + " is not found in " + binDir)
	}

	var cmdArgs []string
	if len(args) > 0 {
		cmdArgs = append([]string{}, args...)
	}

	cmd := exec.Command(
		binDir+"/"+name,
		cmdArgs...,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
