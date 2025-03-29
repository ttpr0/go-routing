package util

import (
	"testing"
)

type CVSSimpleTest struct {
	Name   string  `csv:"name"`
	Age    int     `csv:"age"`
	Height float32 `csv:"height"`
	Gender bool    `csv:"gender"`
}

func TestCSVSimple(t *testing.T) {
	file := "./testdata/simple.csv"

	i := 0
	for row := range ReadCSVFromFile[CVSSimpleTest](file, ';') {
		if i == 0 {
			if row.Name != "John" || row.Age != 30 || row.Height != 170 || row.Gender != false {
				t.Errorf("row.Name = %v; want name", row.Name)
			}
		} else if i == 1 {
			if row.Name != "Jane" || row.Age != 25 || row.Height != 160 || row.Gender != true {
				t.Errorf("row.Name = %v; want Jane", row.Name)
			}
		} else if i == 2 {
			if row.Name != "Joe" || row.Age != 35 || row.Height != 175 || row.Gender != true {
				t.Errorf("row.Name = %v; want Joe", row.Name)
			}
		} else {
			t.Errorf("too many rows")
		}
		i++
	}
}

func TestCSVError(t *testing.T) {
	file := "./testdata/error.csv"

	i := 0
	for row := range ReadCSVFromFile[CVSSimpleTest](file, ';') {
		if i == 0 {
			if row.Name != "John" || row.Age != 30 || row.Height != 170.5 || row.Gender != false {
				t.Errorf("row.Name = %v; want name", row.Name)
			}
		} else if i == 1 {
			if row.Name != "Jane" || row.Age != 25 || row.Height != 160.9 || row.Gender != true {
				t.Errorf("row.Name = %v; want Jane", row.Name)
			}
		} else if i == 2 {
			if row.Name != "'Joe" || row.Age != 35 || row.Height != 175.0 || row.Gender != true {
				t.Errorf("row.Name = %v; want Joe", row.Name)
			}
		} else if i == 3 {
			if row.Name != "" || row.Age != 28 || row.Height != 0 || row.Gender != false {
				t.Errorf("row.Name = %v; want Mark", row.Name)
			}
		} else {
			t.Errorf("too many rows")
		}
		i++
	}
}
