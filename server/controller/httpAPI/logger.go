package httpAPI

import (
	"fmt"
	"io"
	"math/bits"
	"strconv"
	"time"
)

const timeLayout = "2006/01/02 15:04:05"

var logFormat = "%-" + strconv.Itoa(len(timeLayout)) + "s │%-7s│%-26s│%-5s│%-16s│%-12s│%-12s│%-12s│\n"
var fileLogFormatCsv = "%-" + strconv.Itoa(len(timeLayout)) + "s;%-7s;%-26s;%-5s;%-16s;%-12s;%-12s;%-12s\n"

var header = []any{"DATE TIME", "METHOD", "URL", "CODE", "REMOTE ADDR", "READ", "WRITE", "TIME"}

type HttpLogger struct {
	file io.Writer
}

func NewHttpLogger(file io.Writer) *HttpLogger {
	fmt.Printf(logFormat, header...)

	if file != nil {
		fmt.Fprintf(file, fileLogFormatCsv, header...)
	}

	return &HttpLogger{
		file: file,
	}
}

func (l *HttpLogger) Println(v ...any) {
	fmt.Fprintln(l.file, v...)
}

func (l *HttpLogger) Printf(format string, v ...any) {
	fmt.Fprintf(l.file, format, v...)
}

func (l *HttpLogger) Log(method string, url string, code int, remoteAddr string, read uint64, write uint64, duration time.Duration) {
	timeS := time.Now().Format(timeLayout)
	readS := l.formatBytes(read)
	writeS := l.formatBytes(write)
	durationS := l.formatDuration(duration)

	fmt.Printf(logFormat, timeS, method, url, strconv.Itoa(code), remoteAddr, readS, writeS, durationS)
	if l.file != nil {
		fmt.Fprintf(l.file, fileLogFormatCsv, timeS, method, url, strconv.Itoa(code), remoteAddr, readS, writeS, durationS)
	}
}

func (l *HttpLogger) formatBytes(bytes uint64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}

	base := uint(bits.Len64(bytes) / 10)
	val := float64(bytes) / float64(uint64(1<<(base*10)))

	return fmt.Sprintf("%.1f %ciB", val, " KMGTPE"[base])
}

func (l *HttpLogger) formatDuration(d time.Duration) string {
	d = d.Round(time.Millisecond)
	ms := d.Milliseconds()
	s := ms / 1000
	if s == 0 {
		return fmt.Sprintf("%dms", ms)
	}
	return fmt.Sprintf("%ds%dms", s, ms-(s*1000))
}
