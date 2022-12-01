package main

import (
	"fmt"
	"os"
	"sote-farmer/packages"
	"strings"

	"github.com/integrii/flaggy"
	"gitlab.com/soteapps/packages/v2021/sLogger"
)

const (
	// Name of the Business Service
	LOGMESSAGEPREFIX = "sote-farmer"
	VERSION          = "v2021.1.1"
)

var (
	doExport   string
	doImport   bool
	doClean    bool
	doObscure  bool
	debugLevel bool

	// Start up values for the business service
	targetEnvironment string
	// Config file name
	configFile string

	// Bulk file name
	bulkFile string
)

func init() {
	sLogger.SetLogMessagePrefix(LOGMESSAGEPREFIX)

	// Set your program's name and description.  These appear in help output.
	flaggy.SetName(LOGMESSAGEPREFIX)
	flaggy.SetVersion(VERSION)
	flaggy.SetDescription(`
	Pulling data directly from the database.
	The harvester application can be run at anytime so long as the database is available.
		Constraints: 
			- does not read Cognito users or groups.
			- Information users are supported, system and company are not.
			- only one schema can be read at a time.
	
			Harvest File JSON format:
			{
				"table-group-name": "<output_file_name>",
				"schema": "<schema_name>",
				"tables": [
					"<table_name>"
				]
				
				,"custom_db": {  <Optional / Database Configuration>
					"name": "<database_name>",
					"user": "<username>",
					"password": "<password>",
					"host": "<hostname>",
					"port": <port_number>,
					"sslmode": "disable/allow/prefer/require"
				}
			}`)

	// You can disable various things by changing bool on the default parser
	// (or your own parser if you have created one).
	flaggy.DefaultParser.ShowHelpOnUnexpected = true

	// You can set a help prepend or append on the default parser.
	flaggy.DefaultParser.AdditionalHelpPrepend = "https://gitlab.com/getsote/utilities"

	flaggy.String(&doExport, "e", "export",
		"Export table data to (raw, csv or json) format, that exports the query result directly to the file.")

	flaggy.Bool(&doImport, "i", "import", "Load data into the database table(s) from a previously backed up file.")

	flaggy.Bool(&doClean, "ct", "clean-tables", "Erase all data in the table(s).")

	flaggy.String(&targetEnvironment, "t", "targetEnv", "Pulls configuration information from aws based on the environment supplied (custom:{{ENV}}|development|staging|demo|production).  "+
		"This requires that you have aws credentials/config setup on the system at ~/.aws. (default: '')")

	flaggy.String(&configFile, "c", "configFile", "The configuration file, contains the DATABASE schema and tables names.")

	flaggy.String(&bulkFile, "b", "bulkFile", "The database output file, uses the PostgreSQL copy protocol to perform bulk data insertion.")

	flaggy.Bool(&doObscure, "o", "obscure", "obscure fields")

	flaggy.Bool(&debugLevel, "d", "debug", "set the log level to DEBUG")

	flaggy.Parse()
}

func main() {
	sLogger.DebugMethod()

	if debugLevel {
		sLogger.SetLogLevelDebug()
	}

	helpConfigFile := configFile
	if helpConfigFile == "" {
		helpConfigFile = "sample.json"
	}

	helpClearTables := ""
	if doClean {
		helpClearTables = "--clean-tables"
	}

	targetEnv := strings.Split(targetEnvironment, ":")[0]

	if targetEnv == "" || (!doClean && !doImport && doExport == "") {
		flaggy.ShowHelpAndExit(`
Example:

Clean Table(s)
sote-farmer --targetEnv=custom:development --configFile=sample.json --clean-tables

Export Data
sote-farmer --targetEnv=staging --configFile=sample.json --export=raw
sote-farmer --targetEnv=staging --configFile=sample.json --export=csv
sote-farmer --targetEnv=staging --configFile=sample.json --export=json
sote-farmer --targetEnv=custom:production --configFile=sample.json --export=json

Import Data
sote-farmer --targetEnv=development --bulkFile=organizations.raw --import
sote-farmer --targetEnv=staging --bulkFile=organizations.json --import
sote-farmer --targetEnv=demo --bulkFile=organizations-csv.zip --import
sote-farmer --targetEnv=custom:development --configFile=sample.json --bulkFile=organizations.raw --import --clean-tables
		`)
	} else {
		if (targetEnv == "demo" || targetEnv == "production") && (doImport || doClean) {
			printAndExit(`
--import or --clean-tables are limited operations in "demo" and "production" environments.
			`)
		} else if doExport != "" && !strings.Contains("raw|csv|json", doExport) {
			printAndExit(fmt.Sprintf(`
Required: --export=(raw, csv or json) format

Example:

sote-farmer --targetEnv=%[1]s --configFile=%[2]s --export=raw
sote-farmer --targetEnv=%[1]s --configFile=%[2]s --export=csv
sote-farmer --targetEnv=%[1]s --configFile=%[2]s --export=json
			`, targetEnvironment, helpConfigFile))
		} else if doExport != "" && configFile == "" {
			printAndExit(fmt.Sprintf(`
Required: --configFile=(The configuration file)

Example:

sote-farmer --targetEnv=%s --configFile=sample.json --export=%s
				`, targetEnvironment, doExport))
		} else if !strings.Contains("custom:{{ENV}}|development|staging|demo|production", targetEnv) {
			var cmd string
			if doImport {
				if bulkFile == "" {
					cmd = fmt.Sprintf(`
sote-farmer --targetEnv=development --configFile=%[1]s --bulkFile=organizations.raw --import %[2]s
sote-farmer --targetEnv=development --configFile=%[1]s --bulkFile=organizations.json --import %[2]s
sote-farmer --targetEnv=development --configFile=%[1]s --bulkFile=organizations-csv.zip --import %[2]s
					`, helpConfigFile, helpClearTables)
				} else {
					cmd = fmt.Sprintf("sote-farmer --targetEnv=development --configFile=%s --bulkFile=%s --import %s", helpConfigFile, bulkFile, helpClearTables)
				}
			} else if doExport != "" {
				cmd = fmt.Sprintf("sote-farmer --targetEnv=staging --configFile=%s --export=%s", helpConfigFile, doExport)
			} else if doClean {
				cmd = fmt.Sprintf("sote-farmer --targetEnv=development --configFile=%s --clean-tables", helpConfigFile)
			}
			printAndExit(fmt.Sprintf(`
Required: --targetEnv=(custom:{{ENV}}|development|staging|demo|production)

Example:

%s
`, cmd))
		} else if doImport {
			if bulkFile != "" && targetEnv == "custom" && configFile == "" {
				printAndExit(fmt.Sprintf(`
Required: --configFile=(The configuration file)

Example:

sote-farmer --targetEnv=development --configFile=sample.json --bulkFile=%s --import %s
				`, bulkFile, helpClearTables))
			} else if bulkFile == "" {
				argConfigFile := ""
				if targetEnv == "custom" && configFile == "" {
					argConfigFile = " --configFile=sample.json"
				}
				printAndExit(fmt.Sprintf(`
Required: --bulkFile=(The database output file)

Example:

sote-farmer --targetEnv=%[1]s%[2]s --bulkFile=organizations.raw --import %[3]s
sote-farmer --targetEnv=%[1]s%[2]s --bulkFile=organizations.json --import %[3]s
sote-farmer --targetEnv=%[1]s%[2]s --bulkFile=organizations-csv.zip --import %[3]s
			`, targetEnvironment, argConfigFile, helpClearTables))
			}
		} else if doClean && configFile == "" {
			printAndExit(fmt.Sprintf(`
Required: --configFile=(The configuration file)

Example:

sote-farmer --targetEnv=%s --configFile=sample.json --clean-tables
				`, targetEnvironment))
		}
	}

	if doExport != "" {
		if soteErr := packages.DbExport(targetEnvironment, configFile, doExport, doObscure); soteErr.ErrCode != nil {
			panic(soteErr.FmtErrMsg)
		}
	} else if doImport {
		if soteErr := packages.DbImport(targetEnvironment, configFile, bulkFile, doClean); soteErr.ErrCode != nil {
			panic(soteErr.FmtErrMsg)
		}
	} else if doClean {
		if soteErr := packages.DbClean(targetEnvironment, configFile); soteErr.ErrCode != nil {
			panic(soteErr.FmtErrMsg)
		}
	}
}

func printAndExit(message string) {
	fmt.Println(message)
	os.Exit(3)
}
