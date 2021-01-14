package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/url"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	config_util "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/pkg/value"
	"github.com/prometheus/prometheus/prompb"
	"github.com/prometheus/prometheus/storage/remote"
)

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

var flagURL = flag.String("url", "http://cortex:9009/api/v1/push", "remote write url")
var flagSendInterval = flag.Duration("send-interval", 10*time.Second, "interval how often series is remote written")

func valOrStaleNan(prob float64, val float64) float64 {
	if random.Float64() < prob {
		return math.Float64frombits(value.StaleNaN)
	}
	return val
}

func run() error {

	var nodes = []struct {
		name          string
		nanProbabilty float64
		available     int64
		total         int64
	}{
		{
			name:          "node1",
			available:     32,
			total:         128,
			nanProbabilty: 0.01,
		},
		{
			name:          "node2",
			available:     64,
			total:         128,
			nanProbabilty: 0.1,
		},
		{
			name:          "node3",
			available:     96,
			total:         128,
			nanProbabilty: 0,
		},
	}

	remoteWriteURL, err := url.Parse(*flagURL)
	if err != nil {
		return err
	}

	client, err := remote.NewWriteClient("remote-write", &remote.ClientConfig{
		URL: &config_util.URL{
			URL: remoteWriteURL,
		},
		Timeout: model.Duration(30 * time.Second),
	})
	if err != nil {
		return err
	}

	tick := time.NewTicker(*flagSendInterval)

	for {

		var req prompb.WriteRequest

		for _, n := range nodes {
			req.Timeseries = append(req.Timeseries,
				prompb.TimeSeries{
					Labels: []prompb.Label{
						{
							Name:  "__name__",
							Value: "nan_test_metric_total",
						},
						{
							Name:  "instance",
							Value: n.name,
						},
					},
					Samples: []prompb.Sample{
						{
							Value:     valOrStaleNan(n.nanProbabilty, float64(n.total)*1024.0*1024.0),
							Timestamp: int64(model.TimeFromUnixNano(time.Now().UnixNano())),
						},
					},
				},
				prompb.TimeSeries{
					Labels: []prompb.Label{
						{
							Name:  "__name__",
							Value: "nan_test_metric_available",
						},
						{
							Name:  "instance",
							Value: n.name,
						},
					},
					Samples: []prompb.Sample{
						{
							Value:     valOrStaleNan(n.nanProbabilty, float64(n.available)*1024.0*1024.0),
							Timestamp: int64(model.TimeFromUnixNano(time.Now().UnixNano())),
						},
					},
				})
		}

		data, err := proto.Marshal(&req)
		if err != nil {
			return err
		}

		compressed := snappy.Encode(nil, data)

		ctx, cancel := context.WithTimeout(context.TODO(), 20*time.Second)
		defer cancel()

		if err := client.Store(ctx, compressed); err != nil {
			return fmt.Errorf("error storing: %w", err)
		}
		log.Print("pushed metrics")

		<-tick.C

	}

	return nil
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
