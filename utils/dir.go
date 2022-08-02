package utils

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"sync"
)

var m sync.Mutex

func GetTempDir() string {
	ret := "/tmp/alpr"
	if _, err := os.Stat(ret); os.IsNotExist(err) {
		m.Lock()
		defer m.Unlock()
		err = os.Mkdir(ret, 0777)
		if err != nil {
			log.Println(err.Error())
			return ret
		}
		wd, _ := os.Getwd()
		tempImageFileName := path.Join(wd, "static", "temp.jpg")
		data, err := ioutil.ReadFile(tempImageFileName)
		if err != nil {
			log.Println(err.Error())
			return ret
		}
		err = ioutil.WriteFile(path.Join(ret, "temp.jpg"), data, 0644)
		if err != nil {
			log.Println(err.Error())
			return ret
		}

	}
	return ret
}

func RemovePrevTempImageFiles() {
	root := GetTempDir()
	files, err := ioutil.ReadDir(root)
	if err != nil {
		log.Println("an error occurred while deleting pref temp image files, err: ", err.Error())
	}
	for _, file := range files {
		if len(file.Name()) > 8 { //uuid + .jpeg
			os.Remove(path.Join(root, file.Name()))
		}
	}
}
