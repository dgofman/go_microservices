/*
	This is a wrapper for Sote Golang developers to access services from YouTrack.
*/
package packages

import (
	"fmt"

	"gitlab.com/soteapps/packages/v2021/sError"
	"gitlab.com/soteapps/packages/v2021/sHTTPClient"
	"gitlab.com/soteapps/packages/v2021/sLogger"
)

type AgileManager interface {
	New()
	Close()
	ListProjects()
}

type YouTrackManager struct {
	urlBase       string
	pta           string
	httpClientPtr *sHTTPClient.HTTPManager
	route         string
	fields        string
	query         string
}

const MAX_RECORDS = 1000

const API = "/youtrack/api"
const HUB = "/hub/api/rest"

/*
	New will create a Sote YouTrack Manager.

	youTrackURL is the base URL for your YouTrack instance. Examples are https://{Your Cloud instance}.myjetbrains.com for YouTrack Cloud
	or https://{your stand-alone instance}.com for your stand-alone instance. Trailing slash will be added if not found.

	permanentToken is covered here: https://www.jetbrains.com/help/youtrack/standalone/Manage-Permanent-Token.html
*/
func New(youTrackURL, permanentToken string) (YouTrackManagerPtr *YouTrackManager, soteErr sError.SoteError) {
	sLogger.DebugMethod()

	// Initialize the values for YouTrackManager
	YouTrackManagerPtr = &YouTrackManager{
		urlBase:       youTrackURL,
		pta:           permanentToken,
		httpClientPtr: nil,
		route:         "",
		fields:        "",
		query:         "",
	}

	// Validating the input values are supplied.
	if len(youTrackURL) > 0 {
		YouTrackManagerPtr.urlBase = youTrackURL
	} else {
		soteErr = sError.GetSError(200513, sError.BuildParams([]string{"Your base YouTrack URL"}), nil)
	}

	if len(permanentToken) > 0 {
		YouTrackManagerPtr.pta = permanentToken
	} else {
		soteErr = sError.GetSError(200513, sError.BuildParams([]string{"Your Permanent Token Authorization for YouTrack "}), nil)
	}

	if YouTrackManagerPtr.httpClientPtr, soteErr = sHTTPClient.New(youTrackURL, permanentToken); soteErr.ErrCode != nil {
		panic(soteErr.FmtErrMsg)
	}

	return
}

func getFields(fields []string, defaultField string) string {
	if len(fields) > 0 {
		return "?fields=" + fields[0]
	}
	return "?fields=" + defaultField
}

/*
	Close will set the YouTrackManager to properties to nil or empty
*/
func (ytmPtr *YouTrackManager) Close() {
	sLogger.DebugMethod()

	// Clears the values in the YouTrack Manager
	ytmPtr.pta = ""
	ytmPtr.urlBase = ""
}

/*
	This resource provides operations with issue link types.
	Represents the settings of a link type in YouTrack.

	This table describes attributes of the IssueLinkType entity.

	To receive an attribute in the response from server, specify it explicitly in the request parameter fields
*/
func (ytmPtr *YouTrackManager) GetIssueLinkTypes(fields ...string) (results string, soteErr sError.SoteError) {
	sLogger.DebugMethod()

	ytmPtr.route = fmt.Sprintf("%s/issueLinkTypes", API)
	ytmPtr.fields = getFields(fields, "id,name,shortName,leader(login,name,id),description,createdBy(login,name,id),archived,fromEmail,replyToEmail,team,issues,iconUrl,customFields,template")

	if soteErr = ytmPtr.httpClientPtr.Get(ytmPtr.route+ytmPtr.fields, nil, false); soteErr.ErrCode == nil {
		results = string(ytmPtr.httpClientPtr.RetPack.([]byte))
	}

	return
}

/*
	Represents a user in YouTrack properties of a user account, like credentials, or email and so on.
	To receive an attribute in the response from server, specify it explicitly in the request parameter fields
*/
func (ytmPtr *YouTrackManager) ListUsers(fields ...string) (results string, soteErr sError.SoteError) {
	return ytmPtr.ListUsersPagination(MAX_RECORDS, 0, fields...)
}

/*
	Represents a user in YouTrack properties of a user account, like credentials, or email and so on.
	To receive an attribute in the response from server, specify it explicitly in the request parameter fields
*/
func (ytmPtr *YouTrackManager) ListUsersPagination(max int, offset int, fields ...string) (results string, soteErr sError.SoteError) {
	sLogger.DebugMethod()

	ytmPtr.route = fmt.Sprintf("%s/users", API)
	ytmPtr.fields = getFields(fields, "id,login,fullName,email,name,ringId,banned,online,guest,avatarUrl,jabberAccount,online,avatarUrl,banned,profiles(id,general(timezone(presentation,offset),locale(locale,language,community,name)),notifications(notifyOnOwnChanges,jabberNotificationsEnabled,emailNotificationsEnabled,mentionNotificationsEnabled,duplicateClusterNotificationsEnabled,mailboxIntegrationNotificationsEnabled,usePlainTextEmails,autoWatchOnComment,autoWatchOnCreate,autoWatchOnVote,autoWatchOnUpdate))")

	if soteErr = ytmPtr.httpClientPtr.Get(ytmPtr.route+fmt.Sprintf("/%s&$top=%d&$skip=%d", ytmPtr.fields, max, offset), nil, false); soteErr.ErrCode == nil {
		results = string(ytmPtr.httpClientPtr.RetPack.([]byte))
	}

	return
}

/*
	Represents a YouTrack project.
	To receive an attribute in the response from server, specify it explicitly in the request parameter fields
*/
func (ytmPtr *YouTrackManager) ListProjects(fields ...string) (results string, soteErr sError.SoteError) {
	return ytmPtr.ListProjectsPagination(MAX_RECORDS, 0, fields...)
}

/*
	Represents a YouTrack project.
	To receive an attribute in the response from server, specify it explicitly in the request parameter fields
*/
func (ytmPtr *YouTrackManager) ListProjectsPagination(max int, offset int, fields ...string) (results string, soteErr sError.SoteError) {
	sLogger.DebugMethod()

	ytmPtr.route = fmt.Sprintf("%s/admin/projects", API)
	ytmPtr.fields = getFields(fields, "description,defaultSmtp,shortName,organization(id,name),template,fromEmail,scheduledForRemoval,fieldsSorted,archived,isDemo,iconUrl,leader(id),fromPersonal,defaultVisibilityGroup(id,name),relevantVisibilityGroups(id,name),ringId,team(id),replyToEmail,hubResourceId,historicalShortNames,name,id,issues(id),plugins(id,timeTrackingSettings(id,enabled,estimate(field(id,name),id),timeSpent(id,field(id,name),id)))")

	if soteErr = ytmPtr.httpClientPtr.Get(ytmPtr.route+fmt.Sprintf("/%s&$top=%d&$skip=%d", ytmPtr.fields, max, offset), nil, false); soteErr.ErrCode == nil {
		results = string(ytmPtr.httpClientPtr.RetPack.([]byte))
	}

	return
}

/*
	Represents a YouTrack project.
	To receive an attribute in the response from server, specify it explicitly in the request parameter fields
*/
func (ytmPtr *YouTrackManager) GetProject(projectId string, fields ...string) (results string, soteErr sError.SoteError) {
	sLogger.DebugMethod()

	ytmPtr.route = fmt.Sprintf("%s/admin/projects/%s/", API, projectId)
	ytmPtr.fields = getFields(fields, "id,name,shortName,leader(login,name,id),description,createdBy(login,name,id),archived,fromEmail,replyToEmail,team(id),iconUrl,template,issues(id,created,updated,resolved,numberInProject,project(id),summary,description,usesMarkdown,wikifiedDescription,reporter(login,name,id),updater(login,name,id),isDraft,draftOwner(login,name,id),votes,commentsCount,comments(id),tags(id),links(id),externalIssue(id),customFields(id),voters(id),watchers(id),attachments(id),subtasks(id),parent(id),visibility(id),idReadable),customFields(field(id),project(id),canBeEmpty,emptyFieldText,ordinal,isPublic,hasRunningJob)")

	if soteErr = ytmPtr.httpClientPtr.Get(ytmPtr.route+ytmPtr.fields, nil, false); soteErr.ErrCode == nil {
		results = string(ytmPtr.httpClientPtr.RetPack.([]byte))
	}

	return
}

/*
	Represents an issue in YouTrack.
	To receive an attribute in the response from server, specify it explicitly in the request parameter fields
*/
func (ytmPtr *YouTrackManager) ListProjectIssues(projectId string, fields ...string) (results string, soteErr sError.SoteError) {
	return ytmPtr.ListProjectIssuesPagination(projectId, MAX_RECORDS, 0, fields...)
}

/*
	Represents an issue in YouTrack.
	To receive an attribute in the response from server, specify it explicitly in the request parameter fields
*/
func (ytmPtr *YouTrackManager) ListProjectIssuesPagination(projectId string, max int, offset int, fields ...string) (results string, soteErr sError.SoteError) {
	sLogger.DebugMethod()

	ytmPtr.route = fmt.Sprintf("%s/admin/projects/%s/issues", API, projectId)
	ytmPtr.fields = getFields(fields, "id,idReadable,created,updated,resolved,summary,description,reporter,updater,isDraft,comments,tags,links")

	if soteErr = ytmPtr.httpClientPtr.Get(ytmPtr.route+fmt.Sprintf("/%s&$top=%d&$skip=%d", ytmPtr.fields, max, offset), nil, false); soteErr.ErrCode == nil {
		results = string(ytmPtr.httpClientPtr.RetPack.([]byte))
	}

	return
}

/*
	Represents time tracking settings of your server.
	To receive an attribute in the response from server, specify it explicitly in the request parameter fields
*/
func (ytmPtr *YouTrackManager) GetProjectTimeTrackingSettings(projectId string, fields ...string) (results string, soteErr sError.SoteError) {
	sLogger.DebugMethod()

	ytmPtr.route = fmt.Sprintf("%s/admin/projects/%s/timeTrackingSettings", API, projectId)
	ytmPtr.fields = getFields(fields, "id,enabled,estimate,timeSpent,workItemTypes(id,name,autoAttached),project(id)")

	if soteErr = ytmPtr.httpClientPtr.Get(ytmPtr.route+ytmPtr.fields, nil, false); soteErr.ErrCode == nil {
		results = string(ytmPtr.httpClientPtr.RetPack.([]byte))
	}

	return
}

/*
	Represents a file that is attached to an issue or a comment.
	To receive an attribute in the response from server, specify it explicitly in the request parameter fields
*/
func (ytmPtr *YouTrackManager) GetAttachments(issueId string, fields ...string) (results string, soteErr sError.SoteError) {
	sLogger.DebugMethod()

	ytmPtr.route = fmt.Sprintf("%s/issues/%s/attachments", API, issueId)
	ytmPtr.fields = getFields(fields, "id,name,author(id),created,updated,size,mimeType,extension,removed,visibility(id),issue(id),comment(id),draft,charset,metaData,url,thumbnailURL,base64Content")

	if soteErr = ytmPtr.httpClientPtr.Get(ytmPtr.route+ytmPtr.fields, nil, false); soteErr.ErrCode == nil {
		results = string(ytmPtr.httpClientPtr.RetPack.([]byte))
	}

	return
}

/*
	Represents a file that is attached to an issue or a comment.
	To receive an attribute in the response from server, specify it explicitly in the request parameter fields
*/
func (ytmPtr *YouTrackManager) GetAttachment(issueId string, attachmentId string, fields ...string) (results string, soteErr sError.SoteError) {
	sLogger.DebugMethod()

	ytmPtr.route = fmt.Sprintf("%s/issues/%s/attachments/%s", API, issueId, attachmentId)
	ytmPtr.fields = getFields(fields, "id,name,author(id),created,updated,size,mimeType,extension,removed,visibility(id),issue(id),comment(id),draft,charset,metaData,url,thumbnailURL,base64Content")

	if soteErr = ytmPtr.httpClientPtr.Get(ytmPtr.route+ytmPtr.fields, nil, false); soteErr.ErrCode == nil {
		results = string(ytmPtr.httpClientPtr.RetPack.([]byte))
	}

	return
}

/*
	Represents a change in an issue or in its related entities. In the UI, you see these changes as the Activity stream. It shows a feed of all updates of the issue: issue history, comments, attachments, VCS changes, work items, and so on.
	To receive an attribute in the response from server, specify it explicitly in the request parameter fields
*/
func (ytmPtr *YouTrackManager) ListActivities(fields ...string) (results string, soteErr sError.SoteError) {
	return ytmPtr.ListActivitiesPagination(MAX_RECORDS, 0, fields...)
}

/*
	Represents a change in an issue or in its related entities. In the UI, you see these changes as the Activity stream. It shows a feed of all updates of the issue: issue history, comments, attachments, VCS changes, work items, and so on.
	To receive an attribute in the response from server, specify it explicitly in the request parameter fields
*/
func (ytmPtr *YouTrackManager) ListActivitiesPagination(max int, offset int, fields ...string) (results string, soteErr sError.SoteError) {
	sLogger.DebugMethod()

	ytmPtr.route = fmt.Sprintf("%s/activities", API)
	ytmPtr.fields = getFields(fields, "id,author(id),timestamp,added(id,name,created,visibility(id),issue(id),project(id),idreadable),removed(id,name,created,visibility(id),issue(id),project(id)),target(id,idReadable)&categories=*")

	if soteErr = ytmPtr.httpClientPtr.Get(ytmPtr.route+fmt.Sprintf("%s&$top=%d&$skip=%d", ytmPtr.fields, max, offset), nil, false); soteErr.ErrCode == nil {
		results = string(ytmPtr.httpClientPtr.RetPack.([]byte))
	}

	return
}

/*
	Get All Project Teams
*/
func (ytmPtr *YouTrackManager) ListProjectTeams(fields ...string) (results string, soteErr sError.SoteError) {
	return ytmPtr.ListProjectTeamsPagination(MAX_RECORDS, 0, fields...)
}

/*
	Get All Project Teams
*/
func (ytmPtr *YouTrackManager) ListProjectTeamsPagination(max int, offset int, fields ...string) (results string, soteErr sError.SoteError) {
	sLogger.DebugMethod()

	ytmPtr.route = fmt.Sprintf("%s/projectteams", HUB)
	ytmPtr.fields = getFields(fields, "id,name,users(id,login),project(id,owner)")

	if soteErr = ytmPtr.httpClientPtr.Get(ytmPtr.route+fmt.Sprintf("%s&$top=%d&$skip=%d", ytmPtr.fields, max, offset), nil, false); soteErr.ErrCode == nil {
		results = string(ytmPtr.httpClientPtr.RetPack.([]byte))
	}

	return
}

/*
	This resource lets you read the list of user groups and specific user group in YouTrack.
*/
func (ytmPtr *YouTrackManager) ListGroups(fields ...string) (results string, soteErr sError.SoteError) {
	return ytmPtr.ListGroupsPagination(MAX_RECORDS, 0, fields...)
}

/*
	This resource lets you read the list of user groups and specific user group in YouTrack.
*/
func (ytmPtr *YouTrackManager) ListGroupsPagination(max int, offset int, fields ...string) (results string, soteErr sError.SoteError) {
	sLogger.DebugMethod()

	ytmPtr.route = fmt.Sprintf("%s/groups", API)
	ytmPtr.fields = getFields(fields, "id,name,teamForProject(id),allUsersGroup,usersCount,ringId,icon")

	if soteErr = ytmPtr.httpClientPtr.Get(ytmPtr.route+fmt.Sprintf("%s&$top=%d&$skip=%d", ytmPtr.fields, max, offset), nil, false); soteErr.ErrCode == nil {
		results = string(ytmPtr.httpClientPtr.RetPack.([]byte))
	}

	return
}

/*
	Represents an agile board configuration.
	To receive an attribute in the response from server, specify it explicitly in the request parameter fields
*/
func (ytmPtr *YouTrackManager) ListAgiles(fields ...string) (results string, soteErr sError.SoteError) {
	return ytmPtr.ListAgilesPagination(MAX_RECORDS, 0, fields...)
}

/*
	Represents an agile board configuration.
	To receive an attribute in the response from server, specify it explicitly in the request parameter fields
*/
func (ytmPtr *YouTrackManager) ListAgilesPagination(max int, offset int, fields ...string) (results string, soteErr sError.SoteError) {
	sLogger.DebugMethod()

	ytmPtr.route = fmt.Sprintf("%s/agiles", API)
	ytmPtr.fields = getFields(fields, "backlog(availableRules(id),folderId,id,isUpdatable,name,query,updateableBy(id,name),visibleFor(id,name)),cardOnSeveralSprints,cardSettings(fields(field(fieldDefaults(bundle(id,isUpdateable)),fieldType(id,presentation,valueType),id,instances(project(id)),localizedName,name,type),id,presentation(id))),colorCoding(id,projectColors(color(id),id,project(id,name)),prototype(id,localizedName,name)),colorizeCustomFields,columnSettings(columns(collapsed,color($type,id),fieldValues(canUpdate,column(id),id,isResolved,name,ordinal,presentation),id,isResolved,isVisible,ordinal,parent($type,id),wipLimit($type,max,min)),field(fieldDefaults(bundle(id,isUpdateable)),fieldType(id,presentation,valueType),id,instances(project(id)),localizedName,name,type),id,showBundleWarning),currentSprint(id),estimationField($type,fieldType(id,valueType),id,name),extensions(reportSettings(doNotUseBurndown,estimationBurndownField,filterType(id),id,subQuery,yaxis(id))),flatBacklog,hideOrphansSwimlane,id,isDemo,isUpdatable,name,originalEstimationField($type,fieldType(id,valueType),id,name),orphansAtTheTop,owner(id),projects(id),readSharingSettings(permittedGroups(id,name),permittedUsers(id,login,name),projectBased),sprints(archived,finish,id,isDefault,isStarted,name,start),sprintsSettings(addNewIssueToKanban,cardOnSeveralSprints,defaultSprint($type,id,name),disableSprints,explicitQuery,hideSubtasksOfCards,isExplicit,sprintSyncField(fieldType(isMultiValue),id,localizedName,name)),swimlaneSettings($type,defaultCardType(id,name),enabled,field(customField(fieldDefaults(bundle(id,isUpdateable)),fieldType(id,presentation,valueType),id,instances(project(id)),localizedName,name,type),id,instant,multiValue,name,presentation),id,values(id,isResolved,name,presentation,value)),updateSharingSettings(permittedGroups(id,name,teamForProject(id)),permittedUsers(id,login,name),projectBased)")

	if soteErr = ytmPtr.httpClientPtr.Get(ytmPtr.route+fmt.Sprintf("%s&$top=%d&$skip=%d", ytmPtr.fields, max, offset), nil, false); soteErr.ErrCode == nil {
		results = string(ytmPtr.httpClientPtr.RetPack.([]byte))
	}

	return
}

/*
	Represents an issue in YouTrack.
	To receive an attribute in the response from server, specify it explicitly in the request parameter fields
*/
func (ytmPtr *YouTrackManager) ListIssues(fields ...string) (results string, soteErr sError.SoteError) {
	return ytmPtr.ListIssuesPagination(MAX_RECORDS, 0, fields...)
}

/*
	Represents an issue in YouTrack.
	To receive an attribute in the response from server, specify it explicitly in the request parameter fields
*/
func (ytmPtr *YouTrackManager) ListIssuesPagination(max int, offset int, fields ...string) (results string, soteErr sError.SoteError) {
	sLogger.DebugMethod()

	ytmPtr.route = fmt.Sprintf("%s/issues", API)
	ytmPtr.fields = getFields(fields, "created,description,watchers(id,hasStar),visibility(id),project(id),idReadable,usesMarkdown,updated,resolved,summary,votes,reporter(id),mentionedIssues(id),mentionedUsers(id),mentionedArticles(id),tags(id),attachments(id),voters(id,hasVote),hasEmail,wikifiedDescription,updater(id),id,numberInProject,isDraft,draftOwner,externalIssue(id),subtasks(id),parent(id),commentsCount,comments(id,text,usesMarkdown,textPreview,created,updated,author(id),issue(id),attachments(id),deleted)")

	if soteErr = ytmPtr.httpClientPtr.Get(ytmPtr.route+fmt.Sprintf("%s&$top=%d&$skip=%d", ytmPtr.fields, max, offset), nil, false); soteErr.ErrCode == nil {
		results = string(ytmPtr.httpClientPtr.RetPack.([]byte))
	}

	return
}

/*
	In YouTrack, a workflow is a set of rules that can be attached to a project.
	These rules define a lifecycle for issues in a project and automate changes that can be applied to issues.
*/
func (ytmPtr *YouTrackManager) ListWorkflowsPagination(max int, offset int, fields ...string) (results string, soteErr sError.SoteError) {
	sLogger.DebugMethod()

	ytmPtr.route = fmt.Sprintf("%s/admin/workflows", API)
	ytmPtr.fields = getFields(fields, "autoAttach,compatible,converted,id,minimalApiVersion,name,restoreStatus(id),rules(id,name,title,type,script,updated,updatedBy(id,login),workflow(id)),title,updated,updatedBy(id,login),usages(isBroken,project(id,name))")

	if soteErr = ytmPtr.httpClientPtr.Get(ytmPtr.route+fmt.Sprintf("%s&$top=%d&$skip=%d", ytmPtr.fields, max, offset), nil, false); soteErr.ErrCode == nil {
		results = string(ytmPtr.httpClientPtr.RetPack.([]byte))
	}

	return
}
