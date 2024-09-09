package delegation

// Code generated by github.com/launchdarkly/go-options.  DO NOT EDIT.

import "fmt"

import (
	"github.com/ucan-wg/go-ucan/did"
	"time"
)

type ApplyOptionFunc func(c *config) error

func (f ApplyOptionFunc) apply(c *config) error {
	return f(c)
}

func newConfig(options ...Option) (config, error) {
	var c config
	err := applyConfigOptions(&c, options...)
	return c, err
}

func applyConfigOptions(c *config, options ...Option) error {
	for _, o := range options {
		if err := o.apply(c); err != nil {
			return err
		}
	}
	return nil
}

type Option interface {
	apply(*config) error
}

type withExpirationImpl struct {
	o *time.Time
}

func (o withExpirationImpl) apply(c *config) error {
	c.Expiration = o.o
	return nil
}

func (o withExpirationImpl) String() string {
	name := "WithExpiration"

	// hack to avoid go vet error about passing a function to Sprintf
	var value interface{} = o.o
	return fmt.Sprintf("%s: %+v", name, value)
}

func WithExpiration(o *time.Time) Option {
	return withExpirationImpl{
		o: o,
	}
}

type withMetaImpl struct {
	o map[string]any
}

func (o withMetaImpl) apply(c *config) error {
	c.Meta = o.o
	return nil
}

func (o withMetaImpl) String() string {
	name := "WithMeta"

	// hack to avoid go vet error about passing a function to Sprintf
	var value interface{} = o.o
	return fmt.Sprintf("%s: %+v", name, value)
}

func WithMeta(o map[string]any) Option {
	return withMetaImpl{
		o: o,
	}
}

type withNoExpirationImpl struct {
	o bool
}

func (o withNoExpirationImpl) apply(c *config) error {
	c.NoExpiration = o.o
	return nil
}

func (o withNoExpirationImpl) String() string {
	name := "WithNoExpiration"

	// hack to avoid go vet error about passing a function to Sprintf
	var value interface{} = o.o
	return fmt.Sprintf("%s: %+v", name, value)
}

func WithNoExpiration(o bool) Option {
	return withNoExpirationImpl{
		o: o,
	}
}

type withNotBeforeImpl struct {
	o *time.Time
}

func (o withNotBeforeImpl) apply(c *config) error {
	c.NotBefore = o.o
	return nil
}

func (o withNotBeforeImpl) String() string {
	name := "WithNotBefore"

	// hack to avoid go vet error about passing a function to Sprintf
	var value interface{} = o.o
	return fmt.Sprintf("%s: %+v", name, value)
}

func WithNotBefore(o *time.Time) Option {
	return withNotBeforeImpl{
		o: o,
	}
}

type withSubjectImpl struct {
	o *did.DID
}

func (o withSubjectImpl) apply(c *config) error {
	c.Subject = o.o
	return nil
}

func (o withSubjectImpl) String() string {
	name := "WithSubject"

	// hack to avoid go vet error about passing a function to Sprintf
	var value interface{} = o.o
	return fmt.Sprintf("%s: %+v", name, value)
}

// WithSubject is a did.DID representing the Subject.
func WithSubject(o *did.DID) Option {
	return withSubjectImpl{
		o: o,
	}
}

type withPowerlineImpl struct {
	o bool
}

func (o withPowerlineImpl) apply(c *config) error {
	c.Powerline = o.o
	return nil
}

func (o withPowerlineImpl) String() string {
	name := "WithPowerline"

	// hack to avoid go vet error about passing a function to Sprintf
	var value interface{} = o.o
	return fmt.Sprintf("%s: %+v", name, value)
}

func WithPowerline(o bool) Option {
	return withPowerlineImpl{
		o: o,
	}
}
