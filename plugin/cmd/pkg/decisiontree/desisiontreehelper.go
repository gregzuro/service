package decisiontree

import (
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/Sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

var HelperFunctions = map[string]govaluate.ExpressionFunction{
	"SlideingWindowGPSInArea": func(args ...interface{}) (interface{}, error) {
		testvar := 3.0
		return testvar, nil
	},
	"SlideingWindowWifiInArea": func(args ...interface{}) (interface{}, error) {
		testvar := 5.0
		return testvar, nil
	},
	"sendMessage": func(args ...interface{}) (interface{}, error) {
		log.WithFields(logrus.Fields{
			"State":     args[1].(string),
			"Operation": "Send Message",
			"Function":  "LocationUpdateThread",
		}).Debug("SendMessage")

		testvar := 5.0
		return testvar, nil
	},
	"SlideingWindowIBeaconInArea": func(args ...interface{}) (interface{}, error) {
		testvar := 5.0
		return testvar, nil
	},
	"ChangeState": func(args ...interface{}) (interface{}, error) {
		log.WithFields(logrus.Fields{
			"State":     args[1].(string),
			"Operation": "Change State",
			"Function":  "LocationUpdateThread",
		}).Debug("ChangeState")

		testvar := 5.0
		return testvar, nil
	},
	"NoAction": func(args ...interface{}) (interface{}, error) {
		testvar := 5.0
		return testvar, nil
	},
	"SendEmail": func(args ...interface{}) (interface{}, error) {
		m := gomail.NewMessage()
		emailFrom := args[1].(string)
		emailToList := args[2].(string)
		emailToArray := strings.Split(emailToList, ":")
		emailSubject := args[3].(string)
		emailBody := args[4].(string)
		m.SetHeader("From", emailFrom)
		m.SetHeader("To", emailToArray...)
		m.SetHeader("Subject", emailSubject)
		m.SetBody("text/html", emailBody)

		d := gomail.NewDialer("secure164.inmotionhosting.com", 465, "jthompson@contraxiom.com", "jt!@#")

		// Send the email to Bob, Cora and Dan.
		if err := d.DialAndSend(m); err != nil {
			log.WithFields(logrus.Fields{
				"Action":    "Not Sent",
				"Error":     err,
				"Operation": "SendEmail",
				"Function":  "DecisionTreeHelper",
			}).Debug("SendEmail")

		} else {
			log.WithFields(logrus.Fields{
				"Action":    "Sent",
				"Operation": "SendEmail",
				"Function":  "DecisionTreeHelper",
			}).Debug("SendEmail")
		}
		testvar := 5.0
		return testvar, nil
	},
}
