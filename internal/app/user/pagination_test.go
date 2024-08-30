package user

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_decodeCursor(t *testing.T) {
	type args struct {
		encodedCursor string
	}
	tests := []struct {
		name       string
		args       args
		wantTs     string
		wantUserID string
		wantErr    error
	}{
		{
			name: "success",
			args: args{
				encodedCursor: "MjAyNC0wOC0yMlQyMDowOToxMS45MzgyMiswMTowMHxjMTJlMjNmMy1mNWUzLTQxYmMtYWVjYS05ZDY2YmQwYjk2YTM=",
			},
			wantTs:     "2024-08-22 20:09:11.93822 +0100 WEST",
			wantUserID: "c12e23f3-f5e3-41bc-aeca-9d66bd0b96a3",
			wantErr:    nil,
		},
		{
			name: "invalid cursor",
			args: args{
				encodedCursor: "MjAyNC0wDowOToxMS45MzgyMiswMTowMHxjMTJlMjNmMy1mNWUzLTQxYmMtYWVjYS05ZDY2YmQwYjk2YTM=",
			},
			wantErr: fmt.Errorf("illegal base64 data at input byte 83"),
		},
		{
			name: "invalid cursor content",
			args: args{
				encodedCursor: base64.StdEncoding.EncodeToString([]byte("a|b|c")),
			},
			wantErr: fmt.Errorf("cursor is invalid"),
		},
		{
			name: "invalid timestamp",
			args: args{
				encodedCursor: base64.StdEncoding.EncodeToString([]byte("a|b")),
			},
			wantErr: fmt.Errorf("cursor is invalid: timestamp"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTs, gotUserID, err := decodeCursor(tt.args.encodedCursor)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}

			if gotTs.String() != tt.wantTs {
				t.Errorf("decodeCursor() gotTs = %v, want %v", gotTs, tt.wantTs)
			}
			if gotUserID != tt.wantUserID {
				t.Errorf("decodeCursor() gotUserID = %v, want %v", gotUserID, tt.wantUserID)
			}
		})
	}
}

func Test_encodeCursor(t *testing.T) {
	type args struct {
		ts     time.Time
		userID string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "success",
			args: args{
				ts:     time.Date(2024, 8, 22, 20, 30, 3, 0, time.UTC),
				userID: "abc",
			},
			want: "MjAyNC0wOC0yMlQyMDozMDowM1p8YWJj",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := encodeCursor(tt.args.ts, tt.args.userID); got != tt.want {
				t.Errorf("encodeCursor() = %v, want %v", got, tt.want)
			}
		})
	}
}
