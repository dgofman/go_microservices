/*
	This is a wrapper for Sote Golang developers to access services from YouTrack.
*/
package packages

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"gitlab.com/soteapps/packages/v2021/sError"
	"gitlab.com/soteapps/packages/v2021/sLogger"
)

type IData interface {
	GetType() interface{}
	GetList() []interface{}
}

type Data struct {
	Arr []interface{}
}

func (d *Data) GetType() interface{} {
	return &d.Arr
}

func (d *Data) GetList() []interface{} {
	return d.Arr
}

type ProjectTeams struct {
	List []interface{} `json:"projectteams"`
}

func (p *ProjectTeams) GetType() interface{} {
	return &p
}

func (p *ProjectTeams) GetList() []interface{} {
	return p.List
}

type RefId struct {
	Id string `json:"id"`
}

type ProjectIssues struct {
	Issues []interface{} `json:"issues"`
}

type Issue struct {
	Id                  string        `json:"id"`
	Created             int64         `json:"created"`
	Updated             int64         `json:"updated"`
	Resolved            int64         `json:"resolved"`
	NumberInProject     int64         `json:"numberInProject"`
	IdReadable          string        `json:"idReadable"`
	Summary             string        `json:"summary"`
	Description         string        `json:"description"`
	WikifiedDescription string        `json:"wikifiedDescription"`
	UsesMarkdown        bool          `json:"usesMarkdown"`
	IsDraft             bool          `json:"isDraft"`
	Votes               int           `json:"votes"`
	CommentsCount       int           `json:"commentsCount"`
	Comments            []interface{} `json:"comments"`
	Tags                []interface{} `json:"tags"`
	Attachments         []interface{} `json:"attachments"`
	Project             interface{}   `json:"project"`
	Reporter            interface{}   `json:"reporter"`
	Updater             interface{}   `json:"updater"`
	DraftOwner          interface{}   `json:"draftOwner"`
	ExternalIssue       interface{}   `json:"externalIssue"`
	Voters              interface{}   `json:"voters"`
	Watchers            interface{}   `json:"watchers"`
	Subtasks            interface{}   `json:"subtasks"`
	Parent              interface{}   `json:"parent"`
}

const MAX_LENGTH = 1000
const BASED_DIR = "data/"
const ARCHIVE_DIR = BASED_DIR + "archive/"
const MD5_FILE = ARCHIVE_DIR + "md5_history.log"

var outputDir string
var md5History map[string]string

func ExportJson(ytmPtr *YouTrackManager, withAttachments bool) {
	sLogger.DebugMethod()

	var file *os.File
	var err error

	md5History = make(map[string]string)
	if _, err = os.Stat(ARCHIVE_DIR); os.IsNotExist(err) {
		err := os.MkdirAll(ARCHIVE_DIR, os.ModeDir)
		if err != nil {
			panic(err)
		}
	}

	if file, err = os.OpenFile(MD5_FILE, os.O_RDWR, 0666); err == nil {
		b, _ := ioutil.ReadAll(file)
		json.Unmarshal(b, &md5History)
	} else if file, err = os.Create(MD5_FILE); err != nil {
		panic(err)
	}
	defer file.Close()

	t := time.Now()
	outputDir = BASED_DIR + "/" + fmt.Sprintf("%d-%02d-%02dT%02d_%02d_%02d/",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())

	ExportJsonUsers(ytmPtr)
	ExportJsonProjects(ytmPtr)
	ExportJsonGroups(ytmPtr)
	ExportJsonProjectTeams(ytmPtr)
	ExportJsonAgiles(ytmPtr)
	ExportJsonIssues(ytmPtr)
	ExportJsonWorkflows(ytmPtr)
	ExportJsonActivities(ytmPtr)
	if withAttachments {
		ExportJsonAttachments(ytmPtr)
	}

	//start from beginning
	file.Truncate(0)
	file.Seek(0, 0)
	//format json
	enc := json.NewEncoder(file)
	enc.SetIndent("", "    ")
	if err := enc.Encode(md5History); err != nil {
		panic(err)
	}
}

func saveResults(fileName string, data IData, exec func(int, int) (string, sError.SoteError)) {
	offset := 0

	for {
		if _, err := os.Stat(ARCHIVE_DIR + fmt.Sprintf("%s_%d-%d.json", fileName, offset, MAX_LENGTH)); err == nil {
			offset += MAX_LENGTH
			continue
		}
		/*matches, _ := filepath.Glob(outputDir + fmt.Sprintf("%s_%d-*.json", fileName, offset))
		for _, file := range matches {
			os.Remove(file)
		}*/
		sLogger.Info(fmt.Sprintf("FileName: %s, Offset: %d", fileName, offset))
		results, soteErr := exec(MAX_LENGTH, offset)
		if soteErr.ErrCode != nil {
			sLogger.Debug(soteErr.FmtErrMsg)
			break
		}
		arr := data.GetType()
		err := json.Unmarshal([]byte(results), arr)

		if err != nil {
			sLogger.Debug(err.Error())
			break
		}
		length := len(data.GetList())
		if length > 0 {
			toFile(fmt.Sprintf("%s_%d-%d.json", fileName, offset, length), data.GetList(), length)
		}
		offset += MAX_LENGTH
		if length < MAX_LENGTH {
			break
		}
	}
}

func toFile(fileName string, results interface{}, length int) {
	b := new(bytes.Buffer)
	enc := json.NewEncoder(b)
	enc.SetIndent("", "    ")
	if err := enc.Encode(results); err != nil {
		panic(err)
	}
	content := b.Bytes()
	md5Hex := fmt.Sprintf("%x", md5.Sum(content))

	if md5History[fileName] == md5Hex {
		return
	}
	md5History[fileName] = md5Hex

	dir := ARCHIVE_DIR
	if length < MAX_LENGTH {
		dir = outputDir
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, os.ModeDir)
		if err != nil {
			panic(err)
		}
	}

	file, err := os.Create(dir + fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	file.Write(content)
}

func ExportJsonUsers(ytmPtr *YouTrackManager) {
	sLogger.DebugMethod()
	saveResults("users", &Data{}, func(max int, offset int) (string, sError.SoteError) {
		return ytmPtr.ListUsersPagination(max, offset)
	})
}

func ExportJsonProjects(ytmPtr *YouTrackManager) {
	sLogger.DebugMethod()
	saveResults("projects", &Data{}, func(max int, offset int) (string, sError.SoteError) {
		return ytmPtr.ListProjectsPagination(max, offset)
	})
}

func ExportJsonActivities(ytmPtr *YouTrackManager) {
	sLogger.DebugMethod()
	saveResults("activities", &Data{}, func(max int, offset int) (string, sError.SoteError) {
		return ytmPtr.ListActivitiesPagination(max, offset)
	})
}

func ExportJsonGroups(ytmPtr *YouTrackManager) {
	sLogger.DebugMethod()
	saveResults("groups", &Data{}, func(max int, offset int) (string, sError.SoteError) {
		return ytmPtr.ListGroupsPagination(max, offset)
	})
}

func ExportJsonAgiles(ytmPtr *YouTrackManager) {
	sLogger.DebugMethod()
	saveResults("agiles", &Data{}, func(max int, offset int) (string, sError.SoteError) {
		return ytmPtr.ListAgilesPagination(max, offset)
	})
}

func ExportJsonIssues(ytmPtr *YouTrackManager) {
	sLogger.DebugMethod()

	saveResults("issues", &Data{}, func(max int, offset int) (string, sError.SoteError) {
		return ytmPtr.ListIssuesPagination(max, offset)
	})
}

func ExportJsonWorkflows(ytmPtr *YouTrackManager) {
	sLogger.DebugMethod()

	saveResults("workflows", &Data{}, func(max int, offset int) (string, sError.SoteError) {
		return ytmPtr.ListWorkflowsPagination(max, offset)
	})
}

func ExportJsonProjectTeams(ytmPtr *YouTrackManager) {
	sLogger.DebugMethod()

	saveResults("teams", &ProjectTeams{}, func(max int, offset int) (string, sError.SoteError) {
		return ytmPtr.ListIssuesPagination(max, offset)
	})
}

func ExportJsonAttachments(ytmPtr *YouTrackManager) {
	sLogger.DebugMethod()

	results, soteErr := ytmPtr.ListProjects("id,issues(id,attachments(id))")
	if soteErr.ErrCode == nil {
		var arr []ProjectIssues
		err := json.Unmarshal([]byte(results), &arr)
		if err != nil {
			panic(err)
		}
		output := make([]interface{}, 0)
		for _, project := range arr {
			for _, issue := range project.Issues {
				issue, err := json.Marshal(&issue)
				if err != nil {
					panic(err)
				}
				var refIssue Issue
				err = json.Unmarshal(issue, &refIssue)
				if err != nil {
					panic(err)
				}
				for _, attachment := range refIssue.Attachments {
					attachment, err := json.Marshal(&attachment)
					if err != nil {
						panic(err)
					}
					var refAttachment RefId
					err = json.Unmarshal(attachment, &refAttachment)
					if err != nil {
						panic(err)
					}
					sLogger.Info("Get Attachment: " + refAttachment.Id)
					results, soteErr = ytmPtr.GetAttachment(refIssue.Id, refAttachment.Id)
					if soteErr.ErrCode != nil {
						panic(soteErr.FmtErrMsg)
					}
					var attRef interface{}
					err = json.Unmarshal([]byte(results), &attRef)
					if err != nil {
						panic(err)
					}
					output = append(output, attRef)
				}
			}
		}
		/*matches, _ := filepath.Glob(outputDir + "attachments_0-*.json")
		for _, file := range matches {
			os.Remove(file)
		}*/
		toFile(fmt.Sprintf("attachments_0-%d.json", len(output)), output, len(output))
	} else {
		sLogger.Debug(soteErr.FmtErrMsg)
	}
}
