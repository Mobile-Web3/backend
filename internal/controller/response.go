package controller

type apiResponse struct {
	IsSuccess bool        `json:"isSuccess"`
	Error     string      `json:"error"`
	Result    interface{} `json:"result"`
}

func createSuccessResponse(result interface{}) apiResponse {
	return apiResponse{
		IsSuccess: true,
		Result:    result,
	}
}

func createErrorResponse(error string) apiResponse {
	return apiResponse{
		IsSuccess: false,
		Error:     error,
	}
}
