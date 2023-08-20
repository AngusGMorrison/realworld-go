package option

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func Test_Some(t *testing.T) {
	t.Parallel()

	value := "value"
	wantOpt := Option[string]{some: true, value: value}

	gotOpt := Some(value)
	assert.Equal(t, wantOpt, gotOpt)
}

func Test_None(t *testing.T) {
	t.Parallel()

	wantOpt := Option[string]{}

	gotOpt := None[string]()
	assert.Equal(t, wantOpt, gotOpt)
}

func Test_Option_IsSome(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		opt  Option[string]
		want bool
	}{
		{
			name: "Some",
			opt:  Some("value"),
			want: true,
		},
		{
			name: "None",
			opt:  None[string](),
			want: false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.opt.IsSome()
			assert.Equal(t, tc.want, got)
		})
	}
}

func Test_Option_Unwrap(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		opt     Option[string]
		wantVal string
		wantErr error
	}{
		{
			name:    "Some",
			opt:     Some("value"),
			wantVal: "value",
			wantErr: nil,
		},
		{
			name:    "None",
			opt:     None[string](),
			wantVal: "",
			wantErr: ErrEmptyOption,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotVal, gotErr := tc.opt.Unwrap()
			assert.ErrorIs(t, tc.wantErr, gotErr)
			assert.Equal(t, tc.wantVal, gotVal)
		})
	}
}

func Test_Option_UnwrapOrZero(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		opt     Option[string]
		wantVal string
	}{
		{
			name:    "Some",
			opt:     Some("value"),
			wantVal: "value",
		},
		{
			name:    "None",
			opt:     None[string](),
			wantVal: "",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotVal := tc.opt.UnwrapOrZero()
			assert.Equal(t, tc.wantVal, gotVal)
		})
	}
}

func Test_Option_Map(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		opt     Option[string]
		wantOpt Option[int]
		wantErr error
	}{
		{
			name:    "Some, conversion succeeds",
			opt:     Some("1"),
			wantOpt: Some(1),
			wantErr: nil,
		},
		{
			name:    "Some, conversion fails",
			opt:     Some("value"),
			wantOpt: None[int](),
			wantErr: strconv.ErrSyntax,
		},
		{
			name:    "None",
			opt:     None[string](),
			wantOpt: None[int](),
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotOpt, err := Map(tc.opt, strconv.Atoi)
			assert.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.wantOpt, gotOpt)
		})
	}
}

func Test_Option_GoString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		opt  Option[string]
	}{
		{
			name: "IsSome",
			opt:  Some("value"),
		},
		{
			name: "None",
			opt:  None[string](),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			want := fmt.Sprintf("Option[%[1]T]{some: %[2]t, value: %#[1]v}", tc.opt.value, tc.opt.some)

			got := tc.opt.GoString()
			assert.Equal(t, want, got)
		})
	}
}

func Test_Option_String(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		opt  Option[string]
		want string
	}{
		{
			name: "IsSome",
			opt:  Some("value"),
			want: "IsSome(value)",
		},
		{
			name: "None",
			opt:  None[string](),
			want: "None",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.opt.String()
			assert.Equal(t, tc.want, got)
		})
	}
}

func Test_Option_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		bytes   []byte
		wantOpt Option[string]
	}{
		{
			name:    "IsSome",
			bytes:   []byte(`"value"`),
			wantOpt: Some("value"),
		},
		{
			name:    "None",
			bytes:   nil,
			wantOpt: None[string](),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var gotOpt Option[string]
			err := gotOpt.UnmarshalJSON(tc.bytes)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantOpt, gotOpt)
		})
	}
}
