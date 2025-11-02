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
		base.Error.Println(err)
		return false
	}
	if err := os.WriteFile(fpath, j, 0666); err != nil {
		base.Error.Println(err)
		return false
	}
	return true
}

func LoadDb(fpath string) *DbTopLevel {
	// Open the  JSON file
	jsonFile, err := os.Open(fpath)
	if err != nil {
		base.Error.Fatal(err)
	}
	// Remember to close the file at the end of the function
	defer jsonFile.Close()
	// read the opened XML file as a byte array.
	byteValue, _ := io.ReadAll(jsonFile)
	base.Message.Printf("*+ Reading: %s\n", fpath)
	v := NewDb()
	err = json.Unmarshal(byteValue, v)
	if err != nil {
		base.Error.Fatalf("Could not unmarshal json: %s\n", err)
	}
	v.initElements()
	return v
}

func (db *DbTopLevel) testElement(ref NodeRef, element Element) {
	if ref == "" {
		base.Error.Fatalf("Element has no Id:\n  -- %+v\n", element)
	}
	_, nok := db.Elements[ref]
	if nok {
		base.Error.Fatalf("Element Id defined more than once:\n  %s\n", ref)
	}
	db.Elements[ref] = element
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
