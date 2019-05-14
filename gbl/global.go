// global instance
package gbl

import (
	"io"

	"github.com/RichardKnop/machinery/v1"

	"github.com/go-redis/redis"

	"github.com/jinzhu/gorm"
	"github.com/opentracing/opentracing-go"
)

// mysql
var DB *gorm.DB

// redis
//var Redis *redis.ClusterClient
var Redis *redis.Client // for test

// tracer
var Tracer opentracing.Tracer
var TracerCloser io.Closer

// asyncjob server
var MachineryServer *machinery.Server
