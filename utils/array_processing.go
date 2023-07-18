package utils

import (
	"ai-guandan/errorcode"
	"encoding/json"
	log "github.com/sirupsen/logrus"
)

// JsonStrToStruct @description: jsonstring --> struct
// @parameter jsonStr
// @parameter eventStruct
// @return int
func JsonStrToStruct(jsonStr string, eventStruct interface{}) int {
	if err := json.Unmarshal([]byte(jsonStr), &eventStruct); err != nil {
		log.Errorln(err)
		return errorcode.ErrJsonToStruct
	}
	return errorcode.Successfully
}

// MapToStruct @description: map --> struct
// @parameter mapBean
// @parameter eventStruct
func MapToStruct(mapBean map[string]interface{}, eventStruct interface{}) {
	//将 map 转换为指定的结构体
	str, err := MapToJsonStr(mapBean)
	if err != nil {
		log.Errorln(err)
	}
	JsonStrToStruct(str, &eventStruct)
}

// MapToJsonStr @description: map --> jsonstring
// @parameter mapBean
// @return str
// @return err
func MapToJsonStr(mapBean map[string]interface{}) (str string, err error) {
	bytes, err := json.Marshal(mapBean)
	if err != nil {
		log.Errorln(err)
		return
	}
	return string(bytes), err
}

// ObjectAToObjectB @description: ObjectA -> ObjectB(包含数组转换)
// @parameter objA
// @parameter objB
// @return code
func ObjectAToObjectB(objA, objB interface{}) (code int) {
	tempStr, err := json.Marshal(objA)
	if err != nil {
		log.Errorln("error:", err)
		return errorcode.ErrObjectToJson
	}
	errT := json.Unmarshal(tempStr, objB)
	if errT != nil {
		log.Errorln("error:", errT)
		return errorcode.ErrJsonToObject
	}
	return errorcode.Successfully
}
