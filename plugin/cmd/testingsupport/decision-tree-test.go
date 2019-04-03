package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/Sirupsen/logrus"
	"github.com/gregzuro/service/plugin/cmd/pkg/decisiontree"
)

//"github.com/govaluate"
func check(e error) {
	if e != nil {
		panic(e)
	}
}

var log = logrus.New()

func init() {
	log.Formatter = new(logrus.JSONFormatter)
	log.Formatter = new(logrus.TextFormatter) // default
	log.Level = logrus.DebugLevel
}
func main() {

	status := decisiontree.EvaluateOperation(44.0, ">", 43)
	if status == true {
		log.WithFields(logrus.Fields{
			"Operation": "Evaluate Expression",
		}).Debug(status)
	}

	dat, err := ioutil.ReadFile("defaultEnterExitEmail.json")
	check(err)

	var tmpJ json.RawMessage

	json.Unmarshal(dat, &tmpJ)
	dte := decisiontree.CreateDecisionTree(tmpJ)
	dte.DecisionArray = decisiontree.AddLevel(dte.DecisionArray)
	dte.BranchArray = decisiontree.Addjumps(dte.DecisionArray, dte.BranchArray)
	dte.BranchArray = decisiontree.AddResultstoBranches(dte.BranchArray, dte.ResultArray)
	dte.DecisionArray = decisiontree.AddBranchestoDecisionTree(dte.BranchArray, dte.DecisionArray)
	var gc decisiontree.GContext
	gc.Device = "12345"
	dte.DecisionArray = decisiontree.AddVariables(dte.DecisionArray, gc)
	dte.DecisionArray = decisiontree.AddGovaluateExpressions(dte.DecisionArray)
	decisiontree.DumpDecisionTree(dte.DecisionArray, dte.BranchArray)
	decisiontree.EvaluateTree(dte.DecisionArray, gc)
}
