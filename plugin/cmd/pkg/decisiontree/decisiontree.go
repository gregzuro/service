package decisiontree

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/Sirupsen/logrus"
	pb "github.com/gregzuro/service/plugin/cmd/locationeventspb"
	"github.com/gregzuro/service/plugin/cmd/pkg/deviceinfo"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type GContext struct {
	Device        string
	DeviceHistory *deviceinfo.DeviceHistory
	LandmarkPoiID string
}

func (gc *GContext) GetVariables(variableName string) interface{} {

	switch variableName {
	case "state":
		var deviceIndex int = 0
		if gc.DeviceHistory.CurrentIndex > 0 {
			deviceIndex = gc.DeviceHistory.CurrentIndex
		}
		return gc.DeviceHistory.DevHistory[deviceIndex].DeviceState //gc.DeviceHistory.DevHistory

	default:
		return ""
	}
}
func (gc *GContext) CheckVariables(variableName string) bool {

	switch variableName {
	case "state":
		return true
	default:
		return false
	}
}

type RunOnHitElement struct {
	branchName       interface{}
	functionCallName interface{}
}
type ResultElement struct {
	branchName       interface{}
	functionCallName interface{}
}
type BranchElement struct {
	branchName       interface{}
	DTElementIndex   int32
	FunctionCallName interface{}
	RElement         []ResultElement
	ROHElement       []RunOnHitElement
}
type DecisionElement struct {
	Variables     []string
	Sequence      interface{}
	Inbranch      interface{}
	Level         int64
	trueBranch    interface{}
	falseBranch   interface{}
	Property      interface{}
	Operation     interface{}
	Value         interface{}
	BranchID      int64
	expression    *govaluate.EvaluableExpression
	TrueBElement  BranchElement
	FalseBElement BranchElement
	Parameters    map[string]string
}

type Map map[string]json.RawMessage
type Array []json.RawMessage

const (
	DECISION State = 1 + iota
	BRANCH
	ELEMENT
	NOT_DEFINED
)

//var state State = NOT_DEFINED

type DecisionTreeElements struct {
	branchNameState    float64
	state              State
	BranchArray        []BranchElement
	DecisionArray      []DecisionElement
	runOnHitArray      []RunOnHitElement
	ResultArray        []ResultElement
	decisionArrayIndex int32
	branchArrayIndex   int32
	runOnHitArrayIndex int32
	resultArrayIndex   int32
	dtLevel            int32
	dtBranch           int32
	elementValue       *interface{}
}
type DecisionElementbySeq []DecisionElement

func (a DecisionElementbySeq) Len() int {
	return len(a)
}
func (a DecisionElementbySeq) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a DecisionElementbySeq) Less(i, j int) bool {
	return a[i].Level < a[j].Level
}

var currentBranch interface{}

var log = logrus.New()

func init() {
	log.Formatter = new(logrus.JSONFormatter)
	log.Formatter = new(logrus.TextFormatter) // default
	log.Level = logrus.DebugLevel
}

type State int

func CreateDecisionTree(raw json.RawMessage) DecisionTreeElements {
	dte := DecisionTreeElements{
		state:              NOT_DEFINED,
		BranchArray:        []BranchElement{},
		DecisionArray:      []DecisionElement{},
		runOnHitArray:      []RunOnHitElement{},
		ResultArray:        []ResultElement{},
		decisionArrayIndex: -1,
		branchArrayIndex:   -1,
		runOnHitArrayIndex: -1,
		resultArrayIndex:   -1,
		dtLevel:            0,
		dtBranch:           0,
	}
	walkJson(raw, &dte)

	return dte
}

//primes := [6]int{2, 3, 5, 7, 11, 13}
//jsonOrder := []string{"branch","descisions","branches","result","Sequence","truebranch","expression","Operation","Value"}
func walkJson(raw json.RawMessage, dte *DecisionTreeElements) {

	if raw[0] == 123 { //  123 is `{` => object

		var cont Map
		json.Unmarshal(raw, &cont)

		dst := string(raw[:])
		for k, _ := range cont {
			print(k)
		}
		log.WithFields(logrus.Fields{
			"Operation": "Debug Decision Tree",
			"Function":  "walkJson",
		}).Debug("**", dst, "$$")
		for i, v := range cont {
			print(v)

			if i == "expression" {
				dte.elementValue = &(dte.DecisionArray[dte.decisionArrayIndex].Property)

			} else if i == "decisions" {

				dte.dtLevel++
				dte.state = DECISION

			} else if i == "trueBranch" {
				dte.elementValue = &(dte.DecisionArray[dte.decisionArrayIndex].trueBranch)
			} else if i == "falseBranch" {
				dte.elementValue = &(dte.DecisionArray[dte.decisionArrayIndex].falseBranch)
			} else if i == "Inbranch" {
				dte.elementValue = &(dte.DecisionArray[dte.decisionArrayIndex].Inbranch)
			} else if i == "branch" {

				dte.elementValue = &(dte.BranchArray[dte.branchArrayIndex].branchName)

			} else if i == "branches" {
				dte.state = BRANCH

			} else if i == "Operation" {
				dte.elementValue = &(dte.DecisionArray[dte.decisionArrayIndex].Operation)

			} else if i == "Sequence" {
				dte.elementValue = &(dte.DecisionArray[dte.decisionArrayIndex].Sequence)

			} else if i == "Value" {
				dte.elementValue = &(dte.DecisionArray[dte.decisionArrayIndex].Value)

			} else if i == "runOnHit" {
			} else if i == "result" {
				//re := new(ResultElement)
				dte.elementValue = &(dte.BranchArray[dte.branchArrayIndex].FunctionCallName)
				//dte.ResultArray = append(dte.ResultArray, *re)
				//dte.resultArrayIndex++
				//dte.elementValue = &(dte.ResultArray[dte.resultArrayIndex].functionCallName)
				//dte.ResultArray[dte.resultArrayIndex].branchName = dte.BranchArray[dte.branchArrayIndex].branchName

			}

			walkJson(v, dte)

		}
	} else if raw[0] == 91 { // 91 is `[`  => array

		var cont Array
		json.Unmarshal(raw, &cont)
		dst := string(raw[:])

		log.WithFields(logrus.Fields{
			"Operation": "Debug Decision Tree",
			"Function":  "walkJson",
		}).Debug("++", dst, "--")

		for _, v := range cont {
			if dte.state == DECISION {
				de := new(DecisionElement)
				dte.DecisionArray = append(dte.DecisionArray, *de)
				dte.decisionArrayIndex++
			} else if dte.state == BRANCH {
				be := new(BranchElement)
				dte.BranchArray = append(dte.BranchArray, *be)
				dte.branchArrayIndex++
			}
			walkJson(v, dte)
		}

	} else {

		var val interface{}
		json.Unmarshal(raw, &val)
		if *(dte.elementValue) == nil {
			*(dte.elementValue) = val
		} else {
			for iba := len(dte.BranchArray) - 1; iba >= 0; iba-- {
				if dte.BranchArray[iba].branchName == nil {
					dte.BranchArray[iba].branchName = val
					break
				}
			}
		}
		//		*(dte.elementValue) = val

		log.WithFields(logrus.Fields{
			"Operation": "Debug Decision Tree",
			"Function":  "walkJson",
		}).Debug(val)

	}
}
func AddResultstoBranches(branches []BranchElement, results []ResultElement) []BranchElement {

	for i, bv := range branches {

		for _, rv := range results {
			if bv.branchName == rv.branchName {
				//re := new(ResultElement)
				branches[i].RElement = append(bv.RElement, rv)
			}
		}
	}
	return branches
}

func Split(r rune) bool {

	return r == ',' || r == ')' || r == '(' || r == '>' || r == '<' || r == '=' || r == '~' || r == '!' || r == '~' || r == '+' || r == '-' || r == '/' || r == '*' || r == '&' || r == '|' || r == '^' || r == '*' || r == '%' || r == '>' || r == '<'

}

func AddVariables(decisions []DecisionElement, gc GContext) []DecisionElement {

	for i, dv := range decisions {
		if dv.Property == nil {
			continue
		}
		Variables := strings.FieldsFunc(dv.Property.(string), Split)
		//Variables := strings.FieldsFunc("test", Split)
		for _, v := range Variables {
			if gc.CheckVariables(v) {
				decisions[i].Variables = append(decisions[i].Variables, v)
			}
		}

	}
	return decisions
}
func AddGovaluateExpressions(decisions []DecisionElement) []DecisionElement {
	var err error
	for i, de := range decisions {
		dep, OK := de.Property.(string)
		if OK {
			decisions[i].expression, err = govaluate.NewEvaluableExpressionWithFunctions(dep, HelperFunctions)
		} else {
			decisions[i].expression = nil
		}

		//decisions[i].expression, err = govaluate.NewEvaluableExpressionWithFunctions("state", helperFunctions)
		if err != nil {
			return nil
		}
	}

	return decisions

}
func AddLevel(decisions []DecisionElement) []DecisionElement {
	for i, _ := range decisions {
		LevelTest, lOk := (decisions[i].Sequence).(float64)
		branchTest, bOk := (decisions[i].Inbranch).(float64)
		if lOk {
			decisions[i].Level = int64(LevelTest)
		} else {
			decisions[i].Level = 0
		}

		if bOk {
			decisions[i].BranchID = int64(branchTest)
		} else {
			decisions[i].BranchID = 0
		}

	}
	sort.Sort(DecisionElementbySeq(decisions))
	for _, v := range decisions {
		if v.BranchID <= 0 {

			decisions = append(decisions[:0], decisions[1:]...)
			log.WithFields(logrus.Fields{
				"Operation": "Remove Emty Branch",
				"Function":  "Sort",
			}).Debug("AddLevel")
		}

	}
	return decisions
}
func Addjumps(decisions []DecisionElement, branches []BranchElement) []BranchElement {

	for i, bv := range branches {
		for ii, dv := range decisions {
			bidf, bidfOk := bv.branchName.(float64)
			bids, bidsOk := bv.branchName.(string)
			if bidfOk {
				if dv.BranchID == int64(bidf) {
					branches[i].DTElementIndex = int32(ii)
					break
				}
			} else if bidsOk {
				if string(dv.BranchID) == bids {
					branches[i].DTElementIndex = int32(ii)
					break
				}

			}
		}
	}
	return branches
}
func AddBranchestoDecisionTree(branches []BranchElement, decisions []DecisionElement) []DecisionElement {

	for i, dv := range decisions {

		for _, bv := range branches {
			if dv.trueBranch == bv.branchName {
				decisions[i].TrueBElement = bv

			}
			if dv.falseBranch == bv.branchName {
				decisions[i].FalseBElement = bv

			}
		}
	}
	return decisions
}
func DumpDecisionTree(decisions []DecisionElement, branches []BranchElement) {
	log.WithFields(logrus.Fields{
		"Operation": "Dump Decision Tree",
		"Function":  "dumpDecisionTree",
	}).Debug("Start Dump")
	for _, bv := range branches {

		log.WithFields(logrus.Fields{
			"Func":       bv.FunctionCallName,
			"BranchName": bv.branchName,
			"Function":   "dumpDecisionTree",
		}).Debug("True Function Calls")
	}
	for _, dv := range decisions {
		log.WithFields(logrus.Fields{
			"TrueJumpBranch":    dv.TrueBElement.DTElementIndex,
			"FalseJumpBranch":   dv.FalseBElement.DTElementIndex,
			"TrueBranchName":    dv.TrueBElement.branchName,
			"FalseBranchName":   dv.FalseBElement.branchName,
			"TrueFunctionName":  dv.TrueBElement.FunctionCallName,
			"FalseFunctionName": dv.FalseBElement.FunctionCallName,
			"Property":          dv.Property,
			"Operation:":        dv.Operation,
			"Value:":            dv.Value,
			"Sequence:":         dv.Sequence,
			"BranchID":          dv.BranchID,
			"Function":          "dumpDecisionTree",
		}).Debug("Branches")

		for _, re := range dv.TrueBElement.RElement {
			log.WithFields(logrus.Fields{
				"TrueFunc": re.functionCallName,
				"Function": "dumpDecisionTree",
			}).Debug("True Function Calls")
		}
		for _, re := range dv.FalseBElement.RElement {
			log.WithFields(logrus.Fields{
				"FalseFunc": re.functionCallName,
				"Function":  "dumpDecisionTree",
			}).Debug("True Function Calls")

		}
		for _, re := range dv.Variables {
			log.WithFields(logrus.Fields{
				"VariableName": re,
				"Function":     "dumpDecisionTree",
			}).Debug("Variabe Names")

		}

	}
	log.WithFields(logrus.Fields{
		"Operation": "Dump Decision Tree",
		"Function":  "dumpDecisionTree",
	}).Debug("End Dump")

}
func DumpDecisionTreeElements(decisions []DecisionElement) {
	log.WithFields(logrus.Fields{
		"Operation": "Dump Decision Tree",
		"Function":  "dumpDecisionTree",
	}).Debug("Start Dump")

	for _, dv := range decisions {
		log.WithFields(logrus.Fields{
			"TrueJumpBranch":    dv.TrueBElement.DTElementIndex,
			"FalseJumpBranch":   dv.FalseBElement.DTElementIndex,
			"TrueBranchName":    dv.TrueBElement.branchName,
			"FalseBranchName":   dv.FalseBElement.branchName,
			"TrueFunctionName":  dv.TrueBElement.FunctionCallName,
			"FalseFunctionName": dv.FalseBElement.FunctionCallName,
			"Property":          dv.Property,
			"Operation:":        dv.Operation,
			"Value:":            dv.Value,
			"Sequence:":         dv.Sequence,
			"BranchID":          dv.BranchID,
			"Function":          "dumpDecisionTree",
		}).Debug("Branches")

		for _, re := range dv.TrueBElement.RElement {
			log.WithFields(logrus.Fields{
				"TrueFunc": re.functionCallName,
				"Function": "dumpDecisionTree",
			}).Debug("True Function Calls")
		}
		for _, re := range dv.FalseBElement.RElement {
			log.WithFields(logrus.Fields{
				"FalseFunc": re.functionCallName,
				"Function":  "dumpDecisionTree",
			}).Debug("True Function Calls")

		}
		for _, re := range dv.Variables {
			log.WithFields(logrus.Fields{
				"VariableName": re,
				"Function":     "dumpDecisionTree",
			}).Debug("Variabe Names")

		}

	}
	log.WithFields(logrus.Fields{
		"Operation": "Dump Decision Tree",
		"Function":  "dumpDecisionTree",
	}).Debug("End Dump")

}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
func IsNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func ConvertPBDecisionTreeToInternalDecisionTree(decisions []*pb.DecisionElement) []DecisionElement {
	//var retDE []DecisionElement
	retDE := make([]DecisionElement, len(decisions))
	for i, v := range decisions {
		//	de := new(DecisionElement)
		//	retDE = append(retDE, *de)
		retDE[i].BranchID = v.BranchID
		retDE[i].Inbranch = v.Inbranch
		//de.expression = v.
		retDE[i].FalseBElement.DTElementIndex = v.FalseBElement.DTElementIndex
		retDE[i].FalseBElement.FunctionCallName = v.FalseBElement.FunctionCallName
		retDE[i].TrueBElement.DTElementIndex = v.TrueBElement.DTElementIndex
		retDE[i].TrueBElement.FunctionCallName = v.TrueBElement.FunctionCallName
		retDE[i].Property = v.Property
		retDE[i].Operation = v.Operation
		fv, err := strconv.ParseFloat(v.Value, 64)
		if err == nil {
			retDE[i].Value = fv
		} else {
			retDE[i].Value = v.Value
		}

		for ii, vi := range v.Variable {
			s := new(string)
			retDE[i].Variables = append(retDE[i].Variables, *s)
			retDE[i].Variables[ii] = vi
		}
		retDE[i].Parameters = make(map[string]string)
		for key, value := range v.Parameters {
			retDE[i].Parameters[key] = value //fmt.Println("Key:", key, "Value:", value)
		}

		log.WithFields(logrus.Fields{
			"TrueJumpBranch":    retDE[i].TrueBElement.DTElementIndex,
			"FalseJumpBranch":   retDE[i].FalseBElement.DTElementIndex,
			"TrueBranchName":    retDE[i].TrueBElement.branchName,
			"FalseBranchName":   retDE[i].FalseBElement.branchName,
			"TrueFunctionName":  retDE[i].TrueBElement.FunctionCallName,
			"FalseFunctionName": retDE[i].FalseBElement.FunctionCallName,
			"Property":          retDE[i].Property,
			"Operation:":        retDE[i].Operation,
			"Value:":            retDE[i].Value,
			"Sequence:":         retDE[i].Sequence,
			"BranchID":          retDE[i].BranchID,
			"Function":          "ConvertPBDecisionTreeToInternalDecisionTree",
		}).Debug("ConvertDT")
	}
	retDE = AddGovaluateExpressions(retDE)
	return retDE
}
func EvaluateTree(decisions []DecisionElement, gc GContext) {
	var err error
	var i int32 = 0
	for i >= 0 {
		//govaluate
		if decisions[i].BranchID <= 0 {
			i++
			if i >= int32(len(decisions)) {
				return
			}
			continue
		}
		parameters := make(map[string]interface{}, 8)

		var result interface{}
		if len(decisions[i].Variables) != 0 {
			parameters["gc"] = &gc
			for _, dtv := range decisions[i].Variables {
				parameters[dtv] = gc.GetVariables(dtv)
			}
			result, err = decisions[i].expression.Evaluate(parameters)
		} else {
			parameters["gc"] = &gc
			result, err = decisions[i].expression.Evaluate(parameters)
		}
		if err != nil {
			break
		}
		if EvaluateOperation(result,
			decisions[i].Operation.(string),
			decisions[i].Value) {

			sfcn, sfOK := decisions[i].TrueBElement.FunctionCallName.(string)

			if sfOK {
				parameters["gc"] = &gc
				result, err = decisions[i].expression.Evaluate(parameters)
				expression, err := govaluate.NewEvaluableExpressionWithFunctions(sfcn, HelperFunctions)
				if err == nil {
					expression.Evaluate(parameters)
					break
				} else {
					log.WithFields(logrus.Fields{
						"FunctionCallname": sfcn,
						"Error":            err,
						"Function":         "EvaluateTree",
					}).Debug("DecisionTree")
				}

			}

			if decisions[i].TrueBElement.DTElementIndex == 0 {
				if i >= int32(len(decisions))-1 || decisions[i].BranchID != decisions[i+1].BranchID {
					i = -1
				} else {
					i++
				}
			} else {
				i = decisions[i].TrueBElement.DTElementIndex
			}

		} else {

			sfcn, sfOK := decisions[i].FalseBElement.FunctionCallName.(string)
			if sfOK && sfcn != "" {
				parameters["gc"] = &gc
				result, err = decisions[i].expression.Evaluate(parameters)
				expression, err := govaluate.NewEvaluableExpressionWithFunctions(sfcn, HelperFunctions)
				if err == nil {
					expression.Evaluate(parameters)
					break
				} else {
					log.WithFields(logrus.Fields{
						"FunctionCallname": sfcn,
						"Error":            err,
						"Function":         "EvaluateTree",
					}).Debug("DecisionTree")
				}

			}

			//			for _, re := range decisions[i].FalseBElement.RElement {
			//				expression, _ := govaluate.NewEvaluableExpressionWithFunctions(*re.functionCallName.(*string), HelperFunctions)
			//				expression.Evaluate(nil)
			//			}

			if decisions[i].FalseBElement.DTElementIndex == 0 {
				if i >= int32(len(decisions))-1 || decisions[i].BranchID != decisions[i+1].BranchID {
					i = -1
				} else {
					i++
				}
			} else {
				i = decisions[i].FalseBElement.DTElementIndex
			}
		}
	}

}
func EvaluateOperation(PropertyInterface interface{}, Operation string, subjectDataInterface interface{}) bool {
	// TODO: if then else this
	pf, pfOk := PropertyInterface.(float64)
	pi, piOk := PropertyInterface.(int)
	sf, sfOk := subjectDataInterface.(float64)
	si, siOk := subjectDataInterface.(int)
	ps, psOk := PropertyInterface.(string)
	ss, ssOk := subjectDataInterface.(string)
	ssl, sslOk := subjectDataInterface.([]string)
	pb, pbOk := PropertyInterface.(bool)

	if !pfOk {
		pf = float64(pi)
		pfOk = piOk
	}
	if !sfOk {
		sf = float64(si)
		sfOk = siOk
	}
	if pbOk {
		psOk = true
		if pb == true {
			ps = "true"

		} else if pb == false {
			ps = "false"
		}
	}
	switch Operation {
	case ">":
		if pfOk && sfOk {
			if pf > sf {
				return true
			}
		}
		return false

	case ">=":
		if pfOk && sfOk {
			if pf >= sf {
				return true
			}
		}
		return false

	case "<":
		if pfOk && sfOk {
			if pf < sf {
				return true
			}
		}
		return false

	case "<=":
		if pfOk && sfOk {
			if pf <= sf {
				return true
			}
		}
		return false
	case "==":
		if pfOk && sfOk {
			if pf == sf {
				return true
			}
		} else if psOk && ssOk {
			if ps == ss {
				return true
			}
		}
		return false

	case "!=":
		if pfOk && sfOk {
			if pf != sf {
				return true
			}
		} else if psOk && ssOk {
			if ps != ss {
				return true
			}
		}
		return false

	case "in":
		if psOk && sslOk {
			return stringInSlice(ps, ssl)
		}
		return false

	case "nin":
		if psOk && sslOk {
			if ps != ss {
				return !stringInSlice(ps, ssl)
			}
		}
		return false

	default:
		return false
	}
	return false
}
