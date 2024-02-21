package functions

var Table = map[string]func([]interface{}) []interface{}{
	"access":             CmdAccess,
	"atou":               CmdAtoU,
	"atou_if_needed":     CmdAnsiToUtf8IfNeeded,
	"bitand":             CmdBitAnd,
	"bitor":              CmdBitOr,
	"chdir":              CmdChdir,
	"commonprefix":       CmdCommonPrefix,
	"complete_for_files": CmdCompleteForFiles,
	"dirname":            CmdDirName,
	"elevated":           CmdElevated,
	"envadd":             CmdEnvAdd,
	"envdel":             CmdEnvDel,
	"fields":             CmdFields,
	"getenv":             CmdGetEnv,
	"gethistory":         CmdGetHistory,
	"gethistorydetail":   CmdGetHistoryDetail,
	"getkey":             CmdGetKey,
	"getviewwidth":       CmdGetViewWidth,
	"getwd":              CmdGetwd,
	"glob":               CmdGlob,
	"msgbox":             CmdMsgBox,
	"pathjoin":           CmdPathJoin,
	"pushhistory":        CmdPushHistory,
	"resetcharwidth":     CmdResetCharWidth,
	"setenv":             CmdSetEnv,
	"setrunewidth":       CmdSetRuneWidth,
	"shellexecute":       CmdShellExecute,
	"skk":                CmdSkk,
	"stat":               CmdStat,
	"utoa":               CmdUtoA,
	"which":              CmdWhich,
}

var Table2 = map[string]func(*Param) []interface{}{
	"box":            CmdBox,
	"raweval":        CmdRawEval,
	"rawexec":        CmdRawExec,
	"write":          CmdWrite,
	"writerr":        CmdWriteErr,
	"default_prompt": Prompt,
}
