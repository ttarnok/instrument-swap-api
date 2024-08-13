package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http/httptest"
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

	resp := testhelpers.DoTestAPICall(t, "GET", path, nil)

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
// 1. User registration.
func (suite *MainTestSuite) TestBasicUserStory() {
	t := suite.T()
	// ***************************************************************************
	// User registration.
	expectedStatusCode := 201

	input := struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Name:     "John Smith",
		Email:    "johnsmith@example.com",
		Password: "asd123asd1234",
	}

	path := fmt.Sprintf("%s/v1/users", suite.ts.URL)
	resp := testhelpers.DoTestAPICall(t, "POST", path, CreateRequestBody(t, input))
	assert.Equal(t, expectedStatusCode, resp.StatusCode, "status code mismatch")
	respBody := testhelpers.GetResponseBody[map[string]*data.User](t, resp)
	assert.NotEqual(t, respBody["user"].ID, "user ID should not be 0")
	assert.Equal(t, input.Name, respBody["user"].Name, "user name mismatch")
	assert.Equal(t, input.Email, respBody["user"].Email, "email mismatch")
}

// TestMainTestSuite runs the MainTestSuite related tests.
func TestMainTestSuite(t *testing.T) {
	suite.Run(t, new(MainTestSuite))
}
