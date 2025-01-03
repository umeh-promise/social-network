package main

import (
	"expvar"
	"runtime"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/umeh-promise/social/internal/auth"
	"github.com/umeh-promise/social/internal/db"
	"github.com/umeh-promise/social/internal/env"
	"github.com/umeh-promise/social/internal/mailer"
	"github.com/umeh-promise/social/internal/ratelimiter"
	"github.com/umeh-promise/social/internal/store"
	"github.com/umeh-promise/social/internal/store/cache"
	"go.uber.org/zap"
)

const version = "1.1.0"

//	@title			Social API
//	@description	API for Social, a social network api
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath					/v1
//
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description

func main() {
	config := config{
		addr:        env.GetString("ADDR", ":8080"),
		apiURL:      env.GetString("EXTERNAL_URL", "localhost:8080"),
		frontendURL: env.GetString("FRONTEND_URL", "localhost:4000"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://user:password@localhost:5432/social?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		cache: cacheConfig{
			addr:    env.GetString("REDIS_ADDR", "localhost:6379"),
			pwd:     env.GetString("REDIS_PWD", ""),
			db:      env.GetInt("REDIS_DB", 0),
			enabled: env.GetBool("REDIS_ENABLED", false),
		},
		env: env.GetString("ENV", "development"),
		mail: mailConfig{
			exp:       time.Hour * 24 * 3, // 3 days
			fromEmail: env.GetString("FROM_EMAIL", ""),
			sendGrid: sendGridConfig{
				apikey: env.GetString("SENDGRID_API_KEY", ""),
			},
		},

		auth: authConfig{
			basic: basicConfig{
				username: env.GetString("AUTH_BASIC_USER", "admin"),
				password: env.GetString("AUTH_BASIC_PASSWORD", "admin"),
			},
			token: tokenConfig{
				secret: env.GetString("AUTH_SECRET", "basic"),
				exp:    time.Hour * 24 * 3,
				issuer: "social-network",
			},
		},
		rateLimiter: ratelimiter.Config{
			RequestPerTimeFrame: env.GetInt("RATELIMITER_REQUEST_COUNT", 20),
			TimeFrame:           time.Second * 5,
			Enabled:             env.GetBool("RATELIMITER_ENABLED", true),
		},
	}

	// Rate limiter
	rateLimiter := ratelimiter.NewFixedWindowLimiter(
		config.rateLimiter.RequestPerTimeFrame, config.rateLimiter.TimeFrame,
	)

	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	// Database
	db, err := db.New(
		config.db.addr,
		config.db.maxOpenConns,
		config.db.maxIdleConns,
		config.db.maxIdleTime,
	)

	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()
	logger.Info("DB connected successfully")

	// cache storage
	var rdb *redis.Client
	if config.cache.enabled {
		rdb = cache.NewCacheClient(config.cache.addr, config.cache.pwd, config.cache.db)
		logger.Info("redis cache connection pool is established")
	}

	store := store.NewStore(db)
	cacheStorage := cache.NewCacheStorage(rdb)

	mailer := mailer.NewSendgrid(config.mail.sendGrid.apikey, config.mail.fromEmail)

	jwtAuthenticator := auth.NewJWTAuthenticator(config.auth.token.secret, config.auth.token.issuer, config.auth.token.issuer)

	app := &application{
		config:        config,
		store:         store,
		cacheStorage:  cacheStorage,
		logger:        logger,
		mailer:        mailer,
		authenticator: jwtAuthenticator,
		rateLimiter:   rateLimiter,
	}

	expvar.NewString("version").Set(version)
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	router := app.mount()

	logger.Fatal(app.run(router))
}
