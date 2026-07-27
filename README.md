# Python Runner

One of many Python wrappers on the internet. Its only purpose is to simplify the local Python development process.

*Before:*

```
python -m venv .venv
./.venv/scripts/activate
python app.py
```

*After:*

```
pyrun app.py
```

*Python Runner* uses the `python` and `pip` executables from an existing virtual environment (`venv/`). If the directory does not exist, then it creates one with `venv`. Dependencies are thereafter installed if `requirements.txt` is present.

## Installation
*Option 1:*

The easiest way to install `pyrun` is by downloading the executable for your platform from the [releases](https://github.com/asiangoldfish/python-runner/releases). After downloading the file, rename it to `pyrun` and add it to PATH.

*Option 2:*

Alternatively, it can be installed with the Go tooling:

```
go install github.com/asiangoldfish/pyrun@latest
```

*Option 3:*

To manually build and install the executable:

1. Build the program:
   ```
   go build -o bin/ cmd/pyrun/pyrun.go
   ```
2. Copy the executable to a directory in PATH. Example on Linux:
   ```
   cp ./bin/pyrun "$HOME/.local/bin"
   ```

For options 2 and 3, at least Go version 1.18 must be used.

## Usage
Once installed, `pyrun` can be used as follows:

```
pyrun version
```
