/*
Copyright 2024 Kubeworkz Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package backend

import (
	v1 "audit/pkg/backend/v1"
	"audit/pkg/utils/env"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"time"

	"github.com/saashqdev/kubeworkz/pkg/clog"
)

var cacheCh = make(chan *v1.Event, DefaultCacheCapacity)

const (
	SendTimeout          = time.Second * 3
	DefaultSendersNum    = 100
	DefaultBatchSize     = 100
	DefaultBatchInterval = time.Second * 3
	DefaultCacheCapacity = 10000
	CacheTimeout         = time.Second * 3
	MaxRetryTime         = 10
)

var (
	SendElasticSearch bool
)

type Backend struct {
	url                string
	client             http.Client
	sendTimeout        time.Duration
	getSenderTimeout   time.Duration
	senderCh           chan interface{}
	eventBatchInterval time.Duration
	cache              chan *v1.Event
	stopCh             <-chan struct{}
	eventBatchSize     int
}

func NewBackend() *Backend {

	esWebhook := env.ElasticSearchHost()
	esUrl := esWebhook.Host + "/" + esWebhook.Index + "/" + esWebhook.Type
	b := Backend{
		url:                esUrl,
		sendTimeout:        SendTimeout,
		eventBatchInterval: DefaultBatchInterval,
		eventBatchSize:     DefaultBatchSize,
		getSenderTimeout:   SendTimeout,
	}

	b.senderCh = make(chan interface{}, DefaultSendersNum)
	b.cache = cacheCh

	b.client = http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: b.sendTimeout,
	}

	return &b
}

// send event to es/webhook
func (b *Backend) sendEvents(events *v1.EventList) {

	ctx, cancel := context.WithTimeout(context.Background(), b.sendTimeout)
	defer cancel()

	stopCh := make(chan struct{})

	send := func() {
		ctx, cancel := context.WithTimeout(context.Background(), b.getSenderTimeout)
		defer cancel()

		select {
		case <-ctx.Done():
			clog.Info("get auditing event sender timeout")
			return
		case b.senderCh <- struct{}{}:
		}

		start := time.Now()
		defer func() {
			stopCh <- struct{}{}
			clog.Info("send %d auditing logs used %d", len(events.Items), time.Since(start).Milliseconds())
		}()
		retry := 0
		for _, event := range events.Items {

			bs, err := json.Marshal(event)
			if err != nil {
				clog.Error("json marshal error, %s", err)
				return
			}

			response, err := b.client.Post(b.url, "application/json", bytes.NewBuffer(bs))
			if err != nil {
				clog.Error("send audit event error, %s", err)
				retry++
				if retry >= MaxRetryTime {
					b.dealFailSend()
				}
				return
			}

			if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated {
				clog.Error("send audit event error[%d]", response.StatusCode)
				return
			}
			clog.Debug("send event %s success", event.EventName)
		}

	}

	go send()

	defer func() {
		<-b.senderCh
	}()

	select {
	case <-ctx.Done():
		clog.Info("send audit events timeout")
	case <-stopCh:
	}
}

func (b *Backend) dealFailSend() {
	clog.Info("send es fail exceed max retry times")
	b.senderCh = make(chan interface{}, DefaultSendersNum)
	b.cache = make(chan *v1.Event, DefaultCacheCapacity)
}

// get event from backend cache channel
func (b *Backend) getEvents() *v1.EventList {

	ctx, cancel := context.WithTimeout(context.Background(), b.eventBatchInterval)
	defer cancel()

	events := &v1.EventList{}
	for {
		select {
		case event := <-b.cache:
			if event == nil {
				break
			}
			events.Items = append(events.Items, *event)
			if len(events.Items) >= b.eventBatchSize {
				return events
			}
		case <-ctx.Done():
			return events
		case <-b.stopCh:
			return nil
		}
	}
}

func GetCacheCh() chan *v1.Event {
	return cacheCh
}

func (b *Backend) Run() {
	for {
		events := b.getEvents()
		if events == nil {
			break
		}

		if len(events.Items) == 0 {
			continue
		}
		if SendElasticSearch {
			go b.sendEvents(events)
		}
	}
}

// send event to cache channel
func CacheEvent(ch chan *v1.Event, e *v1.Event) {
	select {
	case ch <- e:
		return
	case <-time.After(CacheTimeout):
		clog.Info("cache audit event %s timeout", e.RequestId)
		break
	}
}
