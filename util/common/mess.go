package common

import corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"

func DataSourceGenerator(inlineDataSource bool) func(s string) *corev3.DataSource {
	return func(s string) *corev3.DataSource {
		if !inlineDataSource {
			return &corev3.DataSource{
				Specifier: &corev3.DataSource_Filename{
					Filename: s,
				},
			}
		}
		return &corev3.DataSource{
			Specifier: &corev3.DataSource_InlineString{
				InlineString: s,
			},
		}
	}
}
