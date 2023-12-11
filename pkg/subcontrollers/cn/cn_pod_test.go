package cn

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCnController_addWarehouseEnv1(t *testing.T) {
	type ServerFunc func(rw http.ResponseWriter, req *http.Request)

	type args struct {
		ctx        context.Context
		ServerFunc ServerFunc
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test FE support multi-warehouse",
			args: args{
				ctx: context.Background(),
				ServerFunc: func(rw http.ResponseWriter, req *http.Request) {
					rw.Write([]byte(`{"features": [{"name": "multi-warehouse"}], "version": "", "status": "OK"}`))
				},
			},
			want: true,
		},
		{
			name: "test FE does not support multi-warehouse",
			args: args{
				ctx: context.Background(),
				ServerFunc: func(rw http.ResponseWriter, req *http.Request) {
					rw.Write([]byte(`{"features": [], "version": "", "status": "OK"}`))
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create a mock server
			cc := &CnController{}
			server := httptest.NewServer(http.HandlerFunc(tt.args.ServerFunc))
			defer server.Close()

			assert.Equalf(t, tt.want, cc.addWarehouseEnv(tt.args.ctx, server.URL), "addWarehouseEnv(%v, %v)", tt.args.ctx, server.URL)
		})
	}
}
