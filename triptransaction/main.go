package main

import (
	"fmt"
	"tripfintransadd/packages"

	"gitlab.com/soteapps/packages/v2021/sHelper"
	"gitlab.com/soteapps/packages/v2021/sLogger"
)

func main() {
	env := sHelper.Parameter{
		Version:     "v2021.1.0",
		AppName:     "tripfintransadd",
		Description: `The message creates a financial transaction for a trip.`,
	}.Init()

	sLogger.Info(fmt.Sprintf("ApplicationName: %s, TargetEnvironment: %s, TestMode: %v", env.ApplicationName, env.TargetEnvironment, env.TestMode))

	if soteErr := packages.Run(env); soteErr.ErrCode != nil {
		panic(soteErr.FmtErrMsg)
	}
}
