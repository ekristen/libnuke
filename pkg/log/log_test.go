package log

import "testing"

func TestSorted(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input map[string]string
		want  string
	}{
		{
			name:  "empty",
			input: map[string]string{},
			want:  "[]",
		},
		{
			name: "one",
			input: map[string]string{
				"one": "1",
			},
			want: `[one: "1"]`,
		},
		{
			name: "two",
			input: map[string]string{
				"one": "1",
				"two": "2",
			},
			want: `[one: "1", two: "2"]`,
		},
		{
			name: "out-of-order",
			input: map[string]string{
				"two": "2",
				"one": "1",
			},
			want: `[one: "1", two: "2"]`,
		},
		{
			name: "underscore",
			input: map[string]string{
				"_one": "1",
			},
			want: "[]",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Sorted(tc.input)
			if got != tc.want {
				t.Errorf("sorted(%v) = %v; want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestLog(t *testing.T) {}
