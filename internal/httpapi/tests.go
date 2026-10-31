package httpapi

import (
	"fmt"
	"net/http"
)

// 创建测试接口
// @Summary 创建测试并抽题
// @Description 随机抽题，每个题目有 4 个选项。请求体可省略，count 默认 20；词库不足时返回 400。
// @Tags Tests
// @Accept json
// @Produce json
// @Param request body CreateTestRequest false "题目数量"
// @Success 201 {object} domain.Test
// @Header 201 {string} Location "新测试的路径"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/tests [post]
func (api *API) createTest(response http.ResponseWriter, request *http.Request) {
	requestBody := CreateTestRequest{}
	if request.ContentLength != 0 {
		if !readJSON(response, request, &requestBody) {
			return
		}
	}
	if requestBody.Count == 0 {
		requestBody.Count = 20
	}

	test, err := api.tests.Create(request.Context(), requestBody.Count)
	if err != nil {
		writeFailure(response, err)
		return
	}
	response.Header().Set("Location", fmt.Sprintf("/api/v1/tests/%d", test.ID))
	writeJSON(response, http.StatusCreated, test)
}

// 查询测试接口
// @Summary 查询测试
// @Description 未完成时返回题目；完成后额外返回 correct_count、estimated_vocabulary 和 completed_at。
// @Tags Tests
// @Produce json
// @Param id path int true "测试 ID" minimum(1)
// @Success 200 {object} domain.Test
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/tests/{id} [get]
func (api *API) getTest(response http.ResponseWriter, request *http.Request) {
	id, ok := pathID(response, request)
	if !ok {
		return
	}
	test, err := api.tests.Get(request.Context(), id)
	if err != nil {
		writeFailure(response, err)
		return
	}
	writeJSON(response, http.StatusOK, test)
}

// 提交答案接口
// @Summary 提交答案并判分
// @Description 答案必须覆盖该测试的全部题目，word_id 不能重复，selected_index 必须在 0-3。已完成的测试再次提交返回 409。
// @Tags Tests
// @Accept json
// @Produce json
// @Param id path int true "测试 ID" minimum(1)
// @Param answers body SubmitAnswersRequest true "答案"
// @Success 200 {object} domain.Test
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/tests/{id}/answers [post]
func (api *API) submitAnswers(response http.ResponseWriter, request *http.Request) {
	id, ok := pathID(response, request)
	if !ok {
		return
	}
	var input SubmitAnswersRequest
	if !readJSON(response, request, &input) {
		return
	}
	result, err := api.tests.Submit(request.Context(), id, input.Answers)
	if err != nil {
		writeFailure(response, err)
		return
	}
	writeJSON(response, http.StatusOK, result)
}
