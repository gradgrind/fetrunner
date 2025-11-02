package w365tt

import (
	"encoding/json"
	"fetrunner/base"
	"fetrunner/db"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var inputfiles = []string{
	"../_testdata1/Versuch_D_Margin_hour_constraint_w365.json",
}

func TestFromJSON(t *testing.T) {
	// Test reading a _w365.json file without any processing.
	for _, fjson := range inputfiles {
		if _, err := os.Stat(fjson); err != nil {
			break
		}

		//stempath := strings.TrimSuffix(fjson, filepath.Ext(fjson))
		//logpath := stempath + ".log"
		base.OpenLog("")

		data := ReadJSON(fjson)

		//fmt.Printf("JSON struct: %#v\n", data)

		// Save as JSON
		f := strings.TrimSuffix(fjson, filepath.Ext(fjson)) + "_2.json"
		j, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(f, j, 0666); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("\n ***** JSON written to %s *****\n", f)
	}
}

func TestToDb(t *testing.T) {
	// Test reading a _w365.json file including initial processing.
	for _, fjson := range inputfiles {
		if _, err := os.Stat(fjson); err != nil {
			break
		}

		stempath := strings.TrimSuffix(fjson, filepath.Ext(fjson))
		//logpath := stempath + ".log"
		base.OpenLog("")

		dbn := db.NewDb()
		LoadJSON(dbn, fjson)

		// Save as JSON
		stempath = strings.TrimSuffix(stempath, "_w365")
		f := stempath + "_db.json"
		dbn.SaveDb(f)
		fmt.Printf("\n ***** JSON written to %s *****\n", f)
	}
}
