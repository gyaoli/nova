package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Option func(*options)

type options struct {
	level        zapcore.Level
	encoder      zapcore.Encoder
	writeSyncers []zapcore.WriteSyncer
	core         zapcore.Core
	zapOptions   []zap.Option
}

func defaultOptions() *options {
	return &options{
		encoder:      GetTextEncoder(),
		writeSyncers: []zapcore.WriteSyncer{defaultFileWriteSyncer()},
		level:        strToLevel(defaultFileConfig.LogLevel),
		zapOptions:   []zap.Option{zap.AddCaller()},
	}
}

func (o *options) apply(opts ...Option) {
	for _, opt := range opts {
		if opt != nil {
			opt(o)
		}
	}
}

func WithFileConfig(config FileConfig) Option {
	return func(o *options) {
		config = normalizeFileConfig(config)
		o.level = strToLevel(config.LogLevel)

		o.writeSyncers = []zapcore.WriteSyncer{
			NewFileWriteSyncer(config),
		}
	}
}

func WithConsole() Option {
	return func(o *options) {
		o.writeSyncers = append(o.writeSyncers, zapcore.AddSync(defaultTextWriter))
	}
}

func WithEncoder(encoder zapcore.Encoder) Option {
	return func(o *options) {
		o.encoder = encoder
	}
}

func WithCore(core zapcore.Core) Option {
	return func(o *options) {
		o.core = core
	}

}

func ReplaceWriteSyncers(ws ...zapcore.WriteSyncer) Option {
	return func(o *options) {
		o.writeSyncers = ws
	}
}

func WithWriteSyncers(ws ...zapcore.WriteSyncer) Option {
	return func(o *options) {
		o.writeSyncers = append(o.writeSyncers, ws...)
	}
}

func WithLevel(level zapcore.Level) Option {
	return func(o *options) {
		o.level = level
	}
}

func ReplaceOption(opts ...zap.Option) Option {
	return func(o *options) {
		o.zapOptions = opts
	}
}

func WithOption(opts ...zap.Option) Option {
	return func(o *options) {
		o.zapOptions = append(o.zapOptions, opts...)
	}
}
