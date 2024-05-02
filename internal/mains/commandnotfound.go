//go:build !vanilla
// +build !vanilla

package mains

import (
	"context"

	"github.com/nyaosorg/nyagos/internal/shell"
	"github.com/yuin/gopher-lua"
)

var orgOnCommandNotFound func(context.Context, *shell.Cmd, error) error

func (L *_LuaCallBack) onCommandNotFound(ctx context.Context, sh *shell.Cmd, err error) error {
	nyagosTbl := L.GetGlobal("nyagos")
	hook, ok := L.GetField(nyagosTbl, "on_command_not_found").(*lua.LFunction)
	if !ok {
		return orgOnCommandNotFound(ctx, sh, err)
	}
	args := L.NewTable()
	for key, val := range sh.Args() {
		L.SetTable(args, lua.LNumber(key), lua.LString(val))
	}
	L.Push(hook)
	L.Push(args)
	err1 := execLuaKeepContextAndShell(ctx, &sh.Shell, L.Lua, 1, 1)
	if err1 != nil {
		return err1
	}
	result := L.Get(-1)
	if result == lua.LTrue {
		return nil
	}
	return orgOnCommandNotFound(ctx, sh, err)
}
