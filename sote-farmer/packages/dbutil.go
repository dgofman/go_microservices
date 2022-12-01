package packages

import (
	"archive/zip"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/nwtgck/go-fakelish"
	"gitlab.com/soteapps/packages/v2021/sConfigParams"
	"gitlab.com/soteapps/packages/v2021/sDatabase"
	"gitlab.com/soteapps/packages/v2021/sError"
	"gitlab.com/soteapps/packages/v2021/sLogger"
)

var (
	dbConnInfo          sDatabase.ConnInfo
	csvFileZipExtension = "-csv.zip"
)

type FileInfo struct {
	Format string  `json:"format"`
	Env    string  `json:"environment"`
	Schema string  `json:"schema"`
	Tables []Table `json:"tables"`
}

type Table struct {
	Name    string          `json:"name"`
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
}

type Config struct {
	Name     string                       `json:"table-group-name"`
	Schema   string                       `json:"schema"`
	Tables   []string                     `json:"tables"`
	Obscure  map[string]map[string]string `json:"obscure"`
	Custom   map[string]*ConfigCustomDb   `json:"custom_db"` //optional settings
	customDb *ConfigCustomDb
}

type ConfigCustomDb struct {
	DBName     string `json:"name"`
	DBUser     string `json:"user"`
	DBPassword string `json:"password"`
	DBHost     string `json:"host"`
	DBSSLMode  string `json:"sslmode"`
	DBPort     int    `json:"port"`
}

func initGob() {
	// Register custom datatype for encoding/gob
	gob.Register(time.Now()) //Register time.Time
}

func isCustom(targetEnvironment string) bool {
	return strings.Split(targetEnvironment, ":")[0] == "custom"
}

// Database connection
func initConnInfo(targetEnvironment string, config *ConfigCustomDb) (dbConnInfo sDatabase.ConnInfo, soteErr sError.SoteError) {
	// Using a custom environment, you can apply the database settings from the configuration file.
	if isCustom(targetEnvironment) {
		// Connects to database using custom configuration
		dbConnInfo, soteErr = sDatabase.GetConnection(config.DBName, config.DBUser, config.DBPassword, config.DBHost, config.DBSSLMode, config.DBPort, 3)
	} else {
		// Verify AWS supported environments
		os.Setenv("APP_ENVIRONMENT", targetEnvironment)
		if soteErr = sConfigParams.ValidateEnvironment(targetEnvironment); soteErr.ErrCode == nil {
			if soteErr = sDatabase.GetAWSParams(); soteErr.ErrCode == nil {
				// Connects to the Sote Database
				dbConnInfo, soteErr = sDatabase.GetConnection(sDatabase.DBName, sDatabase.DBUser, sDatabase.DBPassword, sDatabase.DBHost,
					sDatabase.DBSSLMode, sDatabase.DBPort, 3)
			}
		}
	}
	return
}

// Loading configuration file
func loadConfig(targetEnvironment string, configFilePath string) (config Config) {
	file, err := os.Open(configFilePath)
	if err != nil {
		panic(err)
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}
	targetEnv := strings.Split(targetEnvironment, ":")
	if targetEnv[0] == "custom" {
		if len(targetEnv) != 2 {
			panic("Required parameter custom_db:{{ENV}}")
		}
		if config.Custom[targetEnv[1]] == nil {
			panic("Could not load database configuation custom_db:" + targetEnv[1])
		}
		config.customDb = config.Custom[targetEnv[1]]
	}
	return
}

// Export database table records into an external file
func DbExport(targetEnvironment, configFilePath, exportFormat string, doObscure bool) (soteErr sError.SoteError) {
	sLogger.DebugMethod()
	sLogger.Info("Exporting...")

	var (
		config = loadConfig(targetEnvironment, configFilePath)
		output []byte
		err    error
	)

	if dbConnInfo, soteErr = initConnInfo(targetEnvironment, config.customDb); soteErr.ErrCode == nil {
		initGob()

		fileInfo := FileInfo{
			Format: exportFormat,
			Env:    targetEnvironment,
			Schema: config.Schema,
		}

		for _, tableName := range config.Tables {
			tRows, err := dbConnInfo.DBPoolPtr.Query(dbConnInfo.DBContext, fmt.Sprintf("SELECT * FROM %v.%v", config.Schema, tableName))
			if err != nil {
				sLogger.Info(fmt.Sprintf("Table: %v.%v, Error: %v", config.Schema, tableName, err))
				panic(err.Error())
			}
			sLogger.Info(fmt.Sprintf("Table: %v.%v", config.Schema, tableName))
			table := Table{Name: tableName}
			for _, f := range tRows.FieldDescriptions() {
				table.Columns = append(table.Columns, string(f.Name))
			}

			uniqIndexes := map[string]map[string]int{}
			if doObscure && config.Obscure[tableName] != nil {
				tRows, err := dbConnInfo.DBPoolPtr.Query(dbConnInfo.DBContext, fmt.Sprintf("SELECT indexname, indexdef FROM pg_indexes WHERE schemaname='%v' AND tablename='%v'", config.Schema, tableName))
				if err == nil {
					for tRows.Next() {
						row, err := tRows.Values()
						query := fmt.Sprint(row[1])
						if err == nil && strings.HasPrefix(query, "CREATE UNIQUE INDEX") {
							rule := strings.Split(query, "(")[1]
							colNames := strings.Split(rule[:len(rule)-1], ", ")
							indexes := map[string]int{}
							for _, name := range colNames {
								indexes[name] = -1
							}
							for index, name := range table.Columns {
								if indexes[name] == -1 {
									indexes[name] = index
								}
							}
							uniqIndexes[fmt.Sprintf("%v", row[0])] = indexes
						}
					}
				}
			}

			valueTableMap := map[string]bool{}
			for tRows.Next() {
				row, err := tRows.Values()
				if err != nil {
					panic(err.Error())
				}

				retry := 0
				maxRetry := 10
			GenerateObscureValue:
				retry++
				valueRowMap := map[string]bool{}
				if doObscure && config.Obscure[tableName] != nil {
					for index, name := range table.Columns {
						pattern := config.Obscure[tableName][name]
						if pattern != "" {
							row[index] = generateNewValue(pattern)
						}
					}
				}

				null := []byte("null")
				for _, indexes := range uniqIndexes {
					val := ""
					for colName, index := range indexes {
						b, _ := json.Marshal(row[index])
						if row[index] == nil || bytes.Equal(b, null) {
							val = ""
							break //skip validation if some of the value is null
						}
						val += fmt.Sprintf("%s=%s_", colName, b)
					}
					if val != "" {
						if valueTableMap[val] && retry < maxRetry {
							sLogger.Debug(fmt.Sprintf("%v.%v (%v) duplicate value: %v", config.Schema, tableName, retry, val))
							goto GenerateObscureValue
						}
						valueRowMap[val] = true
					}
				}

				for val, _ := range valueRowMap {
					valueTableMap[val] = true
				}

				if retry < maxRetry {
					table.Rows = append(table.Rows, row)
				} else {
					sLogger.Debug(fmt.Sprintf("%v.%v skipped row: %+v", config.Schema, tableName, row))
				}
			}
			tRows.Close()
			fileInfo.Tables = append(fileInfo.Tables, table)
		}
		dbConnInfo.DBPoolPtr.Close()

		if exportFormat == "csv" {
			outFile, err := os.Create(config.Name + csvFileZipExtension)
			if err != nil {
				panic(err)
			}
			defer outFile.Close()
			w := zip.NewWriter(outFile)
			f, err := w.Create("info.json")
			if err != nil {
				panic(err)
			}
			_, err = f.Write([]byte(fmt.Sprintf("{\"format\": \"csv\", \"environment\": \"%v\", \"schema\": \"%v\"}", fileInfo.Env, fileInfo.Schema)))
			if err != nil {
				panic(err)
			}
			for _, table := range fileInfo.Tables {
				f, err := w.Create(table.Name + ".csv")
				if err != nil {
					panic(err)
				}
				var buf bytes.Buffer
				buf.WriteString(strings.Join(table.Columns, ",") + "\n")
				for _, row := range table.Rows {
					line, err := json.Marshal(row)
					if err != nil {
						panic(err)
					}
					buf.Write(line[:len(line)-1][1:])
					buf.WriteByte('\n')
				}
				_, err = f.Write(buf.Bytes())
				if err != nil {
					panic(err)
				}
			}
			err = w.Close()
			if err != nil {
				panic(err)
			}
			return
		} else if exportFormat == "raw" {
			b := new(bytes.Buffer)
			e := gob.NewEncoder(b)
			err := e.Encode(fileInfo)
			if err != nil {
				panic(err)
			}
			output = b.Bytes()
		} else {
			output, err = json.MarshalIndent(fileInfo, "", " ")
			if err != nil {
				panic(err)
			}
		}
		ioutil.WriteFile(config.Name+"."+exportFormat, output, 0644)
	}
	return
}

// Import database records
func DbImport(targetEnvironment, configFilePath, bulkFilePath string, doClean bool) (soteErr sError.SoteError) {
	sLogger.DebugMethod()
	sLogger.Info("Importing...")

	var fileInfo FileInfo

	if strings.HasSuffix(bulkFilePath, csvFileZipExtension) {
		zipReader, err := zip.OpenReader(bulkFilePath)
		if err != nil {
			panic(err)
		}
		defer zipReader.Close()
		for _, zipItem := range zipReader.File {
			zipItemReader, err := zipItem.Open()
			if err != nil {
				panic(err)
			}
			defer zipItemReader.Close()

			buf, err := ioutil.ReadAll(zipItemReader)
			if err != nil {
				panic(err)
			}

			name_ext := strings.Split(zipItem.Name, ".")
			if name_ext[1] == "json" {
				json.Unmarshal(buf, &fileInfo)
			} else {
				lines := strings.Split(string(buf), "\n")
				columns := strings.Split(lines[0], ",")
				var rows [][]interface{}
				for i := 1; i < len(lines); i++ {
					var row []interface{}
					_ = json.Unmarshal([]byte("["+lines[i]+"]"), &row)
					if len(row) == len(columns) {
						rows = append(rows, row)
					}
				}
				fileInfo.Tables = append(fileInfo.Tables, Table{
					Name:    name_ext[0],
					Columns: columns,
					Rows:    rows,
				})
			}
		}
	} else {
		file, err := os.Open(bulkFilePath)
		if err != nil {
			panic(err)
		}

		data, err := ioutil.ReadAll(file)
		if err != nil {
			panic(err)
		}

		if strings.HasSuffix(bulkFilePath, ".json") {
			err = json.Unmarshal(data, &fileInfo)
		} else { //raw
			initGob()
			b := bytes.NewBuffer(data)
			d := gob.NewDecoder(b)
			err = d.Decode(&fileInfo)
		}
		if err != nil {
			panic(err)
		}
	}
	return fileInfoToDb(fileInfo, targetEnvironment, configFilePath, doClean)
}

func fileInfoToDb(fileInfo FileInfo, targetEnvironment, configFilePath string, doClean bool) (soteErr sError.SoteError) {
	var (
		custom *ConfigCustomDb
		schema = fileInfo.Schema
	)

	if isCustom(targetEnvironment) {
		config := loadConfig(targetEnvironment, configFilePath)
		custom = config.customDb
		schema = config.Schema
	}

	if dbConnInfo, soteErr = initConnInfo(targetEnvironment, custom); soteErr.ErrCode == nil {
		for _, table := range fileInfo.Tables {
			if doClean {
				dbConnInfo.DBPoolPtr.Exec(dbConnInfo.DBContext, fmt.Sprintf("DELETE FROM %v.%v", schema, table.Name))
			}

			if fileInfo.Format == "raw" {
				insertCount, err := dbConnInfo.DBPoolPtr.CopyFrom(dbConnInfo.DBContext, pgx.Identifier{schema, table.Name}, table.Columns, pgx.CopyFromRows(table.Rows))
				if err != nil {
					sLogger.Info(fmt.Sprintf("Table: %v.%v, Error: %v", schema, table.Name, err))
				} else {
					sLogger.Info(fmt.Sprintf("Table: %v.%v, inserted rows: %v", schema, table.Name, insertCount))
				}
			} else {
				rows := make([]interface{}, len(table.Rows))
				for r, row := range table.Rows {
					coldata := map[string]interface{}{}
					for c, name := range table.Columns {
						coldata[name] = row[c]
					}
					rows[r] = coldata
				}
				output, err := json.MarshalIndent(rows, "", " ")
				if err != nil {
					sLogger.Info(fmt.Sprintf("Table: %v.%v, Error: %v", schema, table.Name, err))
				} else {
					result, err := dbConnInfo.DBPoolPtr.Exec(dbConnInfo.DBContext,
						fmt.Sprintf("INSERT INTO %[1]s.%[2]s SELECT * FROM json_populate_recordset (NULL::%[1]s.%[2]s, $$%[3]s$$)", schema, table.Name, output))
					if err != nil {
						sLogger.Info(fmt.Sprintf("Table: %v.%v, Error: %v", schema, table.Name, err))
					} else {
						insertCount := result.RowsAffected()
						sLogger.Info(fmt.Sprintf("Table: %v.%v, inserted rows: %v", schema, table.Name, insertCount))

						var id string
						row := dbConnInfo.DBPoolPtr.QueryRow(dbConnInfo.DBContext,
							fmt.Sprintf("SELECT column_name FROM information_schema.columns WHERE column_default like '%s' AND table_schema='%s' AND table_name='%s'", "%nextval%", schema, table.Name))
						err := row.Scan(&id)
						if err != pgx.ErrNoRows {
							if err != nil {
								sLogger.Debug(fmt.Sprintf("Cannot get information Table: %v.%v, Error: %v", schema, table.Name, err))
							} else {
								var seqNum *int64
								row := dbConnInfo.DBPoolPtr.QueryRow(dbConnInfo.DBContext,
									fmt.Sprintf("SELECT pg_catalog.setval(pg_get_serial_sequence('%[1]s.%[2]s', '%[3]s'), MAX(%[3]s)) FROM %[1]s.%[2]s", schema, table.Name, id))
								err := row.Scan(&seqNum)
								if err != nil && err != pgx.ErrNoRows {
									sLogger.Debug(fmt.Sprintf("Cannot reset a serial sequence Table: %v.%v, Error: %v", schema, table.Name, err))
								} else if seqNum != nil {
									sLogger.Debug(fmt.Sprintf("New serial sequence %v.%v, is: %v", schema, table.Name, *seqNum))
									/* Verify
									SELECT MAX(%PK_COLUMN_NAME%) FROM %SCHEMA%.%TABLE%;
									SELECT pg_get_serial_sequence('%SCHEMA%.%TABLE%', '%PK_COLUMN_NAME%');
									SELECT last_value FROM %SEQ_ID%;
									*/
								}
							}
						}
					}
				}
			}
		}
		dbConnInfo.DBPoolPtr.Close()
	}
	return
}

// Clean database table(s)
func DbClean(targetEnvironment, configFilePath string) (soteErr sError.SoteError) {
	sLogger.DebugMethod()
	sLogger.Info("Cleaning...")

	config := loadConfig(targetEnvironment, configFilePath)

	if dbConnInfo, soteErr = initConnInfo(targetEnvironment, config.customDb); soteErr.ErrCode == nil {
		for _, tableName := range config.Tables {
			_, err := dbConnInfo.DBPoolPtr.Exec(dbConnInfo.DBContext, fmt.Sprintf("DELETE FROM %v.%v", config.Schema, tableName))
			if err != nil {
				sLogger.Info(fmt.Sprintf("Table: %v.%v, Error: %v", config.Schema, tableName, err))
				panic(err.Error())
			}
			sLogger.Info(fmt.Sprintf("Table: %v.%v", config.Schema, tableName))
		}
		dbConnInfo.DBPoolPtr.Close()
	}
	return
}

func generateNewValue(pattern string) *string {
	re := regexp.MustCompile(`\[([^{}]*)\]{([^{}]*)}`)
	matches := re.FindAllStringIndex(pattern, -1)
	partObscure := ""
	lastIndex := 0
	for _, indexes := range matches {
		random := rand.New(rand.NewSource(time.Now().UnixNano()))
		_substr := string(pattern[indexes[0]:indexes[1]])
		for _, match := range re.FindAllStringSubmatch(_substr, -1) {
			_range := strings.Split(match[2], ",")
			endIndex := -1
			if len(_range) == 2 {
				endIndex, _ = strconv.Atoi(_range[1])
			}
			if begIndex, err := strconv.Atoi(_range[0]); err == nil {
				if begIndex == 0 && random.Intn(2) == 0 { //random continue or not
					if partObscure == "" {
						return nil
					}
					return &partObscure
				}
				str := ""
				if strings.ToLower(match[1]) == "a-z" { //strings
					if endIndex == -1 {
						str = fakelish.GenerateFakeWordByLength(begIndex)
					} else {
						str = fakelish.GenerateFakeWord(begIndex, endIndex)
					}
					if match[1] == "a-z" {
						str = strings.ToLower(str)
					} else if match[1] == "A-Z" {
						str = strings.ToUpper(str)
					} else {
						str = strings.Title(str)
					}
				} else if match[1] == "0-9" { //numbers
					nums := random.Perm(begIndex)
					if endIndex != -1 {
						nums = random.Perm(random.Intn(endIndex-begIndex) + begIndex)
					}
					var buf bytes.Buffer
					for i := range nums {
						if nums[i] == 0 {
							nums[i] = random.Intn(8) + 1
						}
						buf.WriteString(fmt.Sprintf("%d", nums[i]))
					}
					time.Sleep(1 * time.Nanosecond) //wait next time seq. (avoid duplicate in generation)
					str = buf.String()
				} else {
					str = match[1]
					if endIndex != -1 {
						for i := 1; i < endIndex; i++ {
							str += match[1]
						}
					}
				}
				partObscure += pattern[lastIndex:indexes[0]] + str
				lastIndex = indexes[1]
			}
		}
	}
	length := len(pattern)
	partObscure += pattern[lastIndex:length]
	partObscure = strings.TrimSpace(partObscure)
	return &partObscure
}
