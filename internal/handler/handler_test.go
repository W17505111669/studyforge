package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ==================== Health Check ====================

func TestHealthCheck(t *testing.T) {
	db := setupTestDB(t)
	h := setupTestHandler(db)
	router := setupTestRouter(h)

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200，实际: %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("响应 JSON 解析失败: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("期望 status=ok，实际: %v", resp["status"])
	}

	if resp["service"] != "StudyForge Pro" {
		t.Errorf("期望 service=StudyForge Pro，实际: %v", resp["service"])
	}
}

// ==================== Login ====================

func TestLogin_InvalidCredentials_WrongPassword(t *testing.T) {
	db := setupTestDB(t)
	h := setupTestHandler(db)
	router := setupTestRouter(h)

	createTestUser(t, db, "testuser", "password123")

	body, _ := json.Marshal(map[string]string{
		"username": "testuser",
		"password": "wrongpassword",
	})
	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("期望状态码 401，实际: %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] == nil {
		t.Error("期望响应中包含 error 字段")
	}
}

func TestLogin_InvalidCredentials_NonexistentUser(t *testing.T) {
	db := setupTestDB(t)
	h := setupTestHandler(db)
	router := setupTestRouter(h)

	body, _ := json.Marshal(map[string]string{
		"username": "nonexistent",
		"password": "password123",
	})
	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("期望状态码 401，实际: %d", w.Code)
	}
}

func TestLogin_Success(t *testing.T) {
	db := setupTestDB(t)
	h := setupTestHandler(db)
	router := setupTestRouter(h)

	createTestUser(t, db, "testuser", "password123")

	body, _ := json.Marshal(map[string]string{
		"username": "testuser",
		"password": "password123",
	})
	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, 响应: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("响应 JSON 解析失败: %v", err)
	}

	if _, ok := resp["token"]; !ok {
		t.Error("期望响应中包含 token 字段")
	}

	user, ok := resp["user"].(map[string]interface{})
	if !ok {
		t.Fatal("期望响应中包含 user 对象")
	}

	if user["username"] != "testuser" {
		t.Errorf("期望 username=testuser，实际: %v", user["username"])
	}

	if user["email"] != "testuser@test.com" {
		t.Errorf("期望 email=testuser@test.com，实际: %v", user["email"])
	}

	if user["nickname"] != "Test testuser" {
		t.Errorf("期望 nickname=Test testuser，实际: %v", user["nickname"])
	}
}

func TestLogin_MissingFields(t *testing.T) {
	db := setupTestDB(t)
	h := setupTestHandler(db)
	router := setupTestRouter(h)

	// 缺少 password
	body, _ := json.Marshal(map[string]string{
		"username": "testuser",
	})
	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际: %d", w.Code)
	}
}

// ==================== Register ====================

func TestRegister_Success(t *testing.T) {
	db := setupTestDB(t)
	h := setupTestHandler(db)
	router := setupTestRouter(h)

	body, _ := json.Marshal(map[string]string{
		"username": "newuser",
		"email":    "newuser@test.com",
		"password": "password123",
		"nickname": "New User",
	})
	req := httptest.NewRequest("POST", "/api/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("期望状态码 201，实际: %d, 响应: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("响应 JSON 解析失败: %v", err)
	}

	// 验证 token
	token, ok := resp["token"].(string)
	if !ok || token == "" {
		t.Error("期望响应中包含非空 token 字段")
	}

	// 验证 user 对象
	user, ok := resp["user"].(map[string]interface{})
	if !ok {
		t.Fatal("期望响应中包含 user 对象")
	}

	if user["username"] != "newuser" {
		t.Errorf("期望 username=newuser，实际: %v", user["username"])
	}

	if user["email"] != "newuser@test.com" {
		t.Errorf("期望 email=newuser@test.com，实际: %v", user["email"])
	}

	if user["nickname"] != "New User" {
		t.Errorf("期望 nickname=New User，实际: %v", user["nickname"])
	}

	// 验证 user.id 是 UUID（非空字符串）
	id, _ := user["id"].(string)
	if id == "" {
		t.Error("期望 user.id 非空（应自动生成 UUID）")
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	db := setupTestDB(t)
	h := setupTestHandler(db)
	router := setupTestRouter(h)

	createTestUser(t, db, "testuser", "password123")

	body, _ := json.Marshal(map[string]string{
		"username": "testuser",
		"email":    "different@test.com",
		"password": "password123",
	})
	req := httptest.NewRequest("POST", "/api/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("期望状态码 409，实际: %d, 响应: %s", w.Code, w.Body.String())
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	db := setupTestDB(t)
	h := setupTestHandler(db)
	router := setupTestRouter(h)

	body, _ := json.Marshal(map[string]string{
		"username": "newuser",
		"email":    "not-an-email",
		"password": "password123",
	})
	req := httptest.NewRequest("POST", "/api/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400，实际: %d", w.Code)
	}
}

func TestRegister_ShortUsername(t *testing.T) {
	db := setupTestDB(t)
	h := setupTestHandler(db)
	router := setupTestRouter(h)

	body, _ := json.Marshal(map[string]string{
		"username": "ab",
		"email":    "ab@test.com",
		"password": "password123",
	})
	req := httptest.NewRequest("POST", "/api/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（用户名至少3字符），实际: %d", w.Code)
	}
}

func TestRegister_ShortPassword(t *testing.T) {
	db := setupTestDB(t)
	h := setupTestHandler(db)
	router := setupTestRouter(h)

	body, _ := json.Marshal(map[string]string{
		"username": "newuser",
		"email":    "new@test.com",
		"password": "12345",
	})
	req := httptest.NewRequest("POST", "/api/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 400（密码至少6字符），实际: %d", w.Code)
	}
}

// ==================== List Materials ====================

func TestListMaterials_Empty(t *testing.T) {
	db := setupTestDB(t)
	h := setupTestHandler(db)
	router := setupTestRouter(h)

	createTestUser(t, db, "testuser", "password123")
	token := getAuthToken(t, router, "testuser", "password123")

	req := authRequest(t, "GET", "/api/materials", token, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际: %d, 响应: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("响应 JSON 解析失败: %v", err)
	}

	// 验证分页 wrapper 结构
	data, ok := resp["data"]
	if !ok {
		t.Fatal("期望响应中包含 data 字段")
	}

	arr, ok := data.([]interface{})
	if !ok {
		t.Fatalf("期望 data 为数组，实际类型: %T", data)
	}

	if len(arr) != 0 {
		t.Errorf("期望空数组（新用户无材料），实际长度: %d", len(arr))
	}

	// 验证 total=0
	total, ok := resp["total"].(float64)
	if !ok || total != 0 {
		t.Errorf("期望 total=0，实际: %v", resp["total"])
	}

	// 验证分页字段
	if _, ok := resp["limit"]; !ok {
		t.Error("期望响应中包含 limit 字段")
	}
	if _, ok := resp["offset"]; !ok {
		t.Error("期望响应中包含 offset 字段")
	}
}

func TestListMaterials_RequiresAuth(t *testing.T) {
	db := setupTestDB(t)
	h := setupTestHandler(db)
	router := setupTestRouter(h)

	// 无 token 请求认证路由
	req := httptest.NewRequest("GET", "/api/materials", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("期望无 token 请求返回 401，实际: %d", w.Code)
	}
}

func TestListMaterials_ExpiredToken(t *testing.T) {
	db := setupTestDB(t)
	h := setupTestHandler(db)
	router := setupTestRouter(h)

	// 使用无效的 token
	req := httptest.NewRequest("GET", "/api/materials", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-string")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("期望无效 token 返回 401，实际: %d", w.Code)
	}
}

// ==================== Upload Material ====================

func TestUploadMaterial_Success(t *testing.T) {
	db := setupTestDB(t)
	h := setupTestHandler(db)
	router := setupTestRouter(h)

	createTestUser(t, db, "testuser", "password123")
	token := getAuthToken(t, router, "testuser", "password123")

	body := map[string]string{
		"title":        "Go 并发编程",
		"content_type": "text",
		"content":      "Go 语言的并发模型基于 goroutine 和 channel。",
	}

	req := authRequest(t, "POST", "/api/materials", token, body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("期望状态码 201，实际: %d, 响应: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["title"] != "Go 并发编程" {
		t.Errorf("期望 title=Go 并发编程，实际: %v", resp["title"])
	}

	if resp["status"] != "pending" {
		t.Errorf("期望 status=pending，实际: %v", resp["status"])
	}

	if resp["id"] == nil || resp["id"] == "" {
		t.Error("期望材料 id 非空（应自动生成 UUID）")
	}
}

// ==================== Register + Login 集成 ====================

func TestRegisterThenLogin(t *testing.T) {
	db := setupTestDB(t)
	h := setupTestHandler(db)
	router := setupTestRouter(h)

	// 先注册
	regBody, _ := json.Marshal(map[string]string{
		"username": "flowuser",
		"email":    "flow@test.com",
		"password": "password123",
		"nickname": "Flow User",
	})
	regReq := httptest.NewRequest("POST", "/api/register", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	router.ServeHTTP(regW, regReq)

	if regW.Code != http.StatusCreated {
		t.Fatalf("注册失败，状态码: %d", regW.Code)
	}

	// 再用相同凭据登录
	loginBody, _ := json.Marshal(map[string]string{
		"username": "flowuser",
		"password": "password123",
	})
	loginReq := httptest.NewRequest("POST", "/api/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)

	if loginW.Code != http.StatusOK {
		t.Fatalf("登录失败，状态码: %d, 响应: %s", loginW.Code, loginW.Body.String())
	}

	var loginResp map[string]interface{}
	json.Unmarshal(loginW.Body.Bytes(), &loginResp)

	if _, ok := loginResp["token"]; !ok {
		t.Error("登录响应中缺少 token")
	}
}

// ==================== Register + List Materials 集成 ====================

func TestRegisterThenListMaterials(t *testing.T) {
	db := setupTestDB(t)
	h := setupTestHandler(db)
	router := setupTestRouter(h)

	// 注册用户（通过 API）
	regBody, _ := json.Marshal(map[string]string{
		"username": "matuser",
		"email":    "mat@test.com",
		"password": "password123",
	})
	regReq := httptest.NewRequest("POST", "/api/register", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	router.ServeHTTP(regW, regReq)

	if regW.Code != http.StatusCreated {
		t.Fatalf("注册失败: %d", regW.Code)
	}

	var regResp map[string]interface{}
	json.Unmarshal(regW.Body.Bytes(), &regResp)
	token := regResp["token"].(string)

	// 列出材料（应返回空列表）
	listReq := authRequest(t, "GET", "/api/materials", token, nil)
	listW := httptest.NewRecorder()
	router.ServeHTTP(listW, listReq)

	if listW.Code != http.StatusOK {
		t.Fatalf("列表请求失败: %d", listW.Code)
	}

	var listResp map[string]interface{}
	json.Unmarshal(listW.Body.Bytes(), &listResp)

	data := listResp["data"].([]interface{})
	if len(data) != 0 {
		t.Errorf("新用户应返回空材料列表，实际: %d 条", len(data))
	}
}
