package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/ttarnok/instrument-swap-api/internal/auth"
	"github.com/ttarnok/instrument-swap-api/internal/data"
	"github.com/ttarnok/instrument-swap-api/internal/testhelpers"
)

// MainTestSuite is the testsuite for testing the applications api endpoints with full database integration.
type MainTestSuite struct {
	suite.Suite
	pgContainer    *testhelpers.PostgresContainer
	redisContainer *testhelpers.RedisContainer
	ts             *httptest.Server
	app            *application
	ctx            context.Context
}

// SetupSuite sets up the testsuite, creates a test server with all api endpoints.
// This test server has turned off rate limiter.
func (suite *MainTestSuite) SetupSuite() {

	suite.ctx = context.Background()
	pgContainer, err := testhelpers.CreatePostgresContainer(suite.ctx)
	if err != nil {
		log.Fatal(err)
	}
	suite.pgContainer = pgContainer

	redisContainer, err := testhelpers.CreateRedisContainer(suite.ctx)
	if err != nil {
		log.Fatal(err)
	}
	suite.redisContainer = redisContainer

	//setting up the test application
	var cfg config

	cfg.env = "test"
	cfg.db.dsn = suite.pgContainer.ConnectionString
	cfg.db.maxIdleConns = 25
	cfg.db.maxOpenConns = 25
	cfg.db.maxIdleTime = 15 * time.Minute
	cfg.limiter.requestPerSecond = 2
	cfg.limiter.burst = 4
	cfg.limiter.enabled = false // !!!
	cfg.jwt.secret = "testsecret"
	cfg.redis.address = suite.redisContainer.ConnectionString
	cfg.redis.password = ""
	cfg.redis.db = 0

	db, err := openDB(cfg)
	if err != nil {
		log.Fatal(err)
	}
	opt, err := redis.ParseURL(cfg.redis.address)
	if err != nil {
		log.Fatal(err)
	}
	redisClient, err := auth.NewBlacklistRedisClient(opt.Addr, cfg.redis.password, cfg.redis.db)
	if err != nil {
		log.Fatal(err)
	}

	suite.app = &application{
		config: cfg,
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		models: data.NewModel(db),
		auth:   auth.NewAuth(cfg.jwt.secret, auth.NewBlacklistService(redisClient)),
	}

	suite.ts = httptest.NewServer(suite.app.routes())
}

// TearDownSuite tears down the testsuit. Release all used resources.
func (suite *MainTestSuite) TearDownSuite() {
	defer suite.ts.Close()
	if err := suite.pgContainer.Terminate(suite.ctx); err != nil {
		log.Fatalf("error terminating postgres container: %s", err)
	}
	if err := suite.redisContainer.Terminate(suite.ctx); err != nil {
		log.Fatalf("error terminating redis container: %s", err)
	}
}

// TestHealthCheck tests the healthcheck api endpoint.
func (suite *MainTestSuite) TestHealthCheck() {
	t := suite.T()

	expectedStatusCode := 200

	expectedRespBody := make(map[string]map[string]string)
	expectedRespBody["liveliness"] = make(map[string]string)
	expectedRespBody["liveliness"]["environment"] = "test"
	expectedRespBody["liveliness"]["status"] = "available"
	expectedRespBody["liveliness"]["version"] = "-"

	path := fmt.Sprintf("%s/v1/liveliness", suite.ts.URL)

	resp := testhelpers.DoTestAPICall(t, "GET", path, nil, nil)

	assert.Equal(t, expectedStatusCode, resp.StatusCode, "status code mismatch")

	respBody := testhelpers.GetResponseBody[map[string]map[string]string](t, resp)

	assert.Equal(t, expectedRespBody, respBody, "response body mismatch")

}

func CreateRequestBody[T any](t *testing.T, input T) io.Reader {

	reqBody, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	return bytes.NewBuffer(reqBody)
}

// TestBasicUserStory tests a basic user story.
// This test does the following test calls in the following order.
// 1. User registration, on happy path.
// 2. User login, on happy path.
// 3. Create an instrument on happy path.
// 4. Create another instrument, on happy path.
// 5. Get the first instrument for the user, on the happy path.
// 6. Get all instruments for the user, on the happy path.
func (suite *MainTestSuite) TestBasicUserStory() {
	t := suite.T()

	// ***************************************************************************
	// 1. User registration, on happy path.
	expectedStatusCode := http.StatusCreated

	inputRegister := struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Name:     "John Smith",
		Email:    "johnsmith@example.com",
		Password: "asd123asd1234",
	}

	path := fmt.Sprintf("%s/v1/users", suite.ts.URL)
	resp := testhelpers.DoTestAPICall(t, "POST", path, CreateRequestBody(t, inputRegister), nil)
	assert.Equal(t, expectedStatusCode, resp.StatusCode, "status code mismatch")
	respRegisterBody := testhelpers.GetResponseBody[map[string]*data.User](t, resp)
	assert.NotEqual(t, respRegisterBody["user"].ID, "user ID should not be 0")
	assert.Equal(t, inputRegister.Name, respRegisterBody["user"].Name, "user name mismatch")
	assert.Equal(t, inputRegister.Email, respRegisterBody["user"].Email, "email mismatch")

	// ***************************************************************************
	// 2. User login, on happy path.
	expectedStatusCode = http.StatusCreated
	inputLogin := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Email:    "johnsmith@example.com",
		Password: "asd123asd1234",
	}
	path = fmt.Sprintf("%s/v1/token", suite.ts.URL)
	resp = testhelpers.DoTestAPICall(t, "POST", path, CreateRequestBody(t, inputLogin), nil)
	assert.Equal(t, expectedStatusCode, resp.StatusCode, "status code mismatch")
	respLoginBody := testhelpers.GetResponseBody[map[string]string](t, resp)
	accessToken, ok := respLoginBody["access"]
	assert.True(t, ok, "access token should be sent")
	assert.NotEmpty(t, accessToken, "access token should not be empty")
	refreshToken, ok := respLoginBody["refresh"]
	assert.True(t, ok, "refresh token should be sent")
	assert.NotEmpty(t, refreshToken, "refresh token should not be empty")

	// ***************************************************************************
	// 3. Create an instrument on happy path.
	expectedStatusCode = http.StatusCreated
	inputInstrument1 := struct {
		Name            string   `json:"name"`
		Manufacturer    string   `json:"manufacturer"`
		ManufactureYear int32    `json:"manufacture_year"`
		Type            string   `json:"type"`
		EstimatedValue  int64    `json:"estimated_value"`
		Condition       string   `json:"condition"`
		Description     string   `json:"description"`
		FamousOwners    []string `json:"famous_owners"`
	}{
		Name:            "M1",
		Manufacturer:    "Korg",
		ManufactureYear: 1991,
		Type:            "synthesizer",
		EstimatedValue:  1000,
		Condition:       "excellent",
		Description:     "A fairly nice workstation.",
		FamousOwners:    []string{"The Orb", "Orbital"},
	}
	path = fmt.Sprintf("%s/v1/instruments", suite.ts.URL)
	headers := make(map[string]string)
	headers["Authorization"] = strings.Join([]string{"Bearer", accessToken}, " ")
	resp = testhelpers.DoTestAPICall(t, "POST", path, CreateRequestBody(t, inputInstrument1), headers)
	assert.Equal(t, expectedStatusCode, resp.StatusCode, "status code mismatch")
	respInstrumentBody1 := testhelpers.GetResponseBody[map[string]*data.Instrument](t, resp)
	createdInstrument1, ok := respInstrumentBody1["instrument"]

	assert.True(t, ok, "the newly created instrument should be sent")
	assert.NotEmpty(t, createdInstrument1.ID, "the newly created instrumnent id shoud not be empty or 0")
	assert.NotEmpty(t, createdInstrument1.CreatedAt, "created at date should not be empty")
	assert.Equal(t, inputInstrument1.Name, createdInstrument1.Name, "instrument name mismatch")
	assert.Equal(t, inputInstrument1.Manufacturer, createdInstrument1.Manufacturer, "instrument manufacturer mismatch")
	assert.Equal(t, inputInstrument1.ManufactureYear, createdInstrument1.ManufactureYear, "instrument manufacture year mismatch")
	assert.Equal(t, inputInstrument1.Type, createdInstrument1.Type, "instrument type mismatch")
	assert.Equal(t, inputInstrument1.EstimatedValue, createdInstrument1.EstimatedValue, "estimated value mismatch")
	assert.Equal(t, inputInstrument1.Condition, createdInstrument1.Condition, "condition mismatch")
	assert.Equal(t, inputInstrument1.Description, createdInstrument1.Description, "description mismatch")
	assert.Equal(t, inputInstrument1.FamousOwners, createdInstrument1.FamousOwners, "famous users mismatch")
	assert.Equal(t, respRegisterBody["user"].ID, createdInstrument1.OwnerUserID, "owner instrumetn id mismatch")
	assert.NotEmpty(t, createdInstrument1.Version, "version should not be empty")

	// ***************************************************************************
	// 4. Create another instrument, on happy path.
	expectedStatusCode = http.StatusCreated
	inputInstrument2 := struct {
		Name            string   `json:"name"`
		Manufacturer    string   `json:"manufacturer"`
		ManufactureYear int32    `json:"manufacture_year"`
		Type            string   `json:"type"`
		EstimatedValue  int64    `json:"estimated_value"`
		Condition       string   `json:"condition"`
		Description     string   `json:"description"`
		FamousOwners    []string `json:"famous_owners"`
	}{
		Name:            "Minimoog",
		Manufacturer:    "Moog",
		ManufactureYear: 1978,
		Type:            "synthesizer",
		EstimatedValue:  10000,
		Condition:       "excellent",
		Description:     "A really nice subtractive synth.",
		FamousOwners:    []string{"Tangerine Dream"},
	}
	path = fmt.Sprintf("%s/v1/instruments", suite.ts.URL)
	headers = make(map[string]string)
	headers["Authorization"] = strings.Join([]string{"Bearer", accessToken}, " ")
	resp = testhelpers.DoTestAPICall(t, "POST", path, CreateRequestBody(t, inputInstrument2), headers)
	assert.Equal(t, expectedStatusCode, resp.StatusCode, "status code mismatch")
	respInstrumentBody2 := testhelpers.GetResponseBody[map[string]*data.Instrument](t, resp)
	createdInstrument2, ok := respInstrumentBody2["instrument"]

	assert.True(t, ok, "the newly created instrument should be sent")
	assert.NotEmpty(t, createdInstrument2.ID, "the newly created instrumnent id shoud not be empty or 0")
	assert.NotEmpty(t, createdInstrument2.CreatedAt, "created at date should not be empty")
	assert.Equal(t, inputInstrument2.Name, createdInstrument2.Name, "instrument name mismatch")
	assert.Equal(t, inputInstrument2.Manufacturer, createdInstrument2.Manufacturer, "instrument manufacturer mismatch")
	assert.Equal(t, inputInstrument2.ManufactureYear, createdInstrument2.ManufactureYear, "instrument manufacture year mismatch")
	assert.Equal(t, inputInstrument2.Type, createdInstrument2.Type, "instrument type mismatch")
	assert.Equal(t, inputInstrument2.EstimatedValue, createdInstrument2.EstimatedValue, "estimated value mismatch")
	assert.Equal(t, inputInstrument2.Condition, createdInstrument2.Condition, "condition mismatch")
	assert.Equal(t, inputInstrument2.Description, createdInstrument2.Description, "description mismatch")
	assert.Equal(t, inputInstrument2.FamousOwners, createdInstrument2.FamousOwners, "famous users mismatch")
	assert.Equal(t, respRegisterBody["user"].ID, createdInstrument2.OwnerUserID, "owner instrumetn id mismatch")
	assert.NotEmpty(t, createdInstrument2.Version, "version should not be empty")

	// ***************************************************************************
	// 5. Get the first instrument for the user, on the happy path.
	expectedStatusCode = http.StatusOK
	path = fmt.Sprintf("%s/v1/instruments/%d", suite.ts.URL, createdInstrument1.ID)
	headers = make(map[string]string)
	headers["Authorization"] = strings.Join([]string{"Bearer", accessToken}, " ")
	resp = testhelpers.DoTestAPICall(t, "GET", path, nil, headers)
	assert.Equal(t, expectedStatusCode, resp.StatusCode, "status code mismatch")
	respInstrumentBody3 := testhelpers.GetResponseBody[map[string]*data.Instrument](t, resp)
	retrievedInstrument, ok := respInstrumentBody3["instrument"]
	assert.True(t, ok, "the newly created instrument should be sent")
	assert.Equal(t, retrievedInstrument, createdInstrument1, "instrument mismatch")

	// ***************************************************************************
	// 6. Get all instruments for the user, on the happy path.
	type ResponseType struct {
		Instruments []*data.Instrument `json:"instruments"`
		Metadata    data.MetaData      `json:"metadata"`
	}

	expectedMetadata := data.MetaData{
		CurrentPage:  1,
		PageSize:     20,
		FirstPage:    1,
		LastPage:     1,
		TotalRecords: 2,
	}

	expectedStatusCode = http.StatusOK
	path = fmt.Sprintf("%s/v1/instruments", suite.ts.URL)
	headers = make(map[string]string)
	headers["Authorization"] = strings.Join([]string{"Bearer", accessToken}, " ")
	resp = testhelpers.DoTestAPICall(t, "GET", path, nil, headers)
	assert.Equal(t, expectedStatusCode, resp.StatusCode, "status code mismatch")
	respInstrumentsBody := testhelpers.GetResponseBody[ResponseType](t, resp)
	assert.Equal(t, 2, len(respInstrumentsBody.Instruments), "response should contain 2 instruments")
	assert.Equal(t, []*data.Instrument{createdInstrument1, createdInstrument2}, respInstrumentsBody.Instruments, "instruments mismatch")
	assert.Equal(t, expectedMetadata, respInstrumentsBody.Metadata)

}

// TestMainTestSuite runs the MainTestSuite related tests.
func TestMainTestSuite(t *testing.T) {
	suite.Run(t, new(MainTestSuite))
}
