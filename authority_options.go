package ucan

// Code generated by github.com/launchdarkly/go-options.  DO NOT EDIT.

import "fmt"

import (
	"time"
)

type ApplyAuthorityOptionFunc func(c *authorityConfig) error

func (f ApplyAuthorityOptionFunc) apply(c *authorityConfig) error {
	return f(c)
}

func applyAuthorityConfigOptions(c *authorityConfig, options ...AuthorityOption) error {
	for _, o := range options {
		if err := o.apply(c); err != nil {
			return err
		}
	}
	return nil
}

type AuthorityOption interface {
	apply(*authorityConfig) error
}

type withExpirationImpl struct {
	o time.Duration
}

func (o withExpirationImpl) apply(c *authorityConfig) error {
	c.expiration = o.o
	return nil
}

func (o withExpirationImpl) String() string {
	name := "WithExpiration"

	// hack to avoid go vet error about passing a function to Sprintf
	var value interface{} = o.o
	return fmt.Sprintf("%s: %+v", name, value)
}

func WithExpiration(o time.Duration) AuthorityOption {
	return withExpirationImpl{
		o: o,
	}
}

type withNonceLengthImpl struct {
	o int
}

func (o withNonceLengthImpl) apply(c *authorityConfig) error {
	c.nonceLength = o.o
	return nil
}

func (o withNonceLengthImpl) String() string {
	name := "WithNonceLength"

	// hack to avoid go vet error about passing a function to Sprintf
	var value interface{} = o.o
	return fmt.Sprintf("%s: %+v", name, value)
}

func WithNonceLength(o int) AuthorityOption {
	return withNonceLengthImpl{
		o: o,
	}
}
