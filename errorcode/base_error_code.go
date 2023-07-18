package errorcode

// 标注化服务层错误码的统一
// error 错误信息属于内部错误具体信息
// mapstring 是针对外用户输出信息

const (
	Successfully          = 200
	ErrObjectToJson       = 603
	ErrJsonToObject       = 604
	ErrHttpPostCall       = 608
	ErrJsonToStruct       = 609
	ErrHttpGetCall        = 613
	AiWhippedEggErrorBase = 2000
	ErrService            = 10000
)

var StatusCode = map[int]string{}

func init() {
	StatusCode[Successfully] = "Successfully"
	StatusCode[ErrObjectToJson] = "结构体转Json失败"
	StatusCode[ErrJsonToObject] = "Json转结构体失败"
	StatusCode[ErrService] = "服务端内部错误"
	StatusCode[ErrHttpPostCall] = "调用http post请求出错"
	StatusCode[ErrHttpGetCall] = "调用http get请求出错"
	StatusCode[ErrJsonToStruct] = "json转换为结构体出错"
}

func RegisterErrorCode(errorCodeM map[int]string) int {
	for code, mgs := range errorCodeM {
		StatusCode[code] = mgs
	}
	return 0
}

func QueryErrorMessage(er int) string {
	return StatusCode[er]
}
