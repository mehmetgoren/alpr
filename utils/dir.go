package utils

import (
	"io/ioutil"
	"log"
	"os"
	"path"
)

func GetStaticDir() string {
	wd, _ := os.Getwd()
	return path.Join(wd, "static")
}

func RemovePrevTempImageFiles() {
	root := GetStaticDir()
	files, err := ioutil.ReadDir(root)
	if err != nil {
		log.Println("an error occurred while deleting pref temp image files, err: ", err.Error())
	}
	for _, file := range files {
		if len(file.Name()) > 5 { //uuid + .jpeg
			os.Remove(path.Join(root, file.Name()))
		}
	}
}
