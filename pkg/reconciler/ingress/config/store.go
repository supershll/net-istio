/*
Copyright 2018 The Knative Authors.

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

package config

import (
	"context"
	"knative.dev/networking/pkg/apis/networking/v1alpha1"
	"strings"

	network "knative.dev/networking/pkg"
	"knative.dev/pkg/configmap"
)

type cfgKey struct{}

// Config of Istio.
// +k8s:deepcopy-gen=false
type Config struct {
	Istio   *Istio
	Network *network.Config
}

// FromContext fetch config from context.
func FromContext(ctx context.Context) *Config {
	return ctx.Value(cfgKey{}).(*Config)
}

// ludqfix
func GenConfigIstioWithIngress(ctx context.Context, ing *v1alpha1.Ingress) *Istio {
	ci := FromContext(ctx).Istio
	if ing != nil {
		userDefinedIngressGateway := strings.TrimSpace(ing.Annotations["networking.knative.dev/gateway.ingress"])
		if userDefinedIngressGateway != "" {
			ci.IngressGateways = make([]Gateway, 1)
			userDefinedIngressGatewayUrl := strings.TrimSpace(ing.Annotations["networking.knative.dev/gateway.ingress.url"])
			userDefinedExternalGatewaySplitArr := strings.Split(userDefinedIngressGateway, "/")
			gateway := Gateway{
				Name:       userDefinedExternalGatewaySplitArr[1],
				Namespace:  userDefinedExternalGatewaySplitArr[0],
				ServiceURL: userDefinedIngressGatewayUrl,
			}
			ci.IngressGateways[0] = gateway
		}

		userDefinedLocalGateway := strings.TrimSpace(ing.Annotations["networking.knative.dev/gateway.local"])
		if userDefinedLocalGateway != "" {
			ci.LocalGateways = make([]Gateway, 1)
			userDefinedLocalGatewayUrl := strings.TrimSpace(ing.Annotations["networking.knative.dev/gateway.local.url"])
			userDefinedLocalGatewaySplitArr := strings.Split(userDefinedLocalGateway, "/")
			gateway := Gateway{
				Name:       userDefinedLocalGatewaySplitArr[1],
				Namespace:  userDefinedLocalGatewaySplitArr[0],
				ServiceURL: userDefinedLocalGatewayUrl,
			}
			ci.LocalGateways[0] = gateway
		}
	}

	return ci
}

// ToContext adds config to given context.
func ToContext(ctx context.Context, c *Config) context.Context {
	return context.WithValue(ctx, cfgKey{}, c)
}

// Store is configmap.UntypedStore based config store.
// +k8s:deepcopy-gen=false
type Store struct {
	*configmap.UntypedStore
}

// NewStore creates a configmap.UntypedStore based config store.
//
// logger must be non-nil implementation of configmap.Logger (commonly used
// loggers conform)
//
// onAfterStore is a variadic list of callbacks to run
// after the ConfigMap has been processed and stored.
//
// See also: configmap.NewUntypedStore().
func NewStore(logger configmap.Logger, onAfterStore ...func(name string, value interface{})) *Store {
	store := &Store{
		UntypedStore: configmap.NewUntypedStore(
			"ingress",
			logger,
			configmap.Constructors{
				IstioConfigName:    NewIstioFromConfigMap,
				network.ConfigName: network.NewConfigFromConfigMap,
			},
			onAfterStore...,
		),
	}

	return store
}

// ToContext adds Store contents to given context.
func (s *Store) ToContext(ctx context.Context) context.Context {
	return ToContext(ctx, s.Load())
}

// Load fetches config from Store.
func (s *Store) Load() *Config {
	return &Config{
		Istio:   s.UntypedLoad(IstioConfigName).(*Istio).DeepCopy(),
		Network: s.UntypedLoad(network.ConfigName).(*network.Config).DeepCopy(),
	}
}
