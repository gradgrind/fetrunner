package db

import (
	"encoding/json"
	"fetrunner/base"
	"io"
	"os"
)

func (db *DbTopLevel) SaveDb(fpath string) bool {
	// Save as JSON
	j, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		db.Logger.Error("%v", err)
		return false
	}
	if err := os.WriteFile(fpath, j, 0644); err != nil {
		db.Logger.Error("%v", err)
		return false
	}
	return true
}

func LoadDb(logger base.BasicLogger, fpath string) (*DbTopLevel, error) {
	// Open the  JSON file
	jsonFile, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	// Remember to close the file at the end of the function
	defer jsonFile.Close()
	// read the opened XML file as a byte array.
	byteValue, _ := io.ReadAll(jsonFile)
	logger.Info("*+ Reading: %s\n", fpath)
	v := NewDb(logger)
	err = json.Unmarshal(byteValue, v)
	if err != nil {
		return nil, err
	}
	v.initElements()
	return v, nil
}

func (db *DbTopLevel) testElement(ref NodeRef, element Element) bool {
	if ref == "" {
		db.Logger.Error("Element has no Id:\n  -- %+v\n", element)
		return false
	}
	_, nok := db.Elements[ref]
	if nok {
		db.Logger.Error("Element Id defined more than once:\n  %s\n", ref)
		return false
	}
	db.Elements[ref] = element
	return true
}

func (db *DbTopLevel) initElements() {
	for _, e := range db.Days {
		db.testElement(e.Id, e)
	}
	for _, e := range db.Hours {
		db.testElement(e.Id, e)
	}
	for _, e := range db.Teachers {
		db.testElement(e.Id, e)
	}
	for _, e := range db.Subjects {
		db.testElement(e.Id, e)
	}
	for _, e := range db.Rooms {
		db.testElement(e.Id, e)
	}
	for _, e := range db.RoomGroups {
		db.testElement(e.Id, e)
	}
	for _, e := range db.RoomChoiceGroups {
		db.testElement(e.Id, e)
	}
	for _, e := range db.Groups {
		db.testElement(e.Id, e)
	}
	for _, e := range db.Classes {
		db.testElement(e.Id, e)
	}
	for _, e := range db.Courses {
		db.testElement(e.Id, e)
	}
	for _, e := range db.SuperCourses {
		db.testElement(e.Id, e)
	}
	for _, e := range db.SubCourses {
		db.testElement(e.Id, e)
	}
	for _, e := range db.Activities {
		db.testElement(e.Id, e)
	}
	//TODO: Handle Constraints
}
