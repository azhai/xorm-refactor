package dialect

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gomodule/redigo/redis"
	"github.com/k0kubun/pp"
)

const REDIS_DEFAULT_PORT uint16 = 6379

type Redis struct {
	addr    string
	options []redis.DialOption
	Values  url.Values
}

func NewRedis(addr, opts string) (*Redis, error) {
	r := &Redis{addr: addr}
	err := r.ParseOptions(opts)
	return r, err
}

func (Redis) Name() string {
	return "redis"
}

func (Redis) ImporterPath() string {
	return "github.com/gomodule/redigo/redis"
}

func (Redis) QuoteIdent(ident string) string {
	return WrapWith(ident, "'", "'")
}

func (r *Redis) ParseDSN(params ConnParams) string {
	r.Values = url.Values{}
	r.addr = params.GetAddr("127.0.0.1", REDIS_DEFAULT_PORT)
	if params.Password != "" {
		r.options = append(r.options, redis.DialPassword(params.Password))
		r.Values.Set("auth", params.Password)
	}
	if dbno, err := strconv.Atoi(params.Database); err == nil {
		r.options = append(r.options, redis.DialDatabase(dbno))
		r.Values.Set("select", params.Database)
	}
	return r.addr
}

func (r *Redis) ParseOptions(opts string) (err error) {
	r.options = make([]redis.DialOption, 0)
	if opts = strings.TrimSpace(opts); opts == "" {
		return
	}
	r.Values, err = url.ParseQuery(opts)
	if err != nil {
		return
	}
	if val := r.Values.Get("auth"); val != "" {
		r.options = append(r.options, redis.DialPassword(val))
	}
	if val := r.Values.Get("select"); val != "" {
		if dbno, err := strconv.Atoi(val); err == nil {
			r.options = append(r.options, redis.DialDatabase(dbno))
		}
	}
	return
}

func (r *Redis) Connect(verbose bool) (redis.Conn, error) {
	if r.addr == "" {
		return nil, fmt.Errorf("the address of redis server is empty")
	}
	if verbose {
		pp.Printf("Connect: %s %s %s\n", r.Name(), r.addr, r.Values.Encode())
	}
	return redis.Dial("tcp", r.addr, r.options...)
}
