package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/stretchr/testify/suite"
	"github.com/ttarnok/instrument-swap-api/internal/auth"
	"github.com/ttarnok/instrument-swap-api/internal/data"
	"github.com/ttarnok/instrument-swap-api/internal/testhelpers"
)

type MainTestSuite struct {
	suite.Suite
	pgContainer    *testhelpers.PostgresContainer
	redisContainer *testhelpers.RedisContainer
	ts             *httptest.Server
	app            *application
	ctx            context.Context
}

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

func (suite *MainTestSuite) TearDownSuite() {
	defer suite.ts.Close()
	if err := suite.pgContainer.Terminate(suite.ctx); err != nil {
		log.Fatalf("error terminating postgres container: %s", err)
	}
	if err := suite.redisContainer.Terminate(suite.ctx); err != nil {
		log.Fatalf("error terminating redis container: %s", err)
	}
}

func (suite *MainTestSuite) TestHealthCheck() {
	t := suite.T()

	path := fmt.Sprintf("%s/v1/liveliness", suite.ts.URL)

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		t.Fatal(err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Error("expected status code 200, got", resp.StatusCode)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	respEnvelope := make(map[string]map[string]string)
	err = json.Unmarshal(body, &respEnvelope)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(respEnvelope)

}

func TestMainTestSuite(t *testing.T) {
	suite.Run(t, new(MainTestSuite))
}
