//go:build !vanilla
// +build !vanilla

package mains

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/yuin/gopher-lua"

	"github.com/nyaosorg/go-box/v2"
	"github.com/nyaosorg/go-readline-ny"
	"github.com/nyaosorg/go-readline-ny/nameutils"

	"github.com/nyaosorg/nyagos/internal/texts"
)

type _KeyLuaFunc struct {
	Chank *lua.LFunction
}

func getBufferForCallBack(L Lua) (*readline.Buffer, int) {
	table, ok := L.Get(1).(*lua.LTable)
	if !ok {
		return nil, lerror(L, "bindKeyExec: call with : not .")
	}
	userdata, ok := L.GetField(table, "buffer").(*lua.LUserData)
	if !ok {
		return nil, lerror(L, "bindKey.Call: invalid object")
	}
	buffer, ok := userdata.Value.(*readline.Buffer)
	if !ok {
		return nil, lerror(L, "can not find readline.Buffer")
	}
	return buffer, 0
}

func callReplace(L Lua) int {
	buffer, stackRc := getBufferForCallBack(L)
	if buffer == nil {
		return stackRc
	}
	pos, ok := L.Get(-2).(lua.LNumber)
	if !ok {
		return lerror(L, "not a number")
	}
	if pos <= 0 {
		return lerror(L, fmt.Sprintf(":replace: pos=%d: Too small.", pos))
	}
	str := L.ToString(-1)
	posZeroBase := int(pos) - 1
	if posZeroBase > len(buffer.Buffer) {
		return lerror(L, fmt.Sprintf(":replace: pos=%d: Too big.", pos))
	}
	buffer.ReplaceAndRepaint(posZeroBase, string(str))
	L.Push(lua.LTrue)
	L.Push(lua.LNil)
	return 2
}

func callInsert(L Lua) int {
	buffer, stackRc := getBufferForCallBack(L)
	if buffer == nil {
		return stackRc
	}
	text := L.ToString(2)
	buffer.InsertAndRepaint(string(text))
	L.Push(lua.LTrue)
	return 1
}

func callKeyFunc(L Lua) int {
	buffer, stackRc := getBufferForCallBack(L)
	if buffer == nil {
		return stackRc
	}
	key := L.ToString(2)
	function, err := nameutils.GetFunc(key)
	if err != nil {
		return lerror(L, err.Error())
	}
	ctx := context.Background()
	switch function.Call(ctx, buffer) {
	case readline.ENTER:
		L.Push(lua.LTrue)
		L.Push(lua.LTrue)
		return 2
	case readline.ABORT:
		L.Push(lua.LTrue)
		L.Push(lua.LFalse)
		return 2
	default:
		L.Push(lua.LNil)
		return 1
	}
}

func callLastWord(L Lua) int {
	this, stackCount := getBufferForCallBack(L)
	if this == nil {
		return stackCount
	}
	word, pos := this.CurrentWord()
	L.Push(lua.LString(word))
	L.Push(lua.LNumber(pos + 1))
	return 2
}

func callFirstWord(L Lua) int {
	this, stackCount := getBufferForCallBack(L)
	if this == nil {
		return stackCount
	}
	word := texts.FirstWord(this.String())
	L.Push(lua.LString(word))
	L.Push(lua.LNumber(0))
	return 2
}

func callBoxListing(L Lua) int {
	// stack +1: readline.Buffer
	// stack +2: table
	// stack +3: index or value
	this, stackCount := getBufferForCallBack(L)
	if this == nil {
		return stackCount
	}
	fmt.Print("\n")
	table := L.ToTable(2)
	size := table.Len()
	list := make([]string, size)
	for i := 0; i < size; i++ {
		list[i] = L.GetTable(table, lua.LNumber(i+1)).String()
	}
	box.Print(context.TODO(), list, os.Stdout)
	this.RepaintAll()
	return 0
}

func (f _KeyLuaFunc) String() string {
	return f.Chank.String()
}
func (f *_KeyLuaFunc) Call(ctx context.Context, buffer *readline.Buffer) readline.Result {
	L, ok := ctx.Value(luaKey).(Lua)
	if !ok {
		println("(*mains._KeyLuaFunc)Call: lua instance not found")
		return readline.CONTINUE
	}
	L.Push(f.Chank)
	pos := -1
	var text strings.Builder
	for i, c := range buffer.Buffer {
		if i == buffer.Cursor {
			pos = text.Len() + 1
		}
		c.Moji.WriteTo(&text)
	}
	if pos < 0 {
		pos = text.Len() + 1
	}

	table := L.NewTable()
	L.SetField(table, "pos", lua.LNumber(pos))
	L.SetField(table, "text", lua.LString(text.String()))
	userdata := L.NewUserData()
	userdata.Value = buffer
	L.SetField(table, "buffer", userdata)
	L.SetField(table, "call", L.NewFunction(callKeyFunc))
	L.SetField(table, "insert", L.NewFunction(callInsert))
	L.SetField(table, "replacefrom", L.NewFunction(callReplace))
	L.SetField(table, "lastword", L.NewFunction(callLastWord))
	L.SetField(table, "firstword", L.NewFunction(callFirstWord))
	L.SetField(table, "boxprint", L.NewFunction(callBoxListing))

	defer setContext(getContext(L), L)
	setContext(ctx, L)

	L.Push(table)
	err := L.PCall(1, 1, nil)
	if err != nil {
		println(err.Error())
	} else {
		switch value := L.Get(-1).(type) {
		case lua.LString:
			buffer.InsertAndRepaint(string(value))
		case lua.LBool:
			if !value {
				buffer.Buffer = buffer.Buffer[:0]
			}
			return readline.ENTER
		}
	}
	return readline.CONTINUE
}

func cmdBindKey(L Lua) int {
	keyTmp, ok := L.Get(-2).(lua.LString)
	if !ok {
		return lerror(L, "bindkey: key error")
	}
	key := strings.Replace(strings.ToUpper(string(keyTmp)), "-", "_", -1)
	switch value := L.Get(-1).(type) {
	case *lua.LFunction:
		if err := nameutils.BindKeyFunc(readline.GlobalKeyMap, key, &_KeyLuaFunc{value}); err != nil {
			return lerror(L, err.Error())
		}
		L.Push(lua.LTrue)
		return 1
	default:
		val := L.ToString(-1)
		err := nameutils.BindKeySymbol(readline.GlobalKeyMap, key, val)
		if err != nil {
			return lerror(L, err.Error())
		}
		L.Push(lua.LTrue)
		return 1
	}
}
