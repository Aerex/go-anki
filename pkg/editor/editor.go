package editor

import (
	"bytes"
	"fmt"
	goio "io"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/aerex/go-anki/pkg/io"
	"github.com/aerex/go-anki/pkg/template"
	"github.com/spf13/viper"
	"gopkg.in/AlecAivazis/survey.v1"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Editor
type Editor interface {
	Edit(in interface{}) (error, []byte, bool)
	Clone() (err error)
	Create() error
	Remove() error
	ConfirmUserError() bool
}

type modelEditor struct {
	FilePath     string
	OrigFilePath string
	Program      string
	IO           *io.IO

	templates template.Template
	retry     bool
}

func NewModelEditor(tmpl template.Template, io *io.IO) Editor {
	return &modelEditor{
		templates: tmpl,
		Program:   getEditor(),
		IO:        io,
	}
}

// Edit writes a card note model to a temp yaml file
func (e *modelEditor) Edit(in interface{}) (error, []byte, bool) {
	// copy over orignal data so user can start fresh when editing

	// Write template to file (e.in)
	if !e.retry {
		if err := e.templates.Execute(in, e.IO); err != nil {
			return err, []byte{}, false
		}
	}

	// close file then clone file to restore so user has
	// a fresh state if  error occurs after editing
	if err := e.Clone(); err != nil {
		return err, []byte{}, false
	}

	// Open file using editor
	if err := e.IO.Eval(fmt.Sprintf("%s %s", e.Program, e.FilePath), nil); err != nil {
		return err, []byte{}, false
	}

	changed, err := e.hasChanges()
	if err != nil {
		return err, []byte{}, false
	}

	// do not continue if file has not changed
	if !changed {
		return nil, []byte{}, false
	}

	// Open file then unmarshall file content to model
	data, err := ioutil.ReadFile(e.FilePath)
	if err != nil {
		return err, []byte{}, changed
	}

	return nil, data, changed
}

func getEditor() string {
	editor := viper.GetString("EDITOR")
	if editor == "" {
		if runtime.GOOS == "windows" {
			return "notepad.exe"
		}
		return "vim"
	}
	return editor
}

func (e *modelEditor) Clone() (err error) {
	// Get filename without ext
	var src, dst *os.File
	_, tmpFileName := path.Split(e.FilePath)
	tmpFileName = strings.Replace(tmpFileName, ".yml", "", -1)

	if src, err = os.Open(e.FilePath); err == nil {
		defer src.Close()
		if dst, err = os.Create(fmt.Sprintf("/tmp/%s.orig", tmpFileName)); err == nil {
			e.OrigFilePath = dst.Name()
			if _, err = goio.Copy(dst, src); err != nil {
				defer dst.Close()
				return
			}
		}
	}
	return
}

// Create will create a new file for a entity
func (e *modelEditor) Create() error {
	f, err := os.CreateTemp("", "anki-editor-*.yml")

	if err != nil {
		return err
	}
	e.FilePath = f.Name()
	e.IO = &io.IO{Output: f}

	return nil
}

// hasChanges will check to see if the edited file has been changed.
// For instance, adding a content to a new card template.
// To determine if a file had been changed the following conditions are checked
// 1) File Size
// 2) File Content
func (e *modelEditor) hasChanges() (bool, error) {
	// Open edited file and read by buffer pages

	file, err := os.Open(e.FilePath)
	if err != nil {
		return false, err
	}
	defer file.Close()
	statEdited, err := file.Stat()
	if err != nil {
		return false, err
	}

	origf, err := os.Open(e.OrigFilePath)
	statOrig, err := origf.Stat()
	if err != nil {
		return false, err
	}
	defer func() {
		os.RemoveAll(e.OrigFilePath)
		origf.Close()
	}()

	// if the edited file size has changed we can
	// assume the file has changed
	if statOrig.Size() != statEdited.Size() {
		return true, nil
	}

	buf1, buf2 := make([]byte, statOrig.Size()), make([]byte, statOrig.Size())
	var fileErr1, fileErr2 error
	var cur1, cur2 int
	for fileErr1 != goio.EOF || fileErr2 != goio.EOF {
		cur1, fileErr1 = file.Read(buf1)
		cur2, fileErr2 = origf.Read(buf2)

		// as we read if at cursor position of each buffer is not the same
		// we can assume that the file has changed
		if cur1 != cur2 {
			return true, nil
		}

		// if the remaning bytes are not equal beyond the current cursor position of
		// each buffer we can asssume the file has changed
		if !bytes.Equal(buf1[:cur1], buf2[:cur2]) {
			return true, nil
		}
	}

	return false, nil
}

// Remove removes all editor files including copies
func (e *modelEditor) Remove() (err error) {
	if err = os.Remove(e.FilePath); err != nil {
		return
	}
	if err = os.Remove(e.OrigFilePath); err != nil {
		return
	}
	return
}

func (e *modelEditor) ConfirmUserError() bool {
	var confirm bool
	survey.AskOne(
		&survey.Confirm{
			Message: "Invalid YAML. Edit file again?",
		}, &confirm, nil)
	e.retry = true
	return confirm
}
