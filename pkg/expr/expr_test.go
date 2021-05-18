package expr

import "testing"

func TestEvalBool(t *testing.T) {
	type args struct {
		expression string
		env        map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "expect true",
			args: args{
				expression: "0 == 0",
				env:        nil,
			},
			want:    true,
			wantErr: false,
		}, {
			name: "expect false",
			args: args{
				expression: "0 != 0",
				env:        nil,
			},
			want:    false,
			wantErr: false,
		}, {
			name: "exitCode is 0",
			args: args{
				expression: "exitCode == 0",
				env: map[string]interface{}{
					"exitCode": 0,
				},
			},
			want:    true,
			wantErr: false,
		}, {
			name: "stdout is empty",
			args: args{
				expression: "len(stdout) == 0 && stdout == \"\"",
				env: map[string]interface{}{
					"stdout": "",
				},
			},
			want:    true,
			wantErr: false,
		}, {
			name: "embedded value assertion",
			args: args{
				expression: "obj.name == \"foo\"",
				env: map[string]interface{}{
					"obj": map[string]interface{}{
						"name": "foo",
					},
				},
			},
			want:    true,
			wantErr: false,
		}, {
			name: "not a bool expression",
			args: args{
				expression: "0",
				env:        nil,
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvalBool(tt.args.expression, tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("EvalBool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("EvalBool() got = %v, want %v", got, tt.want)
			}
		})
	}
}
