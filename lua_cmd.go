package main

import "fmt"
import "io"
import "os"
import "os/exec"
import "strings"

import "./alias"
import "./dos"
import "./history"
import "./interpreter"
import "./lua"
import "./mbcs"

import "github.com/shiena/ansicolor"

func cmdAlias(L *lua.Lua) int {
	name, nameErr := L.ToString(1)
	if nameErr != nil {
		L.PushNil()
		L.PushString(nameErr.Error())
		return 2
	}
	key := strings.ToLower(name)
	switch L.GetType(2) {
	case lua.TSTRING:
		value, err := L.ToString(2)
		if err == nil {
			alias.Table[key] = alias.New(value)
		} else {
			L.PushNil()
			L.PushString(err.Error())
			return 2
		}
	case lua.TFUNCTION:
		regkey := "nyagos.alias." + key
		L.SetField(lua.REGISTORYINDEX, regkey)
		alias.Table[key] = LuaFunction{L, regkey}
	}
	L.PushBool(true)
	return 1
}

func cmdSetEnv(L *lua.Lua) int {
	name, nameErr := L.ToString(1)
	if nameErr != nil {
		L.PushNil()
		L.PushString(nameErr.Error())
		return 2
	}
	value, valueErr := L.ToString(2)
	if valueErr != nil {
		L.PushNil()
		L.PushString(valueErr.Error())
		return 2
	}
	os.Setenv(name, value)
	L.PushBool(true)
	return 1
}

func cmdGetEnv(L *lua.Lua) int {
	name, nameErr := L.ToString(1)
	if nameErr != nil {
		L.PushNil()
		return 1
	}
	value := os.Getenv(name)
	if len(value) > 0 {
		L.PushString(value)
	} else {
		L.PushNil()
	}
	return 1
}

func cmdExec(L *lua.Lua) int {
	statement, statementErr := L.ToString(1)
	if statementErr != nil {
		L.PushNil()
		L.PushString(statementErr.Error())
		return 2
	}
	_, err := interpreter.New().Interpret(statement)

	if err != nil {
		L.PushNil()
		L.PushString(err.Error())
		return 2
	}
	L.PushBool(true)
	return 1
}

func cmdEval(L *lua.Lua) int {
	statement, statementErr := L.ToString(1)
	if statementErr != nil {
		L.PushNil()
		L.PushString(statementErr.Error())
		return 2
	}
	r, w, err := os.Pipe()
	if err != nil {
		L.PushNil()
		L.PushString(err.Error())
		return 2
	}
	go func(statement string, w *os.File) {
		it := interpreter.New()
		it.Stdout = w
		it.Interpret(statement)
		w.Close()
	}(statement, w)

	var result = []byte{}
	for {
		buffer := make([]byte, 256)
		size, err := r.Read(buffer)
		if err != nil || size <= 0 {
			break
		}
		result = append(result, buffer[0:size]...)
	}
	r.Close()
	L.PushAnsiString(result)
	return 1
}

func cmdWrite(L *lua.Lua) int {
	var out io.Writer = os.Stdout
	cmd, cmdOk := LuaInstanceToCmd[L.Id()]
	if cmdOk && cmd != nil && cmd.Stdout != nil {
		out = cmd.Stdout
	}
	switch out.(type) {
	case *os.File:
		out = ansicolor.NewAnsiColorWriter(out)
	}

	n := L.GetTop()
	for i := 1; i <= n; i++ {
		str, err := L.ToString(i)
		if err != nil {
			L.PushNil()
			L.PushString(err.Error())
			return 2
		}
		if i > 1 {
			fmt.Fprint(out, "\t")
		}
		fmt.Fprint(out, str)
	}
	L.PushBool(true)
	return 1
}

func cmdGetwd(L *lua.Lua) int {
	wd, err := os.Getwd()
	if err == nil {
		L.PushString(wd)
		return 1
	} else {
		return 0
	}
}

func cmdWhich(L *lua.Lua) int {
	if L.GetType(-1) != lua.TSTRING {
		return 0
	}
	name, nameErr := L.ToString(-1)
	if nameErr != nil {
		L.PushNil()
		L.PushString(nameErr.Error())
		return 2
	}
	path, err := exec.LookPath(name)
	if err == nil {
		L.PushString(path)
		return 1
	} else {
		L.PushNil()
		L.PushString(err.Error())
		return 2
	}
}

func cmdAtoU(L *lua.Lua) int {
	str, err := mbcs.AtoU(L.ToAnsiString(1))
	if err == nil {
		L.PushString(str)
		return 1
	} else {
		return 0
	}
}

func cmdUtoA(L *lua.Lua) int {
	utf8, utf8err := L.ToString(1)
	if utf8err != nil {
		L.PushNil()
		L.PushString(utf8err.Error())
		return 2
	}
	str, err := mbcs.UtoA(utf8)
	if err == nil {
		if len(str) >= 1 {
			L.PushAnsiString(str[:len(str)-1])
		} else {
			L.PushString("")
		}
		L.PushNil()
		return 2
	} else {
		L.PushNil()
		L.PushString(err.Error())
		return 2
	}
}

func cmdGlob(L *lua.Lua) int {
	if !L.IsString(-1) {
		return 0
	}
	wildcard, wildcardErr := L.ToString(-1)
	if wildcardErr != nil {
		L.PushNil()
		L.PushString(wildcardErr.Error())
		return 2
	}
	list, err := dos.Glob(wildcard)
	if err != nil {
		L.PushNil()
		L.PushString(err.Error())
		return 2
	} else {
		L.NewTable()
		for i := 0; i < len(list); i++ {
			L.PushInteger(i + 1)
			L.PushString(list[i])
			L.SetTable(-3)
		}
		return 1
	}
}

func cmdGetHistory(this *lua.Lua) int {
	if this.GetType(-1) == lua.TNUMBER {
		val, err := this.ToInteger(-1)
		if err != nil {
			this.PushNil()
			this.PushString(err.Error())
		}
		this.PushString(history.Get(val))
	} else {
		this.PushInteger(history.Len())
	}
	return 1
}
