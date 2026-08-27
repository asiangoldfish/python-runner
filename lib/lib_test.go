package lib

import (
	"os"
	"os/exec"
	"testing"
)

const CACHE_DIR = "../.test_cache"
const PYTHON_VERSION = "3.12"
const VERSION_FILE = ".python-version"
const DATA_DIR = "../data"

// TestGetPythonVersion tests whether the program successfully reads the correct
// Python version. We don't test entering a new Python version as this is
// interactive.
func TestGetPythonVersion(t *testing.T) {
	if _, err := os.Stat(CACHE_DIR); os.IsNotExist(err) {
		if err := os.Mkdir(CACHE_DIR, 0755); err != nil {
			t.Error(err.Error())
		}
	}

	// Scenario 1: Wrong version
	if err := os.WriteFile(CACHE_DIR+"/"+VERSION_FILE, []byte("3.14"), 0644); err != nil {
		t.Fatal("Failed to create or write to Python version file" + err.Error())
	}

	version, err := GetPythonVersion(CACHE_DIR)
	if err != nil {
		t.Error(err.Error())
	} else if version == "" {
		t.Error("Version is empty")
	} else if version == PYTHON_VERSION {
		t.Errorf("Expected version is NOT 3.12. Found version is: %s", version)
	}

	// Scenario 2: Empty version
	if err := os.WriteFile(CACHE_DIR+"/"+VERSION_FILE, []byte(""), 0644); err != nil {
		t.Fatal("Failed to create or write to Python version file" + err.Error())
	}

	version, err = GetPythonVersion(CACHE_DIR)
	if err == nil {
		t.Error("An error should be generated when there is Python version file is empty!")
	}

	// Scenario 3: Correct version
	if err := os.WriteFile(CACHE_DIR+"/"+VERSION_FILE, []byte(PYTHON_VERSION), 0644); err != nil {
		t.Fatal("Failed to create or write to Python version file" + err.Error())
	}

	version, err = GetPythonVersion(CACHE_DIR)
	if err != nil {
		t.Error(err.Error())
	} else if version == "" {
		t.Error("Version is empty")
	} else if version != PYTHON_VERSION {
		t.Errorf("Expected version is %s. Found version is: %s", PYTHON_VERSION, version)
	}

	// Clean up
	err = os.Remove(CACHE_DIR + "/" + VERSION_FILE)
}

func TestCreateEnv(t *testing.T) {
	// Fail cases
	// ----------
	err := CreateVenv(CACHE_DIR, "python"+PYTHON_VERSION, "requirements.txt", false)
	if err != nil {
		t.Error("The requirements file does not exist, but the command should still be successful.")
	}

	_, err = os.Stat(CACHE_DIR + "/.venv")
	if err != nil {
		t.Error("Expected to find '.venv/' in " + CACHE_DIR + ", but none was found.")
	}

	err = os.Remove(CACHE_DIR + "/.venv")
}

func TestInstallPackage(t *testing.T) {
	// This should fail as no packages are passed
	if err := InstallPackage([]string{}, false, CACHE_DIR); err == nil {
		t.Error("No packages were passed, so the test should fail.")
	}

	// Single package. This should be success
	if err := InstallPackage([]string{"numpy"}, false, CACHE_DIR); err != nil {
		t.Error("Tried to install pandas, but failed.")
	}

	// Multi-package. This should be success
	if err := InstallPackage([]string{"numpy", "pandas"}, false, CACHE_DIR); err != nil {
		t.Error("Tried to install numpy and pandas, but failed.")
	}

	if err := os.RemoveAll(CACHE_DIR + "/.venv"); err != nil {
		t.Error(err.Error())
	}
	if err := os.Remove(CACHE_DIR + "/requirements.txt"); err != nil {
		t.Error(err.Error())
	}
}

func TestFindPipCommand(t *testing.T) {
	// Prepare test
	cmd := exec.Command("python3", "-m", "venv", CACHE_DIR+"/.venv")

	if err := cmd.Run(); err != nil {
		t.Fatal(err.Error())
	}

	binDir, err := GetBinDir(CACHE_DIR)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Get pip
	if _, err := FindPipCommand(binDir, false); err != nil {
		t.Error(err.Error())
	}

	// Clean up
	if err := os.RemoveAll(CACHE_DIR + "/.venv"); err != nil {
		t.Error(err.Error())
	}
}

func TestExecuteCmd(t *testing.T) {
	// Prepare test
	cmd := exec.Command("python3", "-m", "venv", CACHE_DIR+"/.venv")

	if err := cmd.Run(); err != nil {
		t.Fatal(err.Error())
	}

	// Mkdocs is not installed yet...
	if err := ExecuteCmd("mkdocs", []string{"--help"}, CACHE_DIR); err == nil {
		t.Error("Test command 'mkdocs' is not installed yet. This should fail.")
	}

	binDir, err := GetBinDir(CACHE_DIR)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Install mkdocs
	cmd = exec.Command(binDir + "/pip3", "install", "mkdocs")
	if err := cmd.Run(); err != nil {
		t.Fatal(err.Error())
	}
	
	if err := ExecuteCmd("mkdocs", []string{"--help"}, CACHE_DIR); err != nil {
		t.Error("Test command 'mkdocs' is installed, but failed to run. Error: " + err.Error())
	}

	// Clean up
	// if err := os.RemoveAll(CACHE_DIR + "/.venv"); err != nil {
	// 	t.Error(err.Error())
	// }
}
