package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
)

// 创建词汇接口
// @Summary 新增词汇
// @Description 单词唯一，重复创建返回 409。
// @Tags Words
// @Accept json
// @Produce json
// @Param word body WordInput true "词汇"
// @Success 201 {object} domain.Word
// @Header 201 {string} Location "新词汇的路径"
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/words [post]
func (api *API) createWord(response http.ResponseWriter, request *http.Request) {
	var input WordInput
	if !readJSON(response, request, &input) {
		return
	}
	created, err := api.words.Create(request.Context(), input.toDomain())
	if err != nil {
		writeFailure(response, err)
		return
	}
	response.Header().Set("Location", fmt.Sprintf("/api/v1/words/%d", created.ID))
	writeJSON(response, http.StatusCreated, created)
}

// 列出词汇接口
// @Summary 分页列出词汇
// @Description 按 id 升序返回，可按 level 过滤。
// @Tags Words
// @Produce json
// @Param level query int false "按难度过滤" Enums(1,2,3,4,5)
// @Param limit query int false "每页数量" default(50) minimum(1) maximum(200)
// @Param offset query int false "偏移量" default(0) minimum(0)
// @Success 200 {object} WordListResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/words [get]
func (api *API) listWords(response http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()

	limit := 50
	if value := query.Get("limit"); value != "" {
		parsedValue, err := strconv.Atoi(value)
		if err != nil || parsedValue < 1 || parsedValue > 200 {
			writeError(response, http.StatusBadRequest, "limit must be between 1 and 200")
			return
		}
		limit = parsedValue
	}

	offset := 0
	if value := query.Get("offset"); value != "" {
		parsedValue, err := strconv.Atoi(value)
		if err != nil || parsedValue < 0 {
			writeError(response, http.StatusBadRequest, "offset must be zero or greater")
			return
		}
		offset = parsedValue
	}

	var level *int
	if value := query.Get("level"); value != "" {
		parsedValue, err := strconv.Atoi(value)
		if err != nil || parsedValue < 1 || parsedValue > 5 {
			writeError(response, http.StatusBadRequest, "level must be between 1 and 5")
			return
		}
		level = &parsedValue
	}

	words, total, err := api.words.List(request.Context(), level, limit, offset)
	if err != nil {
		writeFailure(response, err)
		return
	}
	writeJSON(response, http.StatusOK, WordListResponse{Items: words, Total: total})
}

// 查询词汇接口
// @Summary 查询单个词汇
// @Tags Words
// @Produce json
// @Param id path int true "词汇 ID" minimum(1)
// @Success 200 {object} domain.Word
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/words/{id} [get]
func (api *API) getWord(response http.ResponseWriter, request *http.Request) {
	id, ok := pathID(response, request)
	if !ok {
		return
	}
	word, err := api.words.Get(request.Context(), id)
	if err != nil {
		writeFailure(response, err)
		return
	}
	writeJSON(response, http.StatusOK, word)
}

// 全量更新词汇接口
// @Summary 全量替换词汇
// @Description 必须提供 word、meaning、level 三个字段。
// @Tags Words
// @Accept json
// @Produce json
// @Param id path int true "词汇 ID" minimum(1)
// @Param word body WordInput true "词汇"
// @Success 200 {object} domain.Word
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/words/{id} [put]
func (api *API) putWord(response http.ResponseWriter, request *http.Request) {
	id, ok := pathID(response, request)
	if !ok {
		return
	}
	var input WordInput
	if !readJSON(response, request, &input) {
		return
	}
	updated, err := api.words.Put(request.Context(), id, input.toDomain())
	if err != nil {
		writeFailure(response, err)
		return
	}
	writeJSON(response, http.StatusOK, updated)
}

// 局部更新词汇接口
// @Summary 局部更新词汇
// @Description 至少提供一个字段，只更新传入的字段。
// @Tags Words
// @Accept json
// @Produce json
// @Param id path int true "词汇 ID" minimum(1)
// @Param word body WordPatchInput true "要更新的字段"
// @Success 200 {object} domain.Word
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/words/{id} [patch]
func (api *API) patchWord(response http.ResponseWriter, request *http.Request) {
	id, ok := pathID(response, request)
	if !ok {
		return
	}
	var input WordPatchInput
	if !readJSON(response, request, &input) {
		return
	}
	updated, err := api.words.Patch(request.Context(), id, input.toDomain())
	if err != nil {
		writeFailure(response, err)
		return
	}
	writeJSON(response, http.StatusOK, updated)
}

// 删除词汇接口
// @Summary 删除词汇
// @Description 词汇被任何测试引用时返回 409。
// @Tags Words
// @Produce json
// @Param id path int true "词汇 ID" minimum(1)
// @Success 204 "删除成功"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /api/v1/words/{id} [delete]
func (api *API) deleteWord(response http.ResponseWriter, request *http.Request) {
	id, ok := pathID(response, request)
	if !ok {
		return
	}
	if err := api.words.Delete(request.Context(), id); err != nil {
		writeFailure(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}
