package http_test

import (
	"github.com/cloudevents/sdk-go/pkg/cloudevents/canonical"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/google/go-cmp/cmp"
	"net/url"
	"testing"
	"time"
)

func TestCodecV02_Encode(t *testing.T) {
	now := canonical.Timestamp{Time: time.Now()}
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &canonical.URLRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &canonical.URLRef{URL: *schemaUrl}

	testCases := map[string]struct {
		codec   http.CodecV02
		event   canonical.Event
		want    *http.Message
		wantErr error
	}{
		"simple v2 default": {
			codec: http.CodecV02{},
			event: canonical.Event{
				Context: canonical.EventContextV02{
					Type:   "com.example.test",
					Source: *source,
					ID:     "ABC-123",
				},
			},
			want: &http.Message{
				Header: map[string][]string{
					"Ce-Specversion": {"0.2"},
					"Ce-Id":          {"ABC-123"},
					"Ce-Type":        {"com.example.test"},
					"Ce-Source":      {"http://example.com/source"},
					"Content-Type":   {"application/json"},
				},
			},
		},
		"full v2 default": {
			codec: http.CodecV02{},
			event: canonical.Event{
				Context: canonical.EventContextV02{
					ID:          "ABC-123",
					Time:        &now,
					Type:        "com.example.test",
					SchemaURL:   schema,
					ContentType: "application/json",
					Source:      *source,
					Extensions: map[string]interface{}{
						"test": "extended",
					},
				},
				Data: map[string]interface{}{
					"hello": "world",
				},
			},
			want: &http.Message{
				Header: map[string][]string{
					"Ce-Specversion": {"0.2"},
					"Ce-Id":          {"ABC-123"},
					"Ce-Time":        {now.Format(time.RFC3339Nano)},
					"Ce-Type":        {"com.example.test"},
					"Ce-Source":      {"http://example.com/source"},
					"Ce-Schemaurl":   {"http://example.com/schema"},
					"Ce-Test":        {`"extended"`},
					"Content-Type":   {"application/json"},
				},
				Body: []byte(`{"hello":"world"}`),
			},
		},
		"simple v2 binary": {
			codec: http.CodecV02{Encoding: http.BinaryV02},
			event: canonical.Event{
				Context: canonical.EventContextV02{
					Type:   "com.example.test",
					Source: *source,
					ID:     "ABC-123",
				},
			},
			want: &http.Message{
				Header: map[string][]string{
					"Ce-Specversion": {"0.2"},
					"Ce-Id":          {"ABC-123"},
					"Ce-Type":        {"com.example.test"},
					"Ce-Source":      {"http://example.com/source"},
					"Content-Type":   {"application/json"},
				},
			},
		},
		"full v2 binary": {
			codec: http.CodecV02{Encoding: http.BinaryV02},
			event: canonical.Event{
				Context: canonical.EventContextV02{
					ID:          "ABC-123",
					Time:        &now,
					Type:        "com.example.test",
					SchemaURL:   schema,
					ContentType: "application/json",
					Source:      *source,
					Extensions: map[string]interface{}{
						"test": "extended",
						"asmap": map[string]interface{}{
							"a": "apple",
							"b": "banana",
							"c": map[string]interface{}{
								"d": "dog",
								"e": "eel",
							},
						},
					},
				},
				Data: map[string]interface{}{
					"hello": "world",
				},
			},
			want: &http.Message{
				Header: map[string][]string{
					"Ce-Specversion": {"0.2"},
					"Ce-Id":          {"ABC-123"},
					"Ce-Time":        {now.Format(time.RFC3339Nano)},
					"Ce-Type":        {"com.example.test"},
					"Ce-Source":      {"http://example.com/source"},
					"Ce-Schemaurl":   {"http://example.com/schema"},
					"Ce-Test":        {`"extended"`},
					"Ce-Asmap-A":     {`"apple"`},
					"Ce-Asmap-B":     {`"banana"`},
					"Ce-Asmap-C":     {`{"d":"dog","e":"eel"}`},
					"Content-Type":   {"application/json"},
				},
				Body: []byte(`{"hello":"world"}`),
			},
		},
		"simple v2 structured": {
			codec: http.CodecV02{Encoding: http.StructuredV02},
			event: canonical.Event{
				Context: canonical.EventContextV02{
					Type:   "com.example.test",
					Source: *source,
					ID:     "ABC-123",
				},
			},
			want: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: func() []byte {
					body := map[string]interface{}{
						"specversion": "0.2",
						"id":          "ABC-123",
						"type":        "com.example.test",
						"source":      "http://example.com/source",
					}
					return toBytes(body)
				}(),
			},
		},
		"full v2 structured": {
			codec: http.CodecV02{Encoding: http.StructuredV02},
			event: canonical.Event{
				Context: canonical.EventContextV02{
					ID:          "ABC-123",
					Time:        &now,
					Type:        "com.example.test",
					SchemaURL:   schema,
					ContentType: "application/json",
					Source:      *source,
					Extensions: map[string]interface{}{
						"test": "extended",
					},
				},
				Data: map[string]interface{}{
					"hello": "world",
				},
			},
			want: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: func() []byte {
					body := map[string]interface{}{
						"specversion": "0.2",
						"contenttype": "application/json",
						"data": map[string]interface{}{
							"hello": "world",
						},
						"id":   "ABC-123",
						"time": now,
						"type": "com.example.test",
						"-": map[string]interface{}{ // TODO: this could be an issue.
							"test": "extended",
						},
						"schemaurl": "http://example.com/schema",
						"source":    "http://example.com/source",
					}
					return toBytes(body)
				}(),
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got, err := tc.codec.Encode(tc.event)

			if tc.wantErr != nil || err != nil {
				if diff := cmp.Diff(tc.wantErr, err); diff != "" {
					t.Errorf("unexpected error (-want, +got) = %v", diff)
				}
				return
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {

				if msg, ok := got.(*http.Message); ok {
					// It is hard to read the byte dump
					want := string(tc.want.Body)
					got := string(msg.Body)
					if diff := cmp.Diff(want, got); diff != "" {
						t.Errorf("unexpected message body (-want, +got) = %v", diff)
						return
					}
				}

				t.Errorf("unexpected message (-want, +got) = %v", diff)
			}
		})
	}
}

// TODO: figure out extensions for v2

func TestCodecV02_Decode(t *testing.T) {
	now := canonical.Timestamp{Time: time.Now()}
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &canonical.URLRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &canonical.URLRef{URL: *schemaUrl}

	testCases := map[string]struct {
		codec   http.CodecV02
		msg     *http.Message
		want    *canonical.Event
		wantErr error
	}{
		"simple v2 binary": {
			codec: http.CodecV02{},
			msg: &http.Message{
				Header: map[string][]string{
					"ce-specversion": {"0.2"},
					"ce-id":          {"ABC-123"},
					"ce-type":        {"com.example.test"},
					"ce-source":      {"http://example.com/source"},
					"Content-Type":   {"application/json"},
				},
			},
			want: &canonical.Event{
				Context: canonical.EventContextV02{
					SpecVersion: canonical.CloudEventsVersionV02,
					ContentType: "application/json",
					Type:        "com.example.test",
					Source:      *source,
					ID:          "ABC-123",
				},
			},
		},
		"full v2 binary": {
			codec: http.CodecV02{},
			msg: &http.Message{
				Header: map[string][]string{
					"ce-specversion": {"0.2"},
					"ce-id":          {"ABC-123"},
					"ce-time":        {now.Format(time.RFC3339Nano)},
					"ce-type":        {"com.example.test"},
					"ce-source":      {"http://example.com/source"},
					"ce-schemaurl":   {"http://example.com/schema"},
					"ce-test":        {`"extended"`},
					"ce-asmap-a":     {`"apple"`},
					"ce-asmap-b":     {`"banana"`},
					"ce-asmap-c":     {`{"d":"dog","e":"eel"}`},
					"Content-Type":   {"application/json"},
				},
				Body: toBytes(map[string]interface{}{
					"hello": "world",
				}),
			},
			want: &canonical.Event{
				Context: canonical.EventContextV02{
					SpecVersion: canonical.CloudEventsVersionV02,
					ID:          "ABC-123",
					Time:        &now,
					Type:        "com.example.test",
					SchemaURL:   schema,
					ContentType: "application/json",
					Source:      *source,
					Extensions: map[string]interface{}{
						"test": "extended",
						"asmap": map[string]interface{}{
							"a": []string{`"apple"`},
							"b": []string{`"banana"`},
							"c": []string{`{"d":"dog","e":"eel"}`},
						},
					},
				},
				Data: toBytes(map[string]interface{}{
					"hello": "world",
				}),
			},
		},
		"simple v2 structured": {
			codec: http.CodecV02{},
			msg: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: toBytes(map[string]interface{}{
					"specversion": "0.2",
					"id":          "ABC-123",
					"type":        "com.example.test",
					"source":      "http://example.com/source",
				}),
			},
			want: &canonical.Event{
				Context: canonical.EventContextV02{
					SpecVersion: canonical.CloudEventsVersionV02,
					Type:        "com.example.test",
					Source:      *source,
					ID:          "ABC-123",
				},
			},
		},
		"full v2 structured": {
			codec: http.CodecV02{},
			msg: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: toBytes(map[string]interface{}{
					"specversion": "0.2",
					"contenttype": "application/json",
					"data": map[string]interface{}{
						"hello": "world",
					},
					"id":   "ABC-123",
					"time": now,
					"type": "com.example.test",
					"extensions": map[string]interface{}{
						"test": "extended",
					},
					"schemaurl": "http://example.com/schema",
					"source":    "http://example.com/source",
				}),
			},
			want: &canonical.Event{
				Context: canonical.EventContextV02{
					SpecVersion: canonical.CloudEventsVersionV02,
					ID:          "ABC-123",
					Time:        &now,
					Type:        "com.example.test",
					SchemaURL:   schema,
					ContentType: "application/json",
					Source:      *source,
					Extensions: map[string]interface{}{
						"test": "extended",
					},
				},
				Data: toBytes(map[string]interface{}{
					"hello": "world",
				}),
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got, err := tc.codec.Decode(tc.msg)

			if tc.wantErr != nil || err != nil {
				if diff := cmp.Diff(tc.wantErr, err); diff != "" {
					t.Errorf("unexpected error (-want, +got) = %v", diff)
				}
				return
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected event (-want, +got) = %v", diff)
			}
		})
	}
}
