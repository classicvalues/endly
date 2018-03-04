package log

import (
	"time"
	"sync"
	"strings"
	"io"
	"io/ioutil"
	"github.com/viant/toolbox/storage"
)

//File represents a log file
type File struct {
	URL     string
	Content string
	Name    string
	*Type
	ProcessingState *ProcessingState
	LastModified    time.Time
	Size            int
	Records         []*Record
	IndexedRecords  map[string]*Record
	Mutex           *sync.RWMutex
}

//ShiftLogRecord returns and remove the first log record if present
func (f *File) ShiftLogRecord() *Record {
	f.Mutex.Lock()
	defer f.Mutex.Unlock()
	if len(f.Records) == 0 {
		return nil
	}
	result := f.Records[0]
	f.Records = f.Records[1:]
	return result
}

//ShiftLogRecordByIndex returns and remove the first log record if present
func (f *File) ShiftLogRecordByIndex(value string) *Record {
	f.Mutex.Lock()
	defer f.Mutex.Unlock()
	if len(f.Records) == 0 {
		return nil
	}
	result, has := f.IndexedRecords[value]
	if !has {
		result = f.Records[0]
		f.Records = f.Records[1:]
	} else {
		var records = make([]*Record, 0)
		for _, candidate := range f.Records {
			if candidate == result {
				continue
			}
			records = append(records, candidate)
		}
		f.Records = records
	}
	return result
}

//PushLogRecord appends provided log record to the records.
func (f *File) PushLogRecord(record *Record) {
	f.Mutex.Lock()
	defer f.Mutex.Unlock()

	if len(f.Records) == 0 {
		f.Records = make([]*Record, 0)
	}
	f.Records = append(f.Records, record)
	if f.UseIndex() {
		if expr, err := f.GetIndexExpr(); err == nil {
			var indexValue = matchLogIndex(expr, record.Line)
			if indexValue != "" {
				f.IndexedRecords[indexValue] = record
			}
		}
	}
}


//Reset resets processing state
func (f *File) Reset(object storage.Object) {
	f.Mutex.Lock()
	defer f.Mutex.Unlock()
	f.Size = int(object.FileInfo().Size())
	f.LastModified = object.FileInfo().ModTime()
	f.ProcessingState.Reset()
}

//HasPendingLogs returns true if file has pending validation records
func (f *File) HasPendingLogs() bool {
	f.Mutex.Lock()
	defer f.Mutex.Unlock()
	return len(f.Records) > 0
}

func (f *File) readLogRecords(reader io.Reader) error {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	if f.ProcessingState.Position > len(data) {
		return nil
	}
	var line = ""
	var startPosition = f.ProcessingState.Position
	var startLine = f.ProcessingState.Line
	var lineIndex = startLine
	var dataProcessed = 0
	for i := startPosition; i < len(data); i++ {
		dataProcessed++
		aChar := string(data[i])
		if aChar != "\n" && aChar != "\r" {
			line += aChar
			continue
		}

		line = strings.Trim(line, " \r\t")
		lineIndex++
		if f.Exclusion != "" {
			if strings.Contains(line, f.Exclusion) {
				line, dataProcessed = f.ProcessingState.Update(dataProcessed, lineIndex)
				continue
			}
		}
		if f.Inclusion != "" {
			if !strings.Contains(line, f.Inclusion) {
				line, dataProcessed = f.ProcessingState.Update(dataProcessed, lineIndex)
				continue
			}
		}

		if len(line) > 0 {
			f.PushLogRecord(&Record{
				URL:    f.URL,
				Line:   line,
				Number: lineIndex,
			})
		}
		if err != nil {
			return err
		}
		line, dataProcessed = f.ProcessingState.Update(dataProcessed, lineIndex)
	}
	return nil
}