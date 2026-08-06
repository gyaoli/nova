package log

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var myLogger = newLogger()
var defaultTextWriter = os.Stdout
var defaultFileConfig = FileConfig{
	LogLevel:      "info",
	LogFileName:   "console.log",
	LogPath:       "./syslog",
	LogMaxSize:    200,
	LogMaxAge:     15,
	LogMaxBackups: 10,
	Compress:      false,
}

type FileConfig struct {
	LogLevel string

	// LogFileName is the active file name, such as console.log.
	// After rotation, lumberjack renames it to console-2006-01-02T15-04-05.000.log.
	LogFileName string
	LogPath     string

	LogMaxSize    int
	LogMaxAge     int
	LogMaxBackups int
	Compress      bool
}
type Logger struct {
	*zap.Logger
	sugar *zap.SugaredLogger
}

func Init(ops ...Option) error {
	logger := newLogger(ops...)
	myLogger = logger
	zap.ReplaceGlobals(logger.Logger)
	Info("logger init success.")
	return nil
}

func newLogger(ops ...Option) *Logger {
	o := defaultOptions()
	o.apply(ops...)
	encoder := o.encoder
	if encoder == nil {
		encoder = GetTextEncoder()
	}
	core := o.core
	if core == nil {
		core = zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(o.writeSyncers...), o.level)
	}
	logger := &Logger{}
	logger.Logger = zap.New(core, o.zapOptions...)
	logger.sugar = logger.Sugar()
	return logger
}

func GetLogger() *Logger {
	return myLogger
}

func Sync() error {
	return myLogger.Sync()
}

func GetTextEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	encoderConfig.EncodeName = zapcore.FullNameEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func GetJsonEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	encoderConfig.EncodeName = zapcore.FullNameEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}

func NewFileWriteSyncer(config FileConfig) zapcore.WriteSyncer {
	config = normalizeFileConfig(config)
	fmt.Println(filepath.Join(config.LogPath, config.LogFileName))
	return zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join(config.LogPath, config.LogFileName),
		MaxSize:    config.LogMaxSize,
		MaxAge:     config.LogMaxAge,
		MaxBackups: config.LogMaxBackups,
		Compress:   config.Compress,
		LocalTime:  true,
	})
}

func defaultFileWriteSyncer() zapcore.WriteSyncer {
	return NewFileWriteSyncer(defaultFileConfig)
}

func normalizeFileConfig(config FileConfig) FileConfig {
	if config.LogLevel == "" {
		config.LogLevel = defaultFileConfig.LogLevel
	}
	if config.LogPath == "" {
		config.LogPath = defaultFileConfig.LogPath
	}
	if config.LogFileName == "" {
		config.LogFileName = defaultFileConfig.LogFileName
	}
	if config.LogMaxSize <= 0 {
		config.LogMaxSize = defaultFileConfig.LogMaxSize
	}
	if config.LogMaxAge <= 0 {
		config.LogMaxAge = defaultFileConfig.LogMaxAge
	}
	if config.LogMaxBackups <= 0 {
		config.LogMaxBackups = defaultFileConfig.LogMaxBackups
	}
	return config
}

func strToLevel(strLevel string) (level zapcore.Level) {
	switch strLevel {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	case "panic":
		level = zap.PanicLevel
	case "fatal":
		level = zap.FatalLevel
	default:
		level = zap.InfoLevel
	}
	return
}

func With(fields ...zap.Field) *Logger {
	next := myLogger.With(fields...)
	return &Logger{
		Logger: next,
		sugar:  next.Sugar(),
	}
}

func Debug(msg string, fields ...zap.Field) {
	myLogger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	myLogger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	myLogger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	myLogger.Error(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	myLogger.Panic(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	myLogger.Fatal(msg, fields...)

}

func Debugf(template string, args ...any) {
	myLogger.sugar.Debugf(template, args...)
}

func Infof(template string, args ...any) {
	myLogger.sugar.Infof(template, args...)
}

func Warnf(template string, args ...any) {
	myLogger.sugar.Warnf(template, args...)
}

func Errorf(template string, args ...any) {
	myLogger.sugar.Errorf(template, args...)
}

func Panicf(template string, args ...any) {
	myLogger.sugar.Panicf(template, args...)
}

func Fatalf(template string, args ...any) {
	myLogger.sugar.Fatalf(template, args...)
}

func Any(key string, value any) zap.Field {
	return zap.Any(key, value)
}

func String(key, value string) zap.Field {
	return zap.String(key, value)
}

func Bool(key string, v bool) zap.Field {
	return zap.Bool(key, v)
}

func Time(key string, v time.Time) zap.Field {
	return zap.Time(key, v)
}

func Duration(key string, v time.Duration) zap.Field {
	return zap.Duration(key, v)
}

func Int(key string, value int) zap.Field {
	return zap.Int(key, value)
}

func Int64(key string, value int64) zap.Field {
	return zap.Int64(key, value)
}

func Int32(key string, value int32) zap.Field {
	return zap.Int32(key, value)
}

func Int16(key string, value int16) zap.Field {
	return zap.Int16(key, value)
}

func Int8(key string, value int8) zap.Field {
	return zap.Int8(key, value)
}

func Uint(key string, value uint) zap.Field {
	return zap.Uint(key, value)
}

func Uint64(key string, v uint64) zap.Field {
	return zap.Uint64(key, v)
}

func Uint32(key string, value uint32) zap.Field {
	return zap.Uint32(key, value)
}

func Uint16(key string, value uint16) zap.Field {
	return zap.Uint16(key, value)
}

func Uint8(key string, value uint8) zap.Field {
	return zap.Uint8(key, value)
}

func Float64(key string, v float64) zap.Field {
	return zap.Float64(key, v)
}
