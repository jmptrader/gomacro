/*
 * gomacro - A Go intepreter with Lisp-like macros
 *
 * Copyright (C) 2017 Massimiliano Ghilardi
 *
 *     This program is free software: you can redistribute it and/or modify
 *     it under the terms of the GNU General Public License as published by
 *     the Free Software Foundation, either version 3 of the License, or
 *     (at your option) any later version.
 *
 *     This program is distributed in the hope that it will be useful,
 *     but WITHOUT ANY WARRANTY; without even the implied warranty of
 *     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *     GNU General Public License for more details.
 *
 *     You should have received a copy of the GNU General Public License
 *     along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 * builtin.go
 *
 *  Created on: Apr 02, 2017
 *      Author: Massimiliano Ghilardi
 */

package fast

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	r "reflect"
	"time"

	. "github.com/cosmos72/gomacro/base"
)

// =================================== iota ===================================

func (top *Comp) addIota() {
	// https://golang.org/ref/spec#Constants
	// "Literal constants, true, false, iota, and certain constant expressions containing only untyped constant operands are untyped."
	top.Binds["iota"] = BindConst(UntypedZero)
}

func (top *Comp) removeIota() {
	delete(top.Binds, "iota")
}

func (top *Comp) incrementIota() {
	uIota := top.Binds["iota"].Lit.Value.(UntypedLit).Obj
	uIota = constant.BinaryOp(uIota, token.ADD, UntypedOne.Obj)
	top.Binds["iota"] = BindConst(UntypedLit{Kind: r.Int, Obj: uIota})
}

// ============================== initialization ===============================

func (ce *CompEnv) addBuiltins() {
	// https://golang.org/ref/spec#Constants
	// "Literal constants, true, false, iota, and certain constant expressions containing only untyped constant operands are untyped."
	ce.DeclConst("false", nil, UntypedLit{Kind: r.Bool, Obj: constant.MakeBool(false)})
	ce.DeclConst("true", nil, UntypedLit{Kind: r.Bool, Obj: constant.MakeBool(true)})

	// https://golang.org/ref/spec#Variables : "[...] the predeclared identifier nil, which has no type"
	ce.DeclConst("nil", nil, nil)

	// ce.DeclFunc("cap", callCap)
	// ce.DeclFunc("copy", callCopy)
	// ce.DeclFunc("len", callLen)
	ce.DeclFunc("Sleep", func(seconds float64) {
		time.Sleep(time.Duration(seconds * float64(time.Second)))
	})

	ce.DeclBuiltin4("append", compileAppend, 1, MaxInt)
	ce.DeclBuiltin4("cap", compileCap, 1, 1)
	ce.DeclBuiltin4("copy", compileCopy, 2, 2)
	ce.DeclBuiltin4("delete", compileDelete, 2, 2)
	ce.DeclBuiltin4("imag", compileRealImag, 1, 1)
	ce.DeclBuiltin4("len", compileLen, 1, 1)
	ce.DeclBuiltin4("make", compileMake, 1, 3)
	ce.DeclBuiltin4("new", compileNew, 1, 1)
	ce.DeclBuiltin4("real", compileRealImag, 1, 1)

	/*
		binds["Env"] = r.ValueOf(Function{funcEnv, 0})
		binds["Eval"] = r.ValueOf(Function{funcEval, 1})
		binds["MacroExpand"] = r.ValueOf(Function{funcMacroExpand, -1})
		binds["MacroExpand1"] = r.ValueOf(Function{funcMacroExpand1, -1})
		binds["MacroExpandCodewalk"] = r.ValueOf(Function{funcMacroExpandCodewalk, -1})
		binds["Parse"] = r.ValueOf(Function{funcParse, 1})
		binds["Read"] = r.ValueOf(ReadString)
		binds["ReadDir"] = r.ValueOf(callReadDir)
		binds["ReadFile"] = r.ValueOf(callReadFile)
		binds["ReadMultiline"] = r.ValueOf(ReadMultiline)
		binds["Slice"] = r.ValueOf(callSlice)
		binds["String"] = r.ValueOf(func(args ...interface{}) string {
			return env.toString("", args...)
		})
		// return multiple values, extracting the concrete type of each interface
		binds["Values"] = r.ValueOf(Function{funcValues, -1})

		binds["close"] = r.ValueOf(callClose)
		binds["complex"] = r.ValueOf(Function{funcComplex, 2})
	*/
	/*
		binds["panic"] = r.ValueOf(callPanic)
		binds["println"] = r.ValueOf(func(args ...interface{}) {
			// values := toValues(args)
			// env.FprintValues(env.Stdout, values...)
			fmt.Fprintln(env.Stdout, args...)
		})
		binds["recover"] = r.ValueOf(Function{funcRecover, 0})
	*/

	// --------- types ---------
	ce.DeclType("bool", TypeOfBool)
	ce.DeclType("byte", TypeOfByte)
	ce.DeclType("complex64", TypeOfComplex64)
	ce.DeclType("complex128", TypeOfComplex128)
	ce.DeclType("error", TypeOfError)
	ce.DeclType("float32", TypeOfFloat32)
	ce.DeclType("float64", TypeOfFloat64)
	ce.DeclType("int", TypeOfInt)
	ce.DeclType("int8", TypeOfInt8)
	ce.DeclType("int16", TypeOfInt16)
	ce.DeclType("int32", TypeOfInt32)
	ce.DeclType("int64", TypeOfInt64)
	ce.DeclType("rune", TypeOfRune)
	ce.DeclType("string", TypeOfString)
	ce.DeclType("uint", TypeOfUint)
	ce.DeclType("uint8", TypeOfUint8)
	ce.DeclType("uint16", TypeOfUint16)
	ce.DeclType("uint32", TypeOfUint32)
	ce.DeclType("uint64", TypeOfUint64)
	ce.DeclType("uintptr", TypeOfUintptr)

	/*
		// --------- proxies ---------
		if env.Proxies == nil {
			env.Proxies = make(map[string]r.Type)
		}
		proxies := env.Proxies

		proxies["error", TypeOf(*Error_builtin)(nil)).Elem()
	*/

	ce.Apply()
}

// ============================= builtin functions =============================

// --- append() ---

func compileAppend(c *Comp, sym Symbol, node *ast.CallExpr) *Call {
	n := len(node.Args)
	args := make([]*Expr, n)

	args[0] = c.Expr1(node.Args[0])
	t0 := args[0].Type
	if t0.Kind() != r.Slice {
		c.Errorf("first argument to %s must be slice; have <%s>", sym.Name, t0)
		return nil
	}
	telem := t0.Elem()

	for i := 1; i < n; i++ {
		argi := c.Expr1(node.Args[i])
		if argi.Const() {
			argi.ConstTo(telem)
		} else if ti := argi.Type; ti != telem && !ti.AssignableTo(telem) {
			return c.badBuiltinCallArgType(sym.Name, node.Args[i], ti, telem)
		}
		args[i] = argi
	}
	t := r.FuncOf([]r.Type{t0, t0}, []r.Type{t0}, true) // compile as reflect.Append(), which is variadic
	sym.Type = t
	fun := exprLit(Lit{Type: t, Value: r.Append}, &sym)
	return &Call{
		Fun:      fun,
		Args:     args,
		OutTypes: []r.Type{t0},
		Const:    false,
	}
}

// --- cap() ---

func callCap(val r.Value) int {
	return val.Cap()
}

func compileCap(c *Comp, sym Symbol, node *ast.CallExpr) *Call {
	// argument of builtin cap() cannot be a literal
	arg := c.Expr1(node.Args[0])
	tin := arg.Type
	tout := TypeOfInt
	switch tin.Kind() {
	// no cap() on r.Map, see
	// https://golang.org/ref/spec#Length_and_capacity
	// and https://golang.org/pkg/reflect/#Value.Cap
	case r.Array, r.Chan, r.Slice:
		// ok
	case r.Ptr:
		if tin.Elem().Kind() == r.Array {
			// cap() on pointer to array
			arg = c.Deref(arg)
			tin = arg.Type
			break
		}
		fallthrough
	default:
		return c.badBuiltinCallArgType(sym.Name, node.Args[0], tin, "array, channel, slice, pointer to array")
	}
	t := r.FuncOf([]r.Type{tin}, []r.Type{tout}, false)
	sym.Type = t
	fun := exprLit(Lit{Type: t, Value: callCap}, &sym)
	// capacity of arrays is part of their type: cannot change at runtime, we could optimize it.
	// TODO https://golang.org/ref/spec#Length_and_capacity specifies
	// when the array passed to cap() is evaluated and when is not...
	return newCall1(fun, arg, arg.Const(), tout)
}

// --- copy() ---

func copyStringToBytes(dst []byte, src string) int {
	// reflect.Copy does not support this case... use the compiler support
	return copy(dst, src)
}

func compileCopy(c *Comp, sym Symbol, node *ast.CallExpr) *Call {
	args := []*Expr{
		c.Expr1(node.Args[0]),
		c.Expr1(node.Args[1]),
	}
	if args[1].Const() {
		// we also accept a string literal as second argument
		args[1].ConstTo(args[1].DefaultType())
	}
	t0, t1 := args[0].Type, args[1].Type
	var funCopy I = r.Copy
	if t0.Kind() != r.Slice || !t0.AssignableTo(r.SliceOf(t0.Elem())) {
		// https://golang.org/ref/spec#Appending_and_copying_slices
		// copy [...] arguments must have identical element type T and must be assignable to a slice of type []T.
		c.Errorf("first argument to copy should be slice; have %v <%v>", node.Args[0], t0)
		return nil
	} else if t0.Elem().Kind() == r.Uint8 && t1.Kind() == r.String {
		// [...] As a special case, copy also accepts a destination argument assignable to type []byte
		// with a source argument of a string type. This form copies the bytes from the string into the byte slice.
		funCopy = copyStringToBytes
	} else if t1.Kind() != r.Slice || !t1.AssignableTo(r.SliceOf(t1.Elem())) {
		c.Errorf("second argument to copy should be slice or string; have %v <%v>", node.Args[1], t1)
		return nil
	} else if t0.Elem() != t1.Elem() {
		c.Errorf("arguments to copy have different element types: <%v> and <%v>", t0.Elem(), t1.Elem())
	}
	outtypes := []r.Type{t0}
	t := r.FuncOf([]r.Type{t0, t1}, outtypes, false)
	sym.Type = t
	fun := exprLit(Lit{Type: t, Value: funCopy}, &sym)
	return &Call{Fun: fun, Args: args, OutTypes: outtypes, Const: false}
}

// --- delete() ---

// use whatever calling convention is convenient: reflect.Values, interface{}s, primitive types...
// as long as call_builtin supports it, we're fine
func callDelete(vmap r.Value, vkey r.Value) {
	vmap.SetMapIndex(vkey, Nil)
}

func compileDelete(c *Comp, sym Symbol, node *ast.CallExpr) *Call {
	emap := c.Expr1(node.Args[0])
	ekey := c.Expr1(node.Args[1])
	tmap := emap.Type
	if tmap.Kind() != r.Map {
		c.Errorf("first argument to delete must be map; have %v", tmap)
		return nil
	}
	tkey := tmap.Key()
	if ekey.Const() {
		ekey.ConstTo(tkey)
	} else if !ekey.Type.AssignableTo(tkey) {
		c.Errorf("cannot use %v <%v> as type <%v> in delete", node.Args[1], ekey.Type, tkey)
	}
	t := r.FuncOf([]r.Type{tmap, tkey}, ZeroTypes, false)
	sym.Type = t
	fun := exprLit(Lit{Type: t, Value: callDelete}, &sym)
	return &Call{Fun: fun, Args: []*Expr{emap, ekey}, OutTypes: ZeroTypes, Const: false}
}

// --- len() ---

func callLenValue(val r.Value) int {
	return val.Len()
}

func callLenString(val string) int {
	return len(val)
}

func compileLen(c *Comp, sym Symbol, node *ast.CallExpr) *Call {
	arg := c.Expr1(node.Args[0])
	if arg.Const() {
		arg.ConstTo(arg.DefaultType())
	}
	tin := arg.Type
	tout := TypeOfInt
	switch tin.Kind() {
	case r.Array, r.Chan, r.Map, r.Slice, r.String:
		// ok
	case r.Ptr:
		if tin.Elem().Kind() == r.Array {
			// len() on pointer to array
			arg = c.Deref(arg)
			tin = arg.Type
			break
		}
		fallthrough
	default:
		return c.badBuiltinCallArgType(sym.Name, node.Args[0], tin, "array, channel, map, slice, string, pointer to array")
	}
	t := r.FuncOf([]r.Type{tin}, []r.Type{tout}, false)
	sym.Type = t
	fun := exprLit(Lit{Type: t, Value: callLenValue}, &sym)
	if tin.Kind() == r.String {
		fun.Value = callLenString // optimization
	}
	// length of arrays is part of their type: cannot change at runtime, we could optimize it.
	// TODO https://golang.org/ref/spec#Length_and_capacity specifies
	// when the array passed to len() is evaluated and when is not...
	return newCall1(fun, arg, arg.Const(), tout)
}

// --- make() ---

func makeChan1(t r.Type) r.Value {
	return r.MakeChan(t, 0)
}

func makeMap2(t r.Type, n int) r.Value {
	// reflect.MakeMap cannot specify initial capacity
	return r.MakeMap(t)
}

func makeSlice2(t r.Type, n int) r.Value {
	// reflect.MakeSlice requires capacity
	return r.MakeSlice(t, n, n)
}

func compileMake(c *Comp, sym Symbol, node *ast.CallExpr) *Call {
	nargs := len(node.Args)
	nmin, nmax := 1, 2
	tin := c.Type(node.Args[0])
	var funMakes [4]I
	switch tin.Kind() {
	case r.Chan:
		funMakes[1] = makeChan1
		funMakes[2] = r.MakeChan
	case r.Map:
		funMakes[1] = r.MakeMap
		funMakes[2] = makeMap2
	case r.Slice:
		nmin, nmax = 2, 3
		funMakes[2] = makeSlice2
		funMakes[3] = r.MakeSlice
	default:
		return c.badBuiltinCallArgType(sym.Name, node.Args[0], tin, "channel, map, slice")
	}
	if nargs < nmin || nargs > nmax {
		return c.badBuiltinCallArgNum(sym.Name, nmin, nmax, node.Args)
	}
	args := make([]*Expr, nargs)
	argtypes := make([]r.Type, nargs)
	args[0] = exprValue(tin)
	argtypes[0] = TypeOfType
	te := TypeOfInt
	for i := 1; i < nargs; i++ {
		argi := c.Expr1(node.Args[i])
		if argi.Const() {
			argi.ConstTo(te)
		} else if ti := argi.Type; ti != te && !ti.AssignableTo(te) {
			return c.badBuiltinCallArgType(sym.Name, node.Args[i], ti, te)
		}
		args[i] = argi
		argtypes[i] = te
	}
	outtypes := []r.Type{tin}
	t := r.FuncOf(argtypes, outtypes, false)
	sym.Type = t
	funMake := funMakes[nargs]
	if funMake == nil {
		c.Errorf("internal error: no make() alternative to call for %v with %d arguments", tin, nargs)
		return nil
	}
	fun := exprLit(Lit{Type: t, Value: funMake}, &sym)
	return &Call{Fun: fun, Args: args, OutTypes: outtypes, Const: false}
}

// --- new() ---

func compileNew(c *Comp, sym Symbol, node *ast.CallExpr) *Call {
	tin := c.Type(node.Args[0])
	tout := r.PtrTo(tin)
	t := r.FuncOf([]r.Type{TypeOfType}, []r.Type{tout}, false)
	sym.Type = t
	fun := exprLit(Lit{Type: t, Value: r.New}, &sym)
	arg := exprValue(tin)
	return newCall1(fun, arg, false, tout)
}

// --- real() and imag() ---

func callReal32(val complex64) float32 {
	return real(val)
}

func callReal64(val complex128) float64 {
	return real(val)
}

func callImag32(val complex64) float32 {
	return imag(val)
}

func callImag64(val complex128) float64 {
	return imag(val)
}

func compileRealImag(c *Comp, sym Symbol, node *ast.CallExpr) *Call {
	arg := c.Expr1(node.Args[0])
	if arg.Const() {
		arg.ConstTo(arg.DefaultType())
	}
	tin := arg.Type
	var tout r.Type
	var call I
	switch tin.Kind() {
	case r.Complex64:
		tout = TypeOfFloat32
		if sym.Name == "real" {
			call = callReal32
		} else {
			call = callImag32
		}
	case r.Complex128:
		tout = TypeOfFloat64
		if sym.Name == "real" {
			call = callReal64
		} else {
			call = callImag64
		}
	default:
		return c.badBuiltinCallArgType(sym.Name, node.Args[0], tin, "complex")
	}
	t := r.FuncOf([]r.Type{tin}, []r.Type{tout}, false)
	sym.Type = t
	fun := exprLit(Lit{Type: t, Value: call}, &sym)
	// real() and imag() of a constant are constants: they can be computed at compile time
	return newCall1(fun, arg, arg.Const(), tout)
}

// ============================ support functions =============================

// call_builtin compiles a call to a builtin function: append, cap, copy, delete, len, make, new...
func call_builtin(c *Call) I {
	// builtin functions are always literals, i.e. funindex == NoIndex thus not stored in Env.Binds[]
	// we must retrieve them directly from c.Fun.Value
	if !c.Fun.Const() {
		Errorf("internal error: call_builtin() invoked for non-constant function %#v. use one of the callXretY() instead", c.Fun)
	} else if c.Fun.Sym == nil {
		Errorf("internal error: call_builtin() invoked for non-name function %#v. use one of the callXretY() instead", c.Fun)
	}
	args := c.Args
	argfuns := make([]I, len(args))
	for i, arg := range args {
		argfuns[i] = arg.WithFun()
	}
	if false {
		argtypes := make([]r.Type, len(args))
		for i, arg := range args {
			argtypes[i] = arg.Type
		}
		// Debugf("compiling builtin %s() <%v> with arg types %v", c.Fun.Sym.Name, r.TypeOf(c.Fun.Value), argtypes)
	}
	var call I
	switch fun := c.Fun.Value.(type) {
	case func(complex64) float32: // real(), imag()
		argfun := argfuns[0].(func(*Env) complex64)
		call = func(env *Env) float32 {
			arg := argfun(env)
			return fun(arg)
		}
	case func(complex128) float64: // real(), imag()
		argfun := argfuns[0].(func(*Env) complex128)
		call = func(env *Env) float64 {
			arg := argfun(env)
			return fun(arg)
		}
	case func(string) int: // len(string)
		argfun := argfuns[0].(func(*Env) string)
		call = func(env *Env) int {
			arg := argfun(env)
			return fun(arg)
		}
	case func([]byte, string) int: // copy([]byte, string)
		arg0fun := args[0].AsX1()
		if args[1].Const() {
			// string is a literal
			arg1const := args[1].Value.(string)
			call = func(env *Env) int {
				// arg0 is "assignable to []byte"
				arg0 := arg0fun(env)
				if arg0.Type() != TypeOfSliceOfByte {
					arg0 = arg0.Convert(TypeOfSliceOfByte)
				}
				return fun(arg0.Interface().([]byte), arg1const)
			}
		} else {
			arg1fun := args[1].Fun.(func(*Env) string)
			call = func(env *Env) int {
				// arg0 is "assignable to []byte"
				arg0 := arg0fun(env)
				if arg0.Type() != TypeOfSliceOfByte {
					arg0 = arg0.Convert(TypeOfSliceOfByte)
				}
				arg1 := arg1fun(env)
				return fun(arg0.Interface().([]byte), arg1)
			}
		}
	case func(r.Value) int: // cap(), len()
		argfunsX1 := c.MakeArgfuns()
		argfun := argfunsX1[0]
		call = func(env *Env) int {
			arg := argfun(env)
			return fun(arg)
		}
	case func(r.Value, r.Value): // delete()
		argfunsX1 := c.MakeArgfuns()
		call = func(env *Env) {
			arg0 := argfunsX1[0](env)
			arg1 := argfunsX1[1](env)
			fun(arg0, arg1)
		}
	case func(r.Value, r.Value) int: // copy()
		argfunsX1 := c.MakeArgfuns()
		call = func(env *Env) int {
			arg0 := argfunsX1[0](env)
			arg1 := argfunsX1[1](env)
			return fun(arg0, arg1)
		}
	case func(r.Value, ...r.Value) r.Value: // append()
		argfunsX1 := c.MakeArgfuns()
		call = func(env *Env) r.Value {
			args := make([]r.Value, len(argfunsX1))
			for i, argfun := range argfunsX1 {
				args[i] = argfun(env)
			}
			return fun(args[0], args[1:]...)
		}
	case func(r.Type) r.Value: // new(), make()
		arg0 := args[0].Value.(r.Type)
		call = func(env *Env) r.Value {
			return fun(arg0)
		}
	case func(r.Type, int) r.Value: // make()
		arg0 := args[0].Value.(r.Type)
		arg1fun := argfuns[1].(func(*Env) int)
		call = func(env *Env) r.Value {
			arg1 := arg1fun(env)
			return fun(arg0, arg1)
		}
	case func(r.Type, int, int) r.Value: // make()
		arg0 := args[0].Value.(r.Type)
		arg1fun := argfuns[1].(func(*Env) int)
		arg2fun := argfuns[2].(func(*Env) int)
		call = func(env *Env) r.Value {
			arg1 := arg1fun(env)
			arg2 := arg2fun(env)
			return fun(arg0, arg1, arg2)
		}
	default:
		Errorf("unimplemented call_builtin() for function type %v", r.TypeOf(fun))
	}
	return call
}

// callBuiltinFunc invokes the appropriate compiler for a call to a builtin function: cap, copy, len, make, new...
func (c *Comp) callBuiltinFunc(fun *Expr, node *ast.CallExpr) *Call {
	builtin := fun.Value.(Builtin)
	if fun.Sym == nil {
		c.Errorf("invalid call to non-name builtin: %v", node)
		return nil
	}
	nmin := builtin.ArgMin
	nmax := builtin.ArgMax
	n := len(node.Args)
	if n < nmin || n > nmax {
		return c.badBuiltinCallArgNum(fun.Sym.Name, nmin, nmax, node.Args)
	}
	return builtin.compile(c, *fun.Sym, node)
}

func (c *Comp) badBuiltinCallArgNum(name string, nmin int, nmax int, args []ast.Expr) *Call {
	prefix := "not enough"
	nargs := len(args)
	if nargs > nmax {
		prefix = "too many"
	}
	str := fmt.Sprintf("%d", nmin)
	if nmax <= nmin {
	} else if nmax == nmin+1 {
		str = fmt.Sprintf("%s or %d", str, nmax)
	} else if nmax < MaxInt {
		str = fmt.Sprintf("%s to %d", str, nmax)
	} else {
		str = fmt.Sprintf("%s or more", str)
	}
	c.Errorf("%s arguments in call to builtin %s(): expecting %s, found %d: %v", prefix, name, str, nargs, args)
	return nil
}

func (c *Comp) badBuiltinCallArgType(name string, arg ast.Expr, tactual r.Type, texpected interface{}) *Call {
	c.Errorf("cannot use %v <%v> as %v in builtin %s()", arg, tactual, texpected, name)
	return nil
}
