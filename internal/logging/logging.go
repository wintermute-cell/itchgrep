package logging

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

type Log struct {
	debugLogger   *log.Logger
	infoLogger    *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger
	fatalLogger   *log.Logger
}

// single instance, initialized in Init
var log_instance Log

// each time we try to use a logging function, we use this to assert that
// the logger has already been initialized.
func assertInstanceExists() {
	if (Log{}) == log_instance {
		// this error is fatal, as it does not depend on the user system, only
		// on the program correctness itself.
		Init("", true)
		log.Printf("Logging instance not initialized, initializing with defaults.")
	}
}

// callerInfo retrieves the filename and line number of the log-function-caller.
// we need this since as we proxy our logging over this module, the
// log.Lshortfile information is lost. This function retrieves basically the
// same information, just a little deeper in the callstack.
func callerInfo() string {
	_, file, line, ok := runtime.Caller(2) // 2 levels up the call stack to get the caller of Info function
	if !ok {
		return "unknown:0"
	}
	shortFile := filepath.Base(file) // Extract just the filename without the full path
	return shortFile + ":" + strconv.Itoa(line)
}

// Init initializes a globally usable, custom log instance. It expects a path
// where the log file should be placed. The path will be ignored if logToStdout
// is set to true.
func Init(log_path string, logToStdout bool) {
	var file *os.File

	if logToStdout {
		file = os.Stdout
	} else {
		// using filepath is important here, since path separators are OS dependant
		// and we don't know if the log_path ends with a trailing separator.
		file_path := filepath.Join(filepath.Clean(log_path), "log.txt")

		// try to find out if logpath + logfile exist
		if _, err := os.Stat(file_path); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				// file or path does not exists, create it
				if err := os.MkdirAll(log_path, os.ModePerm); err != nil {
					log.Printf(
						"FAILED TO CREATE LOGFILE AT: %s, due to this error: %s",
						file_path, err)
				}
			} else {
				// os.Stat failed, but not due to the file not existing. Reason unknown.
				log.Printf(
					"FAILED TO CHECK IF LOGFILE EXISTS: %s, due to this error: %s",
					file_path, err)
			}
		}
		var err error
		file, err = os.OpenFile(file_path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Printf("FAILED TO OPEN LOG FILE AT: %s, due to this error: %s", log_path, err)
		}
	}

	log_instance.debugLogger = log.New(file, "DEBG: ", log.Ldate|log.Ltime)
	log_instance.infoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime)
	log_instance.warningLogger = log.New(file, "WARN: ", log.Ldate|log.Ltime)
	log_instance.errorLogger = log.New(file, "ERRO: ", log.Ldate|log.Ltime)
	log_instance.fatalLogger = log.New(file, "FATL: ", log.Ldate|log.Ltime)
	// note: ERRO and FATL sounds dumb, but having the same length for every prefix
	// improves readability
}

// Info writes to the log-file specified in the settings,
// using the 'DEBG:' specifier.
func Debug(format string, v ...any) {
	assertInstanceExists()
	caller := callerInfo()
	log_instance.debugLogger.Printf("%s: "+format, append([]interface{}{caller}, v...)...)
}

// Info writes to the log-file specified in the settings,
// using the 'INFO:' specifier.
func Info(format string, v ...any) {
	assertInstanceExists()
	caller := callerInfo()
	log_instance.infoLogger.Printf("%s: "+format, append([]interface{}{caller}, v...)...)
}

// Warning writes to the log-file specified in the settings,
// using the 'WARN:' specifier.
func Warning(format string, v ...any) {
	assertInstanceExists()
	caller := callerInfo()
	log_instance.warningLogger.Printf("%s: "+format, append([]interface{}{caller}, v...)...)
}

// Error writes to the log-file specified in the settings,
// using the 'ERRO:' specifier.
func Error(format string, v ...any) {
	assertInstanceExists()
	caller := callerInfo()
	log_instance.errorLogger.Printf("%s: "+format, append([]interface{}{caller}, v...)...)
}

// Fatal writes to the log-file specified in the settings,
// using the 'FATL:' specifier, and calls os.Exit(1) afterwards.
func Fatal(format string, v ...any) {
	assertInstanceExists()
	caller := callerInfo()
	log_instance.fatalLogger.Printf("%s: "+format, append([]interface{}{caller}, v...)...)
	os.Exit(1)
}
