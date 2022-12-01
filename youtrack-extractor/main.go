package main

import (
	"youtrack-extractor/packages"

	"github.com/integrii/flaggy"
	"gitlab.com/soteapps/packages/v2021/sLogger"
)

const (
	// LOGMESSAGEPREFIX Name of the Business Service
	LOGMESSAGEPREFIX = "youtrack-extractor"
	// VERSION of the Business Service - THIS MUST MATCH THE RELEASE TAG!!!!!
	VERSION = "v2021.1.0"
	// FOURSPACES Formatting for the application display when run with help
	FOURSPACES string = "    "
)

var (
	// Start up values for the business service
	debugMode bool
)

func init() {
	sLogger.SetLogMessagePrefix(LOGMESSAGEPREFIX)

	appDescription := "List of values (LOV) Business Service.\nThis manages all additions, changes, and removal to list of values.\n" +
		"\nVersion: \n" +
		FOURSPACES + "- " + VERSION + "\n" +
		"\nConstraints: \n" +
		FOURSPACES + "- At start up, you must pass the application name and the environment for the business service.\n" +
		FOURSPACES + "- There is no CLI for this business service.\n" +
		"\nNotes:\n" +
		FOURSPACES + "None\n"

	// Set your program's name and description.  These appear in help output.
	flaggy.SetName("youtrack-data")
	flaggy.SetDescription(appDescription)

	// You can disable various things by changing bool on the default parser
	// (or your own parser if you have created one).
	flaggy.DefaultParser.ShowHelpOnUnexpected = true

	// You can set a help prepend or append on the default parser.
	flaggy.DefaultParser.AdditionalHelpPrepend = "https://gitlab.com/soteapps/packages"

	flaggy.Bool(&debugMode, "", "debugMode", "Output debug message to the log - Be careful as the logs may get overloaded.")

	// Set the version and parse all inputs into variables.
	flaggy.SetVersion(VERSION)
	flaggy.Parse()
}

func main() {
	sLogger.DebugMethod()

	// TODO Make the debug mode changeable via a message https://sote.myjetbrains.com/youtrack/issue/BE21-144
	if debugMode {
		sLogger.SetLogLevelDebug()
	}

	//if soteErr := packages.Run(applicationName, targetEnvironment, "", "", false); soteErr.ErrCode != nil {
	//	panic(soteErr.FmtErrMsg)
	//}

	youTrackManagerPtr, soteErr := packages.New("https://sote.myjetbrains.com", "perm:U2NvdHQ=.NTMtMQ==.M7WOkZgn5RWggrIdTPhkTa6fXDmF8h")

	if soteErr.ErrCode != nil {
		panic(soteErr.FmtErrMsg)
	}
	packages.ExportJson(youTrackManagerPtr, false)

}
