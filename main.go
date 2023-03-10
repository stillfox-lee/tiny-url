package main

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.elastic.co/apm/module/apmgin/v2"
	"gorm.io/gorm"
)

const (
	maxRange = 100000
	idKey    = "range:100000-1000000"
)

type Base62Coder interface {
	// Encode base-10 64 bit binary number into base-62 string
	Encode(uint64) (string, error)
	// Decode base-62 string into base-10
	Decode(string) (uint64, error)
}

type UrlShortener interface {
	// ShortenUrl should return a short url. metadata contains some business information.
	// It also need to be idempotent, which means if the same url passed in, it should return
	// the same short url. But metadata should be updated.
	ShortenUrl(ctx context.Context, url string, metadata map[string]string) (string, error)
	RedirectUrl(ctx context.Context, shortUrl string) (string, error)
}

// GetUniqueID will connect to redis an use redis incr command to get a unique id.
// redis key will be like "short:max:100000", "short:max:200000", etc.
func GetUniqueID(wide string) (uint64, error) {
	return 0, nil
}

type ShortenRequest struct {
	OriginUrl string            `json:"origin_url"`
	Metadata  map[string]string `json:"metadata"`
}

type ShortenResponse struct {
	ShortUrl string `json:"short_url"`
}

type service struct {
	db    *gorm.DB
	rdb   *redis.Client
	coder Base62Coder
}

func (s *service) ShortenHandler(c *gin.Context) {
	// get OriginUrl from json body

	shortenRequest := &ShortenRequest{}
	if err := c.ShouldBindJSON(shortenRequest); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	// get unique id from redis
	idS, err := s.rdb.Get(c, idKey).Result()
	if err == redis.Nil {
		c.JSON(500, gin.H{"error": "redis ID range Key init fail"})
		return
	}
	id, err := strconv.ParseUint(idS, 10, 64)
	if err != nil {
		c.JSON(500, gin.H{"error": "redis ID covert fail"})
		return
	}

	shortUrl, err := s.coder.Encode(id)
	if err != nil {
		c.JSON(500, gin.H{"error": "base62 encode fail"})
		return
	}

	// insert into database
	s.db.WithContext(c).Create(&UrlEntity{
		OriginUrl: shortenRequest.OriginUrl,
		ShortUrl:  shortUrl,
		ID:        uint64(id),
	})

	c.JSON(200, ShortenResponse{ShortUrl: shortUrl})
}

type RedirectRequest struct {
	ShortUrl string `json:"short_url"`
}

type RedirectResponse struct {
	OriginUrl string `json:"origin_url"`
}

func (s *service) RedirectHandler(c *gin.Context) {

	request := &RedirectRequest{}
	if err := c.ShouldBindQuery(request); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	id, err := s.coder.Decode(request.ShortUrl)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	originUrl, err := s.rdb.Get(c, request.ShortUrl).Result()

	// if url not in redis, try to get from database
	if err != nil {
		en := &UrlEntity{}
		err = s.db.WithContext(c).Where("id = ?", id).First(en).Error
		// url not exist
		if err != nil {
			c.JSON(404, gin.H{"error": err.Error()})
			return
		}
		// update redis
		s.rdb.Set(c, request.ShortUrl, originUrl, 0)
		c.JSON(200, RedirectResponse{OriginUrl: originUrl})
		return
	}
	c.JSON(200, RedirectResponse{OriginUrl: originUrl})

}

// TODO: redis cache metrics
func main() {
	// use gin framework create a web server, add a route "/url/shorten" and "/url/redirect" to handle the request, and server should listen on port 8080

	coder := NewCoder()
	rdb, err := InitRedisClient("localhost", "6379", "", 0)
	if err != nil {
		panic("failed to connect redis")
	}

	db, err := InitDB("localhost", "dpuser", "1111", "urls", 3306)
	if err != nil {
		panic("failed to connect database")
	}

	svc := &service{
		db:    db,
		rdb:   rdb,
		coder: coder,
	}
	// create gin server
	r := gin.New()
	r.Use(apmgin.Middleware(r))
	r.POST("/url/shorten", svc.ShortenHandler)
	r.GET("/url/redirect", svc.RedirectHandler)
	// web server listen on port 8080

	r.Run(":8080")
}
