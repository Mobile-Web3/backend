package api

type Response struct {
	IsSuccess bool        `json:"isSuccess"`
	Error     string      `json:"error"`
	Result    interface{} `json:"result"`
}

func newSuccessResponse(result interface{}) Response {
	return Response{
		IsSuccess: true,
		Result:    result,
	}
}

func newErrorResponse(error string) Response {
	return Response{
		IsSuccess: false,
		Error:     error,
	}
}
