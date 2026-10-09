package paginate

import (
	"errors"
	"net/url"
	"testing"

	"simiente-santa/backend/internal/platform/apperr"
)

func TestFromValuesDefaults(t *testing.T) {
	tests := []struct {
		name   string
		values url.Values
		want   Params
	}{
		{name: "sin parámetros", values: url.Values{}, want: Params{Limit: DefaultLimit, Offset: 0}},
		{
			name:   "parámetros vacíos",
			values: url.Values{"limit": {""}, "offset": {""}},
			want:   Params{Limit: DefaultLimit, Offset: 0},
		},
		{
			name:   "espacios sobrantes",
			values: url.Values{"limit": {" 25 "}, "offset": {" 5 "}},
			want:   Params{Limit: 25, Offset: 5},
		},
		{
			name:   "valores explícitos",
			values: url.Values{"limit": {"1"}, "offset": {"0"}},
			want:   Params{Limit: 1, Offset: 0},
		},
		{
			name:   "offset grande",
			values: url.Values{"limit": {"100"}, "offset": {"12345"}},
			want:   Params{Limit: 100, Offset: 12345},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromValues(tt.values)
			if err != nil {
				t.Fatalf("FromValues() error inesperado: %v", err)
			}
			if got != tt.want {
				t.Errorf("FromValues() = %+v, se esperaba %+v", got, tt.want)
			}
		})
	}
}

func TestFromValuesCapsLimit(t *testing.T) {
	tests := []struct {
		name  string
		limit string
	}{
		{name: "por encima del tope", limit: "1000"},
		{name: "muy por encima", limit: "999999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromValues(url.Values{"limit": {tt.limit}})
			if err != nil {
				t.Fatalf("FromValues() error inesperado: %v", err)
			}
			if got.Limit != MaxLimit {
				t.Errorf("limit = %d, se esperaba el tope %d", got.Limit, MaxLimit)
			}
			if got.Offset != 0 {
				t.Errorf("offset = %d, se esperaba 0", got.Offset)
			}
		})
	}
}

func TestFromValuesInvalid(t *testing.T) {
	tests := []struct {
		name      string
		values    url.Values
		wantField string
	}{
		{name: "limit cero", values: url.Values{"limit": {"0"}}, wantField: "limit"},
		{name: "limit negativo", values: url.Values{"limit": {"-1"}}, wantField: "limit"},
		{name: "limit no numérico", values: url.Values{"limit": {"abc"}}, wantField: "limit"},
		{name: "limit decimal", values: url.Values{"limit": {"2.5"}}, wantField: "limit"},
		{name: "offset negativo", values: url.Values{"offset": {"-1"}}, wantField: "offset"},
		{name: "offset no numérico", values: url.Values{"offset": {"x"}}, wantField: "offset"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FromValues(tt.values)
			if err == nil {
				t.Fatal("FromValues() = nil, se esperaba apperr.Invalid")
			}

			var domainErr *apperr.Error
			if !errors.As(err, &domainErr) {
				t.Fatalf("error = %T, se esperaba *apperr.Error", err)
			}
			if domainErr.Kind != apperr.KindInvalid {
				t.Errorf("Kind = %v, se esperaba KindInvalid", domainErr.Kind)
			}
			if domainErr.Code() != "invalid" {
				t.Errorf("Code() = %q, se esperaba invalid", domainErr.Code())
			}
			if domainErr.HTTPStatus() != 400 {
				t.Errorf("HTTPStatus() = %d, se esperaba 400", domainErr.HTTPStatus())
			}
			if _, ok := domainErr.Details[tt.wantField]; !ok {
				t.Errorf("Details = %v, falta el campo %q", domainErr.Details, tt.wantField)
			}
		})
	}
}
