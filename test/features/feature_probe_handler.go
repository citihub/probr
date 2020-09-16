package features

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"

	"github.com/citihub/probr/internal/config"
	"github.com/citihub/probr/internal/coreengine"
)

//this is the "TEST HANDLER" impl  and will get called when probr is invoked from the CLI or API
//all we do here is set the godog args based on what has been supplied (e.g. output path)
//and call to the "feature" implementation (i.e the same impl when godog / go test is invoked)

//GodogTestHandler ...
func GodogTestHandler(gd *coreengine.GodogTest) (int, *bytes.Buffer, error) {
	if config.Vars.OutputType == "INMEM" {
		return InMemGodogTestHandler(gd)
	}
	return ToFileGodogTestHandler(gd)
}

func ToFileGodogTestHandler(gd *coreengine.GodogTest) (int, *bytes.Buffer, error) {
	o, err := GetOutputPath(&gd.TestDescriptor.Name)
	if err != nil {
		return -1, nil, err
	}
	status, err := runTestSuite(o, gd)

	//TODO - review!
	//FUDGE! If the tests are skipped due to tags, then an empty file may
	//be left lingering.  This will have a non-zero size as we've actually
	//had to create the file prior to the test run (see line 31).  If it's
	//less than 4 bytes, it's fairly certain that this will indeed be empty
	//and can be removed.
	i, err := o.Stat()
	s := i.Size()

	if s < 4 {
		err = os.Remove(o.Name())
		if err != nil {
			log.Printf("[WARN] error removing empty test result file: %v", err)
		}
	}
	return status, nil, err
}

func InMemGodogTestHandler(gd *coreengine.GodogTest) (int, *bytes.Buffer, error) {
	var t []byte
	o := bytes.NewBuffer(t)
	status, err := runTestSuite(o, gd)
	return status, o, err
}

func runTestSuite(o io.Writer, gd *coreengine.GodogTest) (int, error) {
	f, err := getFeaturesPath(gd)
	if err != nil {
		return -2, err
	}

	tags := config.Vars.Tests.Tags

	opts := godog.Options{
		Format: "cucumber",
		Output: colors.Colored(o),
		Paths:  []string{f},
		Tags:   tags,
	}

	status := godog.TestSuite{
		Name:                 gd.TestDescriptor.Name,
		TestSuiteInitializer: gd.TestSuiteInitializer,
		ScenarioInitializer:  gd.ScenarioInitializer,
		Options:              &opts,
	}.Run()

	return status, nil
}

func getFeaturesPath(gd *coreengine.GodogTest) (string, error) {
	r, err := GetRootDir()
	if err != nil {
		return "", fmt.Errorf("unable to determine root directory - not able to perform tests")
	}

	if gd.FeaturePath != nil {
		//if we've been given a feature path, add to root and return:
		return filepath.Join(r, *gd.FeaturePath), nil
	}

	//otherwise derive it from the group and category data:
	var g = gd.TestDescriptor.Group.String()
	var c = gd.TestDescriptor.Category.String()

	return filepath.Join(r, "test", "features",
		strings.ReplaceAll(strings.ToLower(g), " ", ""),
		strings.ReplaceAll(strings.ToLower(c), " ", ""), "features"), nil

}
