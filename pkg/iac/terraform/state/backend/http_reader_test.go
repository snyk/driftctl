package backend

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewHTTPReader(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		wantURL string
		wantErr bool
	}{
		{
			name: "Should fail on wrong URL",
			args: args{
				url: "wrong_url",
			},
			wantURL: "",
			wantErr: true,
		},
		{
			name: "Should fetch URL",
			args: args{
				url: "https://raw.githubusercontent.com/cloudskiff/driftctl/main/.dockerignore",
			},
			wantURL: "https://raw.githubusercontent.com/cloudskiff/driftctl/main/.dockerignore",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewHTTPReader(tt.args.url)
			if tt.wantErr {
				assert.Error(t, err)
				return
			} else {
				assert.NoError(t, err)
			}
			assert.NotNil(t, got)
			assert.Equal(t, tt.wantURL, got.url)
		})
	}
}

func TestHTTPBackend_Close(t *testing.T) {
	type fields struct {
		url    string
		reader func() io.ReadCloser
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "should fail to close reader",
			fields: fields{
				url: "",
				reader: func() io.ReadCloser {
					return nil
				},
			},
			wantErr: true,
		},
		{
			name: "should close reader",
			fields: fields{
				url: "",
				reader: func() io.ReadCloser {
					m := &MockReaderMock{}
					m.On("Close").Return(nil)
					return m
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HTTPBackend{
				url:    tt.fields.url,
				reader: tt.fields.reader(),
			}
			if err := h.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHTTPBackend_Read(t *testing.T) {
	type fields struct {
		url    string
		reader func() io.ReadCloser
	}
	type args struct {
		p []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantN   int
		wantErr error
	}{
		{
			name: "should fail to read because of nil reader",
			fields: fields{
				url: "",
				reader: func() io.ReadCloser {
					return nil
				},
			},
			wantErr: errors.New("Reader not initialized"),
		},
		{
			name: "should fail to read",
			fields: fields{
				url: "",
				reader: func() io.ReadCloser {
					m := &MockReaderMock{}
					m.On("Read", mock.Anything).Return(0, errors.New("test"))
					return m
				},
			},
			wantErr: errors.New("test"),
		},
		{
			name: "should read",
			fields: fields{
				url: "",
				reader: func() io.ReadCloser {
					m := &MockReaderMock{}
					m.On("Read", mock.Anything).Return(0, nil)
					return m
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HTTPBackend{
				url:    tt.fields.url,
				reader: tt.fields.reader(),
			}
			gotN, err := h.Read(tt.args.p)

			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			if gotN != tt.wantN {
				t.Errorf("Read() gotN = %v, want %v", gotN, tt.wantN)
			}
		})
	}
}
