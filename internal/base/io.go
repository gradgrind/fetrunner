package base

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
)

var (
	TEMPORARY_BASEDIR string
	TEMPORARY_DIR     string
)

func (bd *BaseData) SetTmpDir() {
	logger := bd.Logger
	if TEMPORARY_BASEDIR == "" {
		TEMPORARY_BASEDIR = os.TempDir()
	}
	tmpdir := filepath.Join(TEMPORARY_BASEDIR, "fetrunner")
	fileInfo, err := os.Stat(tmpdir)
	if errors.Is(err, os.ErrNotExist) {
		err = os.Mkdir(tmpdir, 0700)
		if err != nil {
			logger.Error("CREATE_DIRECTORY_FAILED: %s", tmpdir)
			TEMPORARY_BASEDIR = ""
			return
		}
	} else if !fileInfo.IsDir() {
		logger.Error("NOT_A_DIRECTORY: %s", tmpdir)
		TEMPORARY_BASEDIR = ""
		return
	}
	logger.Info("TEMPORARY_DIR: %s", tmpdir)
	TEMPORARY_DIR = tmpdir
}

func (bd *BaseData) SaveDb(fpath string) bool {
	// Save as JSON
	j, err := json.MarshalIndent(bd.Db, "", "  ")
	if err != nil {
		bd.Logger.Error("%v", err)
		return false
	}
	if err := os.WriteFile(fpath, j, 0644); err != nil {
		bd.Logger.Error("%v", err)
		return false
	}
	return true
}

func (bd *BaseData) LoadDb(fpath string) error {
	// Open the  JSON file
	jsonFile, err := os.Open(fpath)
	if err != nil {
		return err
	}
	// Remember to close the file at the end of the function
	defer jsonFile.Close()
	// read the opened XML file as a byte array.
	byteValue, _ := io.ReadAll(jsonFile)
	bd.Logger.Info("*+ Reading: %s", fpath)
	v := NewDb()
	err = json.Unmarshal(byteValue, v)
	if err != nil {
		return err
	}
	bd.Db = v
	bd.initElements()
	return nil
}

func (bd *BaseData) testElement(ref NodeRef, element Element) bool {
	if ref == "" {
		bd.Logger.Error("Element has no Id:\n  -- %+v", element)
		return false
	}
	_, nok := bd.Db.Elements[ref]
	if nok {
		bd.Logger.Error("Element Id defined more than once:\n  %s", ref)
		return false
	}
	bd.Db.Elements[ref] = element
	return true
}

func (bd *BaseData) initElements() {
	for _, e := range bd.Db.Days {
		bd.testElement(e.Id, e)
	}
	for _, e := range bd.Db.Hours {
		bd.testElement(e.Id, e)
	}
	for _, e := range bd.Db.Teachers {
		bd.testElement(e.Id, e)
	}
	for _, e := range bd.Db.Subjects {
		bd.testElement(e.Id, e)
	}
	for _, e := range bd.Db.Rooms {
		bd.testElement(e.Id, e)
	}
	for _, e := range bd.Db.RoomGroups {
		bd.testElement(e.Id, e)
	}
	for _, e := range bd.Db.RoomChoiceGroups {
		bd.testElement(e.Id, e)
	}
	for _, e := range bd.Db.Groups {
		bd.testElement(e.Id, e)
	}
	for _, e := range bd.Db.Classes {
		bd.testElement(e.Id, e)
	}
	for _, e := range bd.Db.Courses {
		bd.testElement(e.Id, e)
	}
	for _, e := range bd.Db.SuperCourses {
		bd.testElement(e.Id, e)
	}
	for _, e := range bd.Db.SubCourses {
		bd.testElement(e.Id, e)
	}
	for _, e := range bd.Db.Activities {
		bd.testElement(e.Id, e)
	}
	//TODO: Handle Constraints
}
