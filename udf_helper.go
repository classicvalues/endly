package endly

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/viant/endly/common"
	"github.com/viant/toolbox"
	"io"
)

func AsMap(source interface{}, state common.Map) (interface{}, error) {
	if source == nil || toolbox.IsMap(source) {
		return source, nil
	}
	if toolbox.IsString(source) {
		buf := new(bytes.Buffer)
		err := toolbox.NewJSONEncoderFactory().Create(buf).Encode(toolbox.AsString(source))
		if err != nil {
			return nil, err
		}
		aMap := make(map[string]interface{})
		err = toolbox.NewJSONDecoderFactory().Create(buf).Decode(aMap)
		if err != nil {
			return nil, err
		}
		return aMap, nil

	}
	return source, nil
}

func AsInt(source interface{}, state common.Map) (interface{}, error) {
	return toolbox.AsInt(source), nil
}

func Md5(source interface{}, state common.Map) (interface{}, error) {
	hash := md5.New()
	io.WriteString(hash, toolbox.AsString(source))
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
