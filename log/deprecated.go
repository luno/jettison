package log

import (
	"fmt"
	"os"
)

// Print wraps a call to fmt.Sprint in a jettison log and writes it to the logger.
// Deprecated: Use log.Info or log.Error instead.
func Print(v ...interface{}) {
	print(v...)
}

// Printf wraps a call to fmt.Sprintf in a jettison log and writes it to the logger.
// Deprecated: Use log.Info or log.Error instead.
func Printf(format string, v ...interface{}) {
	printf(format, v...)
}

// Println wraps a call to fmt.Sprintln in a jettison log and writes it to the logger.
// Deprecated: Use log.Info or log.Error instead.
func Println(v ...interface{}) {
	println(v...)
}

// Panic is equivalent to log.Print followed by a panic.
// Deprecated: Use log.Info or log.Error instead and panic manually.
func Panic(v ...interface{}) {
	panic(print(v...))
}

// Panicf is equivalent to log.Printf followed by a panic.
// Deprecated: Use log.Info or log.Error instead and panic manually.
func Panicf(format string, v ...interface{}) {
	panic(printf(format, v...))
}

// Panicln is equivalent to log.Println followed by a panic.
// Deprecated: Use log.Info or log.Error instead and panic manually.
func Panicln(v ...interface{}) {
	panic(println(v...))
}

// Fatal is equivalent to log.Print followed by a call to os.Exit(1).
// Deprecated: Use log.Info or log.Error instead and exit manually.
func Fatal(v ...interface{}) {
	print(v...)
	os.Exit(1)
}

// Fatalf is equivalent to log.Printf followed by a call to os.Exit(1).
// Deprecated: Use log.Info or log.Error instead and exit manually.
func Fatalf(format string, v ...interface{}) {
	printf(format, v...)
	os.Exit(1)
}

// Fatalln is equivalent to log.Println followed by a call to os.Exit(1).
// Deprecated: Use log.Info or log.Error instead and exit manually.
func Fatalln(v ...interface{}) {
	println(v...)
	os.Exit(1)
}

func print(v ...interface{}) string {
	l := newEntry(fmt.Sprint(v...), LevelInfo, 3)
	return logger.Log(Entry(l))
}

func printf(format string, v ...interface{}) string {
	l := newEntry(fmt.Sprintf(format, v...), LevelInfo, 3)
	return logger.Log(Entry(l))
}

func println(v ...interface{}) string {
	l := newEntry(fmt.Sprintln(v...), LevelInfo, 3)
	return logger.Log(Entry(l))
}
