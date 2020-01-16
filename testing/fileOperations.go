package testing

import (
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
)

func ReadFile(pathToFile string) (string, error) {
	data, err := ioutil.ReadFile(pathToFile)
	if err != nil {
		return "", errors.Wrap(err, "Error in reading file (Maybe file name you entered is wrong)")
	}
	return string(data), nil
}

func WriteDataToFile(data string, pathToFile string) error {
	err := ioutil.WriteFile(pathToFile, []byte(data), 0644)
	if err != nil {
		return errors.Wrap(err, "unable to write data into file")
	}
	return nil
}

func CreateDirectory(pathToDirectory string) error {
	if _, err := os.Stat(pathToDirectory); os.IsNotExist(err) {
		if err = os.Mkdir(pathToDirectory, 0777); err != nil {
			return errors.Wrap(err, "unable to create directory")
		}
	} else {
		if err = os.RemoveAll(pathToDirectory); err != nil {
			return errors.Wrap(err, "unable to deleter directory"+pathToDirectory)
		}
		if err = os.Mkdir(pathToDirectory, 0777); err != nil {
			return errors.Wrap(err, "unable to make directory"+pathToDirectory)
		}
	}
	return nil
}

func CopyFiles(sourcePath string, destinationPath string) error {
	input, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		return errors.Wrap(err, "unable to read file ")
	}

	err = ioutil.WriteFile(destinationPath, input, 0644)
	if err != nil {
		return errors.Wrap(err, "unable to write in file ")
	}
	return nil
}
