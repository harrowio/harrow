package diff

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
)

type ChangeType string

var (
	Addition = ChangeType("added")
	Removal  = ChangeType("removed")
	Context  = ChangeType("context")
)

type Change struct {
	Kind ChangeType
	Line int
	Text string
}

func (self *Change) MarshalJSON() ([]byte, error) {
	payload := []interface{}{
		string(self.Kind),
		self.Line,
		self.Text,
	}

	return json.Marshal(payload)
}

func (self *Change) String() string {
	kind := ""
	switch self.Kind {
	case Addition:
		kind = "+"
	case Removal:
		kind = "-"
	case Context:
		kind = " "
	}

	return fmt.Sprintf("%d %s%s\n", self.Line, kind, self.Text)
}

func (self *Change) UnmarshalJSON(data []byte) error {
	result := []interface{}{}
	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}

	if len(result) != 3 {
		return fmt.Errorf("diff.Change: expected list of length 3, got %d", len(result))
	}

	if kind, ok := result[0].(string); ok {
		self.Kind = ChangeType(kind)
	} else {
		return fmt.Errorf("diff.Change: kind not a string, got %T", result[0])
	}

	if line, ok := result[1].(float64); ok {
		self.Line = int(line)
	} else {
		return fmt.Errorf("diff.Change: line not a number, got %T", result[1])
	}

	if text, ok := result[2].(string); ok {
		self.Text = text
	} else {
		return fmt.Errorf("diff.Change: text not a string, got %T", result[2])
	}

	return nil
}

// TempFilePrefix is the filename prefix used for temporary files
// written for interfacing with the `diff` tool.
const TempFilePrefix = "harrow-diff-"

// Changes returns a list of changes extracted from a unified diff of
// a and b.
func Changes(a, b []byte) ([]*Change, error) {
	diffOutput, err := Unified(a, b)
	if err != nil {
		return nil, err
	}

	if len(diffOutput) == 0 {
		return []*Change{}, nil
	}

	result := []*Change{}

	lines := bytes.Split(diffOutput, []byte("\n"))

	currentLine := 0
	if !bytes.HasPrefix(lines[currentLine], []byte(`---`)) {
		return nil, fmt.Errorf("expected `---`, got %q", lines[currentLine])
	}
	currentLine++

	if !bytes.HasPrefix(lines[currentLine], []byte(`+++`)) {
		return nil, fmt.Errorf("expected `+++`, got %q", lines[currentLine])
	}
	currentLine++

	if !bytes.HasPrefix(lines[currentLine], []byte(`@@`)) {
		return nil, fmt.Errorf("expected `@@`, got %q", lines[currentLine])
	}
	changes, rest := parseHunk(lines[currentLine:])
	result = append(result, changes...)
	for len(rest) > 0 && bytes.HasPrefix(rest[0], []byte(`@@`)) {
		changes, rest = parseHunk(rest)
		result = append(result, changes...)
	}

	return result, nil
}

func parseHunk(lines [][]byte) ([]*Change, [][]byte) {
	currentLine := 0
	removals := 0
	sourceLine := 0
	contexts := 0
	additions := 0
	changes := []*Change{}

	fmt.Sscanf(string(lines[0]), "@@ -%d,", &contexts)
	contexts--
	currentLine++
	for currentLine < len(lines) && !bytes.HasPrefix(lines[currentLine], []byte(`@@`)) {
		kind := Context
		if len(lines[currentLine]) == 0 {
			currentLine++
			continue
		}

		if currentLine >= len(lines) {
			break
		}

		switch lines[currentLine][0] {
		case '-':
			kind = Removal
			removals++
			sourceLine = contexts + removals
		case '+':
			kind = Addition
			additions++
			sourceLine = contexts + additions
		case ' ':
			kind = Context
			contexts++
			sourceLine = contexts + additions
		}

		changes = append(changes, &Change{
			Kind: kind,
			Line: sourceLine,
			Text: string(lines[currentLine][1:]),
		})

		currentLine++
	}

	return changes, lines[currentLine:]
}

// Unified returns a unified textual diff of a and b.
func Unified(a, b []byte) ([]byte, error) {
	aFilename, err := writeToTempFile(a)
	defer os.Remove(aFilename)
	if err != nil {
		return nil, err
	}

	bFilename, err := writeToTempFile(b)
	defer os.Remove(bFilename)
	if err != nil {
		return nil, err
	}

	runDiff := exec.Command("diff", "-u", aFilename, bFilename)
	diffBytes, err := runDiff.Output()

	if err == nil {
		return diffBytes, nil
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		if waitStatus, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			if waitStatus.ExitStatus() == 1 {
				return diffBytes, nil
			}
		}
	}

	return nil, err
}

func writeToTempFile(contents []byte) (string, error) {
	f, err := ioutil.TempFile("", TempFilePrefix)
	if err != nil {
		return "", err
	}

	_, err = f.Write(contents)

	return f.Name(), err
}
