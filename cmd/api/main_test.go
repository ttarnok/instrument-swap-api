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

	cfg.env = "teszt"
	cfg.db.dsn = suite.pgContainer.ConnectionString
	cfg.db.maxIdleConns = 25
	cfg.db.maxOpenConns = 25
	cfg.db.maxIdleTime = 15 * time.Minute
	cfg.limiter.requestPerSecond = 2
	cfg.limiter.burst = 4
	cfg.limiter.enabled = true
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
	expectedRespBody["liveliness"]["environment"] = "teszt"
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
// 2. User login, happy path.
// 3. Create an instrument.
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
	// 2. User login, happy path.
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
	// 3. Create an instrument.
	expectedStatusCode = http.StatusCreated
	inputInstrument := struct {
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
	resp = testhelpers.DoTestAPICall(t, "POST", path, CreateRequestBody(t, inputInstrument), headers)
	assert.Equal(t, expectedStatusCode, resp.StatusCode, "status code mismatch")
	respInstrumentBody := testhelpers.GetResponseBody[map[string]*data.Instrument](t, resp)
	createdInstrument, ok := respInstrumentBody["instrument"]

	assert.True(t, ok, "the newly created instrument should be sent")
	assert.NotEmpty(t, createdInstrument.ID, "the newly created instrumnent id shoud not be empty or 0")
	assert.NotEmpty(t, createdInstrument.CreatedAt, "created at date should not be empty")
	assert.Equal(t, inputInstrument.Name, createdInstrument.Name, "instrument name mismatch")
	assert.Equal(t, inputInstrument.Manufacturer, createdInstrument.Manufacturer, "instrument manufacturer mismatch")
	assert.Equal(t, inputInstrument.ManufactureYear, createdInstrument.ManufactureYear, "instrument manufacture year mismatch")
	assert.Equal(t, inputInstrument.Type, createdInstrument.Type, "instrument type mismatch")
	assert.Equal(t, inputInstrument.FamousOwners, createdInstrument.FamousOwners, "famous users mismatch")
	assert.Equal(t, respRegisterBody["user"].ID, createdInstrument.OwnerUserID, "owner instrumetn id mismatch")
	assert.NotEmpty(t, createdInstrument.Version, "version should not be empty")

}

// TestMainTestSuite runs the MainTestSuite related tests.
func TestMainTestSuite(t *testing.T) {
	suite.Run(t, new(MainTestSuite))
}
