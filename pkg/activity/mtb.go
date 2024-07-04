package activity

import (
	"log"
	"os"
	"time"
)

func CreateActivity(name string, date time.Time, rating int) error {

	mtb := &Activity{
		category: "mtb",
		name:     name,
		date:     date,
		rating:   rating,
	}

	err := mtb.createFolder()
	if err != nil {
		panic("error creating folder")
	}

	for _, file := range []string{"description.md", "elevation.plt", "Makefile", "img-even.tex"} {

		text, err := mtb.initSkeleton(file)
		if err != nil {
			panic("error initializing skeleton")
		}

		file, err := os.Create(mtb.textLocation + "/" + file)
		if err != nil {
			log.Printf("Failed to create file: %v", err)
		}
		defer file.Close()

		_, err = file.WriteString(text)
		if err != nil {
			log.Printf("Failed to write to file: %v", err)
		}
	}

	return nil
}
