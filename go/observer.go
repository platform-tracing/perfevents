package perfevents

import (
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

type Observer struct {
	name string
}

// New observer creates a new observer
func NewObserver() *Observer {
	return &Observer{name: "perfevent_observer"}
}

// OnStartSpan creates a new Observer for the span
func (o *Observer) OnStartSpan(sp opentracing.Span, operationName string, options opentracing.StartSpanOptions) zipkin.SpanObserver {
	return NewSpanObserver(sp, options)
}

// SpanObserver collects perfevent metrics
type SpanObserver struct {
	sp opentracing.Span
	EventDescs []PerfEventInfo
}

// NewSpanObserver creates a new SpanObserver that can emit perfevent
// metrics
func NewSpanObserver(s opentracing.Span, opts opentracing.StartSpanOptions) *SpanObserver {
	so := &SpanObserver{
		sp: s,
	}

	for k, v := range opts.Tags {
		if k == "perfevents" {
			so.OnSetTag(k, v)
		}
	}

	return so
}

func (so *SpanObserver) OnSetOperationName(operationName string) {
}

func (so *SpanObserver) OnSetTag(key string, value interface{}) {
	if key == string(ext.PerfEvent) {
		if v, ok := value.(string); ok {
			_, _, so.EventDescs = InitOpenEventsEnableSelf(v)
		}
	}
}

func (so *SpanObserver) OnFinish(options opentracing.FinishOptions) {
	// log and close the perf events first, if any, since, we don't
	// want to account for the code to finish up the span.
	EventsRead(so.EventDescs)
	for _, event := range so.EventDescs {
		// In any case of an error for an event, event.EventName
		// will contain "" for an event.
		if event.EventName != "" {
			so.sp.LogEvent(event.EventName + ":" +
				FormatDataToString(event))
		}
	}

	EventsDisableClose(so.EventDescs)
}
