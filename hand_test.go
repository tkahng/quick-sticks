package sticks

import (
	"testing"
)

func TestHand_Attack(t *testing.T) {
	type fields struct {
		Fingers int
	}
	type args struct {
		opp *Hand
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "attack hand with 0 fingers",
			fields: fields{
				Fingers: 5,
			},
			args: args{
				opp: &Hand{
					Fingers: 0,
				},
			},
			wantErr: true,
		},
		{
			name: "attack hand with 1 fingers",
			fields: fields{
				Fingers: 5,
			},
			args: args{
				opp: &Hand{
					Fingers: 1,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Hand{
				Fingers: tt.fields.Fingers,
			}
			if err := h.Attack(tt.args.opp); (err != nil) != tt.wantErr {
				t.Errorf("Hand.Attack() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHand_Split(t *testing.T) {
	type fields struct {
		Fingers int
	}
	type args struct {
		other *Hand
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "split hand with 1 fingers",
			fields: fields{
				Fingers: 5,
			},
			args: args{
				other: &Hand{
					Fingers: 1,
				},
			},
			wantErr: true,
		},
		{
			name: "split hand with 2 fingers",
			fields: fields{
				Fingers: 5,
			},
			args: args{
				other: &Hand{
					Fingers: 2,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Hand{
				Fingers: tt.fields.Fingers,
			}
			if err := h.Split(tt.args.other); (err != nil) != tt.wantErr {
				t.Errorf("Hand.Split() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
