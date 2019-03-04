package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/ghodss/yaml"
	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"

	"github.com/openshift/openshift-azure/pkg/util/log"
	"github.com/openshift/openshift-azure/pkg/util/statsd"
)

/*
curl -Gks \
  -H "Authorization: Bearer $(oc serviceaccounts get-token -n openshift-monitoring prometheus-k8s)" \
  --data-urlencode 'match[]={__name__=~".+"}' \
  https://prometheus-k8s.openshift-monitoring.svc:9091/federate
*/

var (
	logLevel  = flag.String("loglevel", "Debug", "Valid values are Debug, Info, Warning, Error")
	gitCommit = "unknown"
)

type authorizingRoundTripper struct {
	http.RoundTripper
	token string
}

func (rt authorizingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+rt.token)
	return rt.RoundTripper.RoundTrip(req)
}

type config struct {
	Interval           time.Duration `json:"intervalNanoseconds,omitempty"`
	PrometheusEndpoint string        `json:"prometheusEndpoint,omitempty"`
	StatsdSocket       string        `json:"statsdSocket,omitempty"`

	Queries []struct {
		Name          string `json:"name,omitempty"`
		Query         string `json:"query,omitempty"`
		CalculateRate bool   `json:"calculate_rate,omitempty"`
	} `json:"queries,omitempty"`

	Account   string `json:"account,omitempty"`
	Namespace string `json:"namespace,omitempty"`

	Region            string `json:"region,omitempty"`
	SubscriptionID    string `json:"subscriptionId,omitempty"`
	ResourceGroupName string `json:"resourceGroupName,omitempty"`
	ResourceName      string `json:"resourceName,omitempty"`

	Token              string `json:"token,omitempty"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify,omitempty"`

	log        *logrus.Entry
	rootCAs    *x509.CertPool
	prometheus v1.API
	rt         http.RoundTripper
	conn       net.Conn
}

func (c *config) load(path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(b, &c); err != nil {
		return err
	}

	b, err = ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt")
	switch {
	case os.IsNotExist(err):
	case err != nil:
		return err
	default:
		c.rootCAs = x509.NewCertPool()
		c.rootCAs.AppendCertsFromPEM(b)
	}

	b, err = ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	switch {
	case os.IsNotExist(err):
	case err != nil:
		return err
	default:
		c.Token = string(b)
	}

	return nil
}

func (c *config) defaultAndValidate() (errs []error) {
	if c.Interval == 0 {
		c.Interval = time.Minute
	}

	if c.Interval < time.Second {
		errs = append(errs, fmt.Errorf("intervalNanoseconds %q too small", int64(c.Interval)))
	}
	if _, err := url.Parse(c.PrometheusEndpoint); err != nil {
		errs = append(errs, fmt.Errorf("prometheusEndpoint: %s", err))
	}
	if _, err := net.ResolveUnixAddr("unix", c.StatsdSocket); err != nil {
		errs = append(errs, fmt.Errorf("statsdSocket: %s", err))
	}
	if len(c.Queries) == 0 {
		errs = append(errs, fmt.Errorf("must configure at least one query"))
	}

	return
}

func (c *config) init() error {
	for {
		var err error
		c.conn, err = net.Dial("unix", c.StatsdSocket)
		if err == nil {
			break
		}
		if err, ok := err.(*net.OpError); ok {
			if err, ok := err.Err.(*os.SyscallError); ok {
				if err.Err == syscall.ENOENT {
					c.log.Warn("socket not found, sleeping...")
					time.Sleep(5 * time.Second)
					continue
				}
			}
		}
		return err
	}

	cli, err := api.NewClient(api.Config{
		Address: c.PrometheusEndpoint,
		RoundTripper: &authorizingRoundTripper{
			RoundTripper: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs:            c.rootCAs,
					InsecureSkipVerify: c.InsecureSkipVerify,
				},
			},
			token: c.Token,
		},
	})
	if err != nil {
		return err
	}

	c.prometheus = v1.NewAPI(cli)

	return nil
}

func run(log *logrus.Entry, configpath string) error {
	c := &config{log: log}

	if err := c.load(configpath); err != nil {
		return err
	}

	if errs := c.defaultAndValidate(); len(errs) > 0 {
		var sb strings.Builder
		for _, err := range errs {
			sb.WriteString(err.Error())
			sb.WriteByte('\n')
		}
		return errors.New(sb.String())
	}

	if c.Interval != time.Minute {
		log.Warnf("intervalNanoseconds is set to %q.  It must be set to %q in production", int64(c.Interval), int64(time.Minute))
	}

	if err := c.init(); err != nil {
		return err
	}
	defer c.conn.Close()

	return c.run()
}

func (c *config) run() error {
	t := time.NewTicker(c.Interval)
	defer t.Stop()

	for {
		if err := c.runOnce(context.Background()); err != nil {
			c.log.Warn(err)
		}
		<-t.C
	}
}

func (c *config) runOnce(ctx context.Context) error {
	var metricsCount int

	for _, query := range c.Queries {
		var prometheusQuery string
		if query.CalculateRate {
			//query for rate of change
			prometheusQuery = fmt.Sprintf("(%s - %s offset 1m) or (%s unless %s offset 1m)", query.Query, query.Query, query.Query, query.Query)
		} else {
			prometheusQuery = query.Query
		}
		value, err := c.prometheus.Query(ctx, prometheusQuery, time.Time{})
		if err != nil {
			return err
		}

		for _, sample := range value.(model.Vector) {
			f := &statsd.Float{
				Metric:    string(sample.Metric[model.MetricNameLabel]),
				Account:   c.Account,
				Namespace: c.Namespace,
				Dims:      map[string]string{},
				TS:        sample.Timestamp.Time(),
				Value:     float64(sample.Value),
			}
			if query.Name != "" {
				f.Metric = query.Name
			}
			for k, v := range sample.Metric {
				if k != model.MetricNameLabel {
					f.Dims[string(k)] = string(v)
				}
			}
			if c.Region != "" {
				f.Dims["region"] = c.Region
			}
			if c.SubscriptionID != "" {
				f.Dims["subscriptionId"] = c.SubscriptionID
			}
			if c.ResourceGroupName != "" {
				f.Dims["resourceGroupName"] = c.ResourceGroupName
			}
			if c.ResourceName != "" {
				f.Dims["resourceName"] = c.ResourceName
			}
			b, err := f.Marshal()
			if err != nil {
				return err
			}
			if _, err = c.conn.Write(b); err != nil {
				return err
			}

			metricsCount++
		}
	}

	c.log.Infof("sent %d metrics", metricsCount)

	return nil
}

func main() {
	flag.Parse()
	logrus.SetLevel(log.SanitizeLogLevel(*logLevel))
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	log := logrus.NewEntry(logrus.StandardLogger())
	log.Printf("metricsbridge starting, git commit %s", gitCommit)

	if len(os.Args) != 2 {
		log.Fatalf("usage: %s config.yaml", os.Args[0])
	}

	if err := run(log, os.Args[1]); err != nil {
		log.Fatal(err)
	}
}
