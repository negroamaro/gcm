////////////////////////////////////////////////////////////////////////////////
// Copyright (c) 2015 Negroamaro. All rights reserved.                        //
////////////////////////////////////////////////////////////////////////////////

package gcm

import (
	"bytes"
	"reflect"
	"runtime"
	"time"
)

var managed map[string]*goroutine

// package initializer.
func init() {
	managed = make(map[string]*goroutine, maxCapacity)
	go statusMonitor()
}

// goroutine status monitor.
func statusMonitor() {
	for {
		for _, v := range managed {
			if len(v.contexts) == 0 {
				v.status = Stopped
			}
		}
		<-time.After(time.Second * 3)
	}
}

// Register
//
//   [args]
//     gofunc : function object, that is run as goroutine.
//
//   [returns]
//     string : unique name of 'gofunc' in the gcm package.
//              you can use this name for 'gofunc' argument for other functions.
//              e.g. UnRegister, Start, ChangeMultiplicity, Stop, GetStatus, GetMultiplicity.
//     error  : if gofunc already registered, return ErrFuncExists
func Register(gofunc interface{}) (string, error) {
	name := getUniqueName(gofunc)
	_, exists := managed[name]
	if exists {
		return name, ErrFuncExists
	}
	managed[name] = &goroutine{
		reflect.ValueOf(gofunc),
		nil,
		0,
		make([]interface{}, maxMultiplicity),
		Stopped}
	return name, nil
}

// UnRegister
//
//   [args]
//     gofunc : function object or function name, that was returned by Register function.
//
//   [returns]
//     error  : if gofunc is not Register, return ErrFuncNotExists
func UnRegister(gofunc interface{}) error {
	name := getUniqueName(gofunc)
	_, exists := managed[name]
	if !exists {
		return ErrFuncNotExists
	}
	delete(managed, name)
	return nil
}

//
func Start(gofunc interface{}, m int, args ...interface{}) error {
	return nil
}

//
func ChangeMultiplicity(gofunc interface{}, m int) error {
	return nil
}

//
func Stop(gofunc interface{}, async bool) error {
	return nil
}

func GetStatus(gofunc interface{}) (Status, error) {
	return Stopped, nil
}

func GetMultiplicity(gofunc interface{}) (int, error) {
	return 0, nil
}

// get unique name of function.
// function name format = <package_name>.<function_name>:'<function_type_name>'
func getUniqueName(gofunc interface{}) string {
	switch gofunc.(type) {
	case string:
		return gofunc.(string)
	default:
		v := reflect.ValueOf(gofunc)
		buf := &bytes.Buffer{}
		buf.WriteString(runtime.FuncForPC(v.Pointer()).Name())
		buf.WriteString(":'")
		buf.WriteString(v.Type().Name())
		buf.WriteString("'")
		return buf.String()
	}
}

type goroutine struct {
	gofunc       reflect.Value
	args         []interface{}
	multiplicity int
	contexts     []interface{}
	status       Status
}

func (g *goroutine) start() error {
	switch g.status {
	case Running:
		return nil
	case Stopping:
		return nil // TODO new error
	}
	for i := 0; i < g.multiplicity; i++ {
		// TODO impl context
		ctx := struct{}{}
		go g.gofunc.Call(getValues(ctx, g.args...))
		g.contexts[i] = ctx
	}
	g.status = Running
	return nil
}

func (g *goroutine) changeMultiplicity(m int) error {
	cur := g.multiplicity
	if m > cur {
		// increase goroutine
		for i := cur; i < m; i++ {
			// TODO impl context
			ctx := struct{}{}
			go g.gofunc.Call(getValues(ctx, g.args...))
			g.contexts[i] = ctx
		}
	} else if m < cur {
		// decrease goroutine
		for i := m; i < cur; i++ {
			ctx := g.contexts[i]
			if ctx != nil {
				// TODO call ctx.Cancel(cancelCallbackFunc)
				// func() {
				// 	   g.contexts[i] = nil
				// }()
			}
		}
	}
	g.multiplicity = m
	return nil
}

func (g *goroutine) stop(async bool) error {
	switch g.status {
	case Stopping:
		return nil
	case Stopped:
		return nil
	}
	for i := 0; i < maxMultiplicity; i++ {
		ctx := g.contexts[i]
		if ctx != nil {
			// TODO call ctx.Cancel(cancelCallbackFunc)
			// func() {
			// 	   g.contexts[i] = nil
			// }()
		}
	}
	g.status = Stopping
	return nil
}

func getValues(ctx interface{}, args ...interface{}) []reflect.Value {
	values := make([]reflect.Value, len(args)+1)
	values[0] = reflect.ValueOf(ctx)
	for i := 1; i < len(values); i++ {
		values[i] = reflect.ValueOf(args[i-1])
	}
	return values
}

// EOF
