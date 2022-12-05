package api

type Response struct {
	IsSuccess bool        `json:"isSuccess"`
	Error     string      `json:"error"`
	Result    interface{} `json:"result"`
}

func NewSuccessResponse(result interface{}) Response {
	return Response{
		IsSuccess: true,
		Result:    result,
	}
}

func NewErrorResponse(error string) Response {
	return Response{
		IsSuccess: false,
		Error:     error,
	}
}
