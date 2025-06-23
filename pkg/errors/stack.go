package errors

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

type Frame struct {
	Function string
	File     string
	Line     int
}

func newFrame(pc uintptr) Frame {
	pc = pc - 1 // 调整 PC 保证能正确获取调用点
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return Frame{"unknown", "unknown", 0}
	}
	file, line := fn.FileLine(pc)
	return Frame{
		Function: fn.Name(),
		File:     file,
		Line:     line,
	}
}

// 打印格式化 Frame
func (f Frame) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		if s.Flag('+') {
			fmt.Fprintf(s, "%s\n\t%s:%d", f.Function, f.File, f.Line)
		} else {
			fmt.Fprint(s, filepath.Base(f.File))
		}
	case 'v':
		f.Format(s, 's')
	}
}

// 可用于日志输出、序列化
func (f Frame) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%s %s:%d", f.Function, f.File, f.Line)), nil
}

type stack []Frame

// 栈格式化支持 %+v
func (s *stack) Format(st fmt.State, verb rune) {
	if verb == 'v' && st.Flag('+') {
		for _, frame := range *s {
			fmt.Fprintf(st, "%+v\n", frame)
		}
	}
}

// 栈文本序列化
func (s *stack) MarshalText() ([]byte, error) {
	var b strings.Builder
	for _, f := range *s {
		b.WriteString(fmt.Sprintf("%s %s:%d\n", f.Function, f.File, f.Line))
	}
	return []byte(b.String()), nil
}

// 获取调用栈
func callers() *stack {
	const maxDepth = 32
	var pcs [maxDepth]uintptr

	n := runtime.Callers(3, pcs[:]) // skip: runtime.Callers + callers + 创建栈的函数
	st := make(stack, 0, n)

	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		name := fn.Name()
		if isInternalRuntime(name) {
			continue
		}
		st = append(st, newFrame(pc))
	}

	return &st
}

// 是否是 runtime 或内部调用，过滤掉
func isInternalRuntime(name string) bool {
	return strings.HasPrefix(name, "runtime.") ||
		strings.HasPrefix(name, "testing.") ||
		strings.HasPrefix(name, "internal/") ||
		strings.Contains(name, "goexit") ||
		strings.HasPrefix(name, "reflect.") ||
		strings.HasPrefix(name, "asm_")
}
