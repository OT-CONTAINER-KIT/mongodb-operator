package controllers

import (
	"go.uber.org/zap"
	mdb "mongodb-operator/api/v1alpha1"
	"mongodb-operator/k8sgo/results"
	"mongodb-operator/k8sgo/status"
	"mongodb-operator/k8sgo/type"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// severity indicates the severity level
// at which the message should be logged
type severity string

const (
	Info  severity = "INFO"
	Debug severity = "DEBUG"
	Warn  severity = "WARN"
	Error severity = "ERROR"
	None  severity = "NONE"
)

// optionBuilder is in charge of constructing a slice of options that
// will be applied on top of the MongoDB resource that has been provided
type optionBuilder struct {
	options []status.Option
}

// GetOptions implements the OptionBuilder interface
func (o *optionBuilder) GetOptions() []status.Option {
	return o.options
}

// statusOptions returns an initialized optionBuilder
func statusOptions() *optionBuilder {
	return &optionBuilder{
		options: []status.Option{},
	}
}

func (o *optionBuilder) withVersion(version string) *optionBuilder {
	o.options = append(o.options,
		versionOption{
			version: version,
		})
	return o
}

type versionOption struct {
	version string
}

func (v versionOption) ApplyOption(mdb *mdb.MongoDBCluster) {
	mdb.Status.Version = v.version
}

func (v versionOption) GetResult() (reconcile.Result, error) {
	return results.OK()
}

func (o *optionBuilder) withState(state string, retryAfter int) *optionBuilder {
	o.options = append(o.options,
		stateOption{
			state:      state,
			retryAfter: retryAfter,
		})
	return o
}

type message struct {
	messageString string
	severityLevel severity
}

type messageOption struct {
	message message
}

func (m messageOption) ApplyOption(mdb *mdb.MongoDBCluster) {
	mdb.Status.Message = m.message.messageString
	if m.message.severityLevel == Error {
		zap.S().Error(m.message.messageString)
	}
	if m.message.severityLevel == Warn {
		zap.S().Warn(m.message.messageString)
	}
	if m.message.severityLevel == Info {
		zap.S().Info(m.message.messageString)
	}
	if m.message.severityLevel == Debug {
		zap.S().Debug(m.message.messageString)
	}
}

func (m messageOption) GetResult() (reconcile.Result, error) {
	return results.OK()
}

func (o *optionBuilder) withMessage(severityLevel severity, msg string) *optionBuilder {
	if results.IsTransientMessage(msg) {
		severityLevel = Debug
		msg = ""
	}
	o.options = append(o.options, messageOption{
		message: message{
			messageString: msg,
			severityLevel: severityLevel,
		},
	})
	return o
}
func (o *optionBuilder) withCreatingState(retryAfter int) *optionBuilder {
	return o.withState(_type.Creating, retryAfter)
}

func (o *optionBuilder) withFailedState() *optionBuilder {
	return o.withState(_type.Failed, 0)
}

func (o *optionBuilder) withPendingState(retryAfter int) *optionBuilder {
	return o.withState(_type.Pending, retryAfter)
}

func (o *optionBuilder) withScalingState(retryAfter int) *optionBuilder {
	return o.withState(_type.Scaling, retryAfter)
}

func (o *optionBuilder) withRunningState() *optionBuilder {
	return o.withState(_type.Running, -1)
}

func (o *optionBuilder) withExpandingState(retryAfter int) *optionBuilder {
	return o.withState(_type.Expanding, retryAfter)
}

type stateOption struct {
	state      string
	retryAfter int
}

func (s stateOption) ApplyOption(mdb *mdb.MongoDBCluster) {
	mdb.Status.State = s.state
}

func (s stateOption) GetResult() (reconcile.Result, error) {
	if s.state == _type.Running {
		return results.OK()
	}
	if s.state == _type.Pending {
		return results.Retry(s.retryAfter)
	}
	if s.state == _type.Failed {
		return results.Fail()
	}
	if s.state == _type.Deleting {
		return results.Retry(s.retryAfter)
	}
	return results.OK()
}
